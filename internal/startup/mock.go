package startup

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alsritter/middlebaby/internal/file/common"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
	"github.com/alsritter/middlebaby/pkg/proxy"
	proxy_http "github.com/alsritter/middlebaby/pkg/proxy/http"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type MockServe struct {
	env        plugin.Env
	server     *http.Server
	mockCenter proxy.MockCenter
	log        logger.Logger
}

// return a MockServe Builder.
func NewMockServe(env plugin.Env, mockCenter proxy.MockCenter, log logger.Logger) *MockServe {
	mock := &MockServe{
		env:        env,
		mockCenter: mockCenter,
		server:     &http.Server{},
		log:        log,
	}

	mock.loadImposter()
	mock.watcher()

	return mock
}

func (m *MockServe) loadImposter() {
	for _, filePath := range m.env.GetConfig().HttpFiles {
		m.loadSingleImposter(filePath, m.mockCenter)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (m *MockServe) watcher() {
	if !m.env.GetConfig().Watcher {
		return
	}

	w, err := common.InitializeWatcher(m.env.GetConfig().HttpFiles...)
	if err != nil {
		m.log.Fatal(nil, "error:", err)
	}

	common.AttachWatcher(w, func(evn watcher.Event) {
		m.loadSingleImposter(evn.Path, m.mockCenter)
		if err := m.Close(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}

		m.Run()
	})
}

func (m *MockServe) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", m.env.GetConfig().Port))
	if err != nil {
		return fmt.Errorf("failed to listen the port: %d, err: %v", m.env.GetConfig().Port, err)
	}

	// call ServeHTTP function handle request.
	// support HTTP2.0 with h2c package.
	m.server.Handler = h2c.NewHandler(m.setupProxy(), &http2.Server{})
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
func (m *MockServe) Close() error {
	m.log.Info(nil, "stopping server...")
	if err := m.server.Shutdown(context.TODO()); err != nil {
		return fmt.Errorf("server Shutdown failed:%+v", err)
	}

	return nil
}

func (m *MockServe) setupProxy() http.Handler {
	h := proxy.NewMockList(m.env.GetConfig().EnableDirect)
	h.AddProxy(proxy_http.NewHttpImposterHandler(m.mockCenter, m.env.GetConfig().CORS))
	h.AddDirect(proxy_http.NewHttpDirectHandler())
	return h
}

// loading single http file to imposter
func (m *MockServe) loadSingleImposter(filePath string, mockCenter proxy.MockCenter) {
	mockCenter.UnLoadAllGlobalHttp()

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

	var imposter []common.HttpImposter
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		m.log.Error(nil, "%w: error while unmarshal configFile file %s", err, filePath)
	}

	mockCenter.AddGlobalHttp(imposter...)
}
