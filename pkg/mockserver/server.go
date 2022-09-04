package mockserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/mockserver/grpchandler"
	"github.com/alsritter/middlebaby/pkg/mockserver/httphandler"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/spf13/pflag"

	"github.com/alsritter/middlebaby/pkg/util/logger"
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
	GetPort() int
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
}

type MockServe struct {
	cfg *Config
	logger.Logger
	server       *http.Server
	apiManager   apimanager.Provider
	httpServer   http.Handler
	grpcServer   http.Handler
	grpcProvider grpchandler.Provider
	httpProvider httphandler.Provider
}

func New(log logger.Logger, cfg *Config,
	apiManager apimanager.Provider, protoManager protomanager.Provider) Provider {
	l := log.NewLogger("mock")
	mock := &MockServe{
		cfg:          cfg,
		Logger:       l,
		server:       &http.Server{},
		apiManager:   apiManager,
		grpcProvider: grpchandler.New(l, apiManager, protoManager),
		httpProvider: httphandler.New(l, &httphandler.Config{EnableDirect: cfg.EnableDirect}, apiManager),
	}
	return mock
}

func (m *MockServe) GetPort() int {
	return m.cfg.MockPort
}

func (m *MockServe) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	if err := m.grpcProvider.Init(ctx, cancelFunc, wg); err != nil {
		return err
	}

	m.grpcServer = m.grpcProvider.GetServer()
	m.httpServer = m.httpProvider.GetServer()

	util.StartServiceAsync(ctx, m, cancelFunc, wg, func() error {
		return m.start()
	}, func() error {
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
	m.server.Handler = h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(
			r.Header.Get("Content-Type"), "application/grpc") {
			m.grpcServer.ServeHTTP(w, r)
		} else {
			m.httpServer.ServeHTTP(w, r)
		}
	}), &http2.Server{})

	if err := http2.ConfigureServer(m.server, &http2.Server{}); err != nil {
		return fmt.Errorf("proxy http2 error: %v", err)
	}

	m.Info(nil, "Mock server started, Listen port: %d", m.cfg.MockPort)
	if err := m.server.Serve(l); err != nil {
		if err.Error() != "http: Server closed" {
			return fmt.Errorf("failed to start the proxy server: %v", err)
		}
	}

	return nil
}

// Close shutdown the current http server
func (m *MockServe) close() error {
	m.Info(nil, "stopping server...")
	if err := m.server.Shutdown(context.TODO()); err != nil {
		return fmt.Errorf("server Shutdown failed: [%v]", err)
	}
	return nil
}
