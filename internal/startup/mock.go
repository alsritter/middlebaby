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

	"alsritter.icu/middlebaby/internal/common"
	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	proxy_http "alsritter.icu/middlebaby/internal/proxy/http"
	"alsritter.icu/middlebaby/internal/startup/plugin"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type MockServe struct {
	env       plugin.Env
	server    *http.Server
	imposters map[string][]common.HttpImposter
}

// return a MockServe Builder.
func NewMockServe(env plugin.Env) *MockServe {
	mock := &MockServe{
		env:       env,
		imposters: make(map[string][]common.HttpImposter),
		server:    &http.Server{},
	}

	if err := http2.ConfigureServer(mock.server, &http2.Server{}); err != nil {
		log.Fatal("proxy http2 error: ", err)
	}

	mock.loadImposter()
	mock.watcher()

	return mock
}

func (m *MockServe) loadImposter() {
	for _, filePath := range m.env.GetConfig().HttpFiles {
		loadImposter(filePath, m.imposters)
	}

}

//Initialize and start the file watcher if the watcher option is true
func (m *MockServe) watcher() {
	if !m.env.GetConfig().Watcher {
		return
	}

	w, err := config.InitializeWatcher(m.env.GetConfig().HttpFiles...)
	if err != nil {
		log.Fatal(err)
	}

	config.AttachWatcher(w, func(evn watcher.Event) {
		loadImposter(evn.Path, m.imposters)

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
	m.server.Handler = h2c.NewHandler(m.setupProxy(), &http2.Server{})

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
	h.AddProxy(proxy_http.NewHttpImposterHandler(mapToSlice(m.imposters), m.env.GetConfig().CORS))
	return h
}

// loading single http file to imposter
func loadImposter(filePath string, imposters map[string][]common.HttpImposter) {
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

	imposters[filePath] = imposter
}

func mapToSlice(m map[string][]common.HttpImposter) []common.HttpImposter {
	s := make([]common.HttpImposter, 0, len(m))
	for _, v := range m {
		s = append(s, v...)
	}
	return s
}
