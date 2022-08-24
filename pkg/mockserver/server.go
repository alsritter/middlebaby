package mockserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/file"
	"github.com/spf13/pflag"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	HttpFiles    []string `yaml:"httpFiles"`    // http mock file.
	EnableDirect bool     `yaml:"enableDirect"` // whether the missed mock allows real requests
	MockPort     int      `yaml:"mockPort"`     // proxy port
	Watcher      bool     `yaml:"watcher"`      // whether to enable file listening
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
}

func New(log logger.Logger, cfg *Config, apiManager apimanager.Provider) Provider {
	mock := &MockServe{
		mit:        NewProxy(cfg.EnableDirect, log),
		cfg:        cfg,
		apiManager: apiManager,
		server:     &http.Server{},
		log:        log.NewLogger("MockServer"),
	}
	mock.loadImposter()
	mock.watcher()
	return mock
}

func (m *MockServe) loadImposter() {
	for _, filePath := range m.cfg.HttpFiles {
		m.loadSingleImposter(filePath)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (m *MockServe) watcher() {
	if !m.cfg.Watcher {
		return
	}

	w, err := file.InitializeWatcher(m.cfg.HttpFiles...)
	if err != nil {
		m.log.Fatal(nil, "error:", err)
	}

	file.AttachWatcher(w, func(evn watcher.Event) {
		m.loadSingleImposter(evn.Path)
		if err := m.close(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}

		if err = m.apiManager.Close(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}

		if err = m.apiManager.Start(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}

		if err = m.start(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}
	})
}

func (m *MockServe) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	util.StartServiceAsync(ctx, m.log, cancelFunc, func() error {
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
	m.server.Handler = h2c.NewHandler(m.mit.px, &http2.Server{})
	if err != nil {
		return err
	}

	if err := http2.ConfigureServer(m.server, &http2.Server{}); err != nil {
		return fmt.Errorf("proxy http2 error: ", err)
	}

	if err := m.server.Serve(l); err != nil {
		if err.Error() != "http: Server closed" {
			return fmt.Errorf("failed to start the proxy server: ", err)
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

// loading single http file to imposter
func (m *MockServe) loadSingleImposter(filePath string) {
	m.apiManager.UnLoadAllGlobalHttp()
	if !filepath.IsAbs(filePath) {
		if fp, err := filepath.Abs(filePath); err != nil {
			m.log.Error(nil, "to absolute representation path err: %s", err)
			return
		} else {
			filePath = fp
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		m.log.Error(nil, "%w: error trying to read config file: %s", err, filePath)
	}
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []interact.HttpImposter
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		m.log.Error(nil, "%w: error while unmarshal configFile file %s", err, filePath)
	}

	m.apiManager.AddGlobalHttp(imposter...)
}
