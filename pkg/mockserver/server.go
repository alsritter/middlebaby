package mockserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	EnableDirect bool `yaml:"enableDirect"` // whether the missed mock allows real requests
	MockPort     int  `yaml:"mockPort"`     // proxy port
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.MockPort == 0 {
		return errors.New("[mockserver] mock server listener port cannot be empty")
	}
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.IntVar(&c.MockPort, prefix+"mockserver.port", c.MockPort, "mock server listener port")
}

// Provider defines the mock server interface
type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc) error
}

type MockServe struct {
	cfg        *Config
	server     *http.Server
	mit        *mitmproxy
	apiManager apimanager.Provider
	log        logger.Logger
	w          *watcher.Watcher
}

func New(log logger.Logger, cfg *Config, apiManager apimanager.Provider) Provider {
	mock := &MockServe{
		mit:        NewProxy(cfg.EnableDirect, log),
		cfg:        cfg,
		apiManager: apiManager,
		server:     &http.Server{},
		log:        log.NewLogger("mock"),
	}
	return mock
}

func (m *MockServe) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	util.StartServiceAsync(ctx, m.log, cancelFunc, func() error {
		return m.start()
	}, func() error {
		if m.w != nil {
			m.w.Close()
		}
		return m.close()
	})

	return nil
}

func (m *MockServe) start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", m.cfg.MockPort))
	if err != nil {
		return fmt.Errorf("failed to listen the port: %d, err: %v", m.cfg.MockPort, err)
	}

	// call ServeHTTP function handle request.
	// support HTTP2.0 with h2c package.
	m.server.Handler = h2c.NewHandler(m.mit.px, &http2.Server{})
	if err != nil {
		return err
	}

	if err := http2.ConfigureServer(m.server, &http2.Server{}); err != nil {
		return fmt.Errorf("proxy http2 error: %v", err)
	}

	if err := m.server.Serve(l); err != nil {
		if err.Error() != "http: Server closed" {
			return fmt.Errorf("failed to start the proxy server: %v", err)
		}
	}

	return nil
}

// Close shutdown the current http server
func (m *MockServe) close() error {
	m.log.Info(nil, "stopping server...")
	if err := m.server.Shutdown(context.TODO()); err != nil {
		return fmt.Errorf("server Shutdown failed:%+v", err)
	}
	return nil
}
