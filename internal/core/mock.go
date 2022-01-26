package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"alsritter.icu/middlebaby/internal/proxy"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/radovskyb/watcher"
	"golang.org/x/sync/errgroup"
)

type MockServeBuilder struct {
	mock *MockServe
}

// return a MockServe Builder.
func NewMockServeBuilder(config *config.Config, group *errgroup.Group, done chan bool) *MockServeBuilder {
	mock := &MockServe{
		config:    config,
		group:     group,
		done:      done,
		imposters: make(map[string][]proxy.Imposter),
	}

	return &MockServeBuilder{mock: mock}
}

func (s *MockServeBuilder) LoadImposter() *MockServeBuilder {
	for _, filePath := range s.mock.config.HttpFiles {
		loadImposter(filePath, s.mock.imposters)
	}

	return s
}

//Initialize and start the file watcher if the watcher option is true
func (s *MockServeBuilder) Watcher() *MockServeBuilder {
	if !s.mock.config.Watcher {
		return nil
	}

	w, err := config.InitializeWatcher(s.mock.config.HttpFiles...)
	if err != nil {
		log.Fatal(err)
	}

	config.AttachWatcher(w, func(evn watcher.Event) {
		loadImposter(evn.Path, s.mock.imposters)

		if err := s.mock.server.Shutdown(); err != nil {
			log.Fatal(err)
		}

		s.Serve()
		s.mock.server.Run()
	})

	s.mock.w = w
	return s
}

// Create a Server for the mock object
func (s *MockServeBuilder) Serve() *MockServeBuilder {
	router := mux.NewRouter()
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.mock.config.Port),
		WriteTimeout: time.Second * 3,
		Handler:      handlers.CORS(proxy.PrepareAccessControl(s.mock.config.CORS)...)(router),
	}

	s.mock.server = proxy.NewServer(router, httpServer, mapToSlice(s.mock.imposters))
	return s
}

// Build return a MockServe.
func (s *MockServeBuilder) Build() *MockServe {
	s.LoadImposter().Watcher().Serve()
	return s.mock
}

type MockServe struct {
	config    *config.Config
	imposters map[string][]proxy.Imposter
	server    *proxy.Server
	w         *watcher.Watcher
	group     *errgroup.Group
	done      chan bool
}

// Start Mock Serve.
func (s *MockServe) Run() {
	s.group.Go(func() error {
		// make sure idle connections returned
		processed := make(chan struct{})
		go func() {
			switch {
			case <-s.done:
			}

			if err := s.server.Shutdown(); nil != err {
				log.Fatalf("server shutdown failed, err: %v\n", err)
			}
			close(processed)
		}()

		// serve
		s.server.Run()
		// waiting for goroutine above processed
		<-processed
		return nil
	})
}

// loading single http file to imposter
func loadImposter(filePath string, imposters map[string][]proxy.Imposter) {
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

	var imposter []proxy.Imposter
	if err := json.Unmarshal(bytes, &imposter); err != nil {
		log.Errorf("%w: error while unmarshal configFile file %s", err, filePath)
	}

	imposters[filePath] = imposter
}

func mapToSlice(m map[string][]proxy.Imposter) []proxy.Imposter {
	s := make([]proxy.Imposter, 0, len(m))
	for _, v := range m {
		s = append(s, v...)
	}
	return s
}
