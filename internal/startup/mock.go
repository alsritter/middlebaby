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
	"github.com/alsritter/middlebaby/internal/log"
	"github.com/alsritter/middlebaby/internal/proxy"
	proxy_http "github.com/alsritter/middlebaby/internal/proxy/http"
	"github.com/alsritter/middlebaby/internal/startup/plugin"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type MockServe struct {
	env        plugin.Env
	server     *http.Server
	mockCenter proxy.MockCenter
}

// return a MockServe Builder.
func NewMockServe(env plugin.Env, mockCenter proxy.MockCenter) *MockServe {
	mock := &MockServe{
		env:        env,
		mockCenter: mockCenter,
		server:     &http.Server{},
	}

	mock.loadImposter()
	mock.watcher()

	return mock
}

func (m *MockServe) loadImposter() {
	for _, filePath := range m.env.GetConfig().HttpFiles {
		loadImposter(filePath, m.mockCenter)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (m *MockServe) watcher() {
	if !m.env.GetConfig().Watcher {
		return
	}

	w, err := common.InitializeWatcher(m.env.GetConfig().HttpFiles...)
	if err != nil {
		log.Fatal(err)
	}

	common.AttachWatcher(w, func(evn watcher.Event) {
		loadImposter(evn.Path, m.mockCenter)
		if err := m.Shutdown(); err != nil {
			log.Fatal(err)
		}

		m.Run()
	})
}

func (m *MockServe) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", m.env.GetConfig().Port))
	if err != nil {
		log.Errorf("failed to listen the port: %d, err: %v", m.env.GetConfig().Port, err)
		return err
	}

	// call ServeHTTP function handle request.
	// support HTTP2.0 with h2c package.
	m.server.Handler = h2c.NewHandler(m.setupProxy(), &http2.Server{})
	if err != nil {
		log.Fatal(err)
	}

	if err := http2.ConfigureServer(m.server, &http2.Server{}); err != nil {
		log.Fatal("proxy http2 error: ", err)
	}

	if err := m.server.Serve(l); err != nil {
		if err.Error() != "http: Server closed" {
			log.Error("failed to start the proxy server: ", err)
			return err
		}
	}

	return nil
}

// shutdown shutdown the current http server
func (m *MockServe) Shutdown() error {
	log.Info("stopping server...")
	if err := m.server.Shutdown(context.TODO()); err != nil {
		log.Fatalf("server Shutdown failed:%+v", err)
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
func loadImposter(filePath string, mockCenter proxy.MockCenter) {
	mockCenter.UnLoadAllGlobalHttp()

	if !filepath.IsAbs(filePath) {
		if fp, err := filepath.Abs(filePath); err != nil {
			log.Errorf("to absolute representation path err: %s", err)
			return
		} else {
			filePath = fp
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("%w: error trying to read config file: %s", err, filePath)
	}
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)

	var imposter []common.HttpImposter
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		log.Errorf("%w: error while unmarshal configFile file %s", err, filePath)
	}

	mockCenter.AddGlobalHttp(imposter...)
}
