package apimanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/file"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/radovskyb/watcher"
	"github.com/spf13/pflag"
)

type Config struct {
	HttpFiles        []string `yaml:"httpFiles"` // http mock file.
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	Watcher          bool     `yaml:"watcher"` // whether to enable file listening
}

func NewConfig() *Config {
	return &Config{
		HttpFiles:        []string{},
		Methods:          []string{},
		Headers:          []string{},
		Origins:          []string{},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		Watcher:          false,
	}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	MockCaseCenter

	Start() error
	Close() error

	MockHttpResponse(ctx context.Context, request *http.Request) (*interact.HttpImposter, error)
	MockGrpcResponse(ctx context.Context, request *interact.GRpcRequest) (*interact.GRpcImposter, error)
}

type Manager struct {
	MockCaseCenter
	httpApis map[string]*interact.HttpImposter
	cfg      *Config
	router   *mux.Router
	log      logger.Logger
}

func New(log logger.Logger, cfg *Config) Provider {
	m := &Manager{
		MockCaseCenter: NewMockCaseCenter(),
		router:         mux.NewRouter(),
		httpApis:       make(map[string]*interact.HttpImposter),
		log:            log.NewLogger("api"),
		cfg:            cfg,
	}
	m.loadImposter()
	m.watcher()

	return m
}

func (m *Manager) Start() error {
	handlers.CORS(PrepareAccessControl(m.cfg)...)(m.router)
	m.addHttpImposterHandler()
	m.printRouter()
	return nil
}

func (m *Manager) Close() error {
	//TODO: Avoid full load.
	m.httpApis = make(map[string]*interact.HttpImposter)
	m.router = mux.NewRouter()
	return nil
}

func (m *Manager) MockHttpResponse(ctx context.Context, request *http.Request) (*interact.HttpImposter, error) {
	// TODO: add a lock, Cannot use when loading mock !!!!
	var match mux.RouteMatch
	if m.router.Match(request, &match) {
		return m.httpApis[match.Route.GetName()], nil
	}
	return nil, fmt.Errorf("cannot mock http request: %v", request)
}

func (m *Manager) MockGrpcResponse(ctx context.Context, request *interact.GRpcRequest) (*interact.GRpcImposter, error) {
	//TODO implement me
	panic("implement me")
}

// Register proxy request to Router.
// It will match: "path", "host", "method", "params".
func (m *Manager) addHttpImposterHandler() {
	for _, imposter := range m.GetAllHttp() {
		u, err := url.Parse(imposter.Request.Url)
		if err != nil {
			m.log.Error(nil, err.Error())
			continue
		}

		r := m.router.
			Path(u.Path).
			Methods(imposter.Request.Method).
			Host(u.Host)

		if imposter.Request.Headers != nil {
			for k, v := range imposter.Request.Headers {
				r.HeadersRegexp(k, v)
			}
		}

		if imposter.Request.Params != nil {
			for k, v := range imposter.Request.Params {
				r.Queries(k, v)
			}
		}

		r.Name(imposter.Id)
		m.httpApis[imposter.Id] = &imposter
	}
}

// print all router.
func (m *Manager) printRouter() {
	m.log.Debug(nil, "print all http router:")
	_ = m.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		met, err1 := route.GetMethods()
		tpl, err2 := route.GetPathTemplate()
		host, err3 := route.GetHostTemplate()
		queries, err4 := route.GetQueriesTemplates()
		m.log.Debug(nil, `
			--------------------
			Method: %v, err1: %v
			path: %s, err2: %v
			Host: %v, err3: %v
			queries: %v, err4: %v
			--------------------
		`, met, err1, tpl, err2, host, err3, queries, err4)
		return nil
	})
}

func (m *Manager) loadImposter() {
	for _, filePath := range m.cfg.HttpFiles {
		m.loadSingleImposter(filePath)
	}
}

//Initialize and start the file watcher if the watcher option is true
func (m *Manager) watcher() {
	if !m.cfg.Watcher {
		return
	}

	w, err := file.InitializeWatcher(m.cfg.HttpFiles...)
	if err != nil {
		m.log.Fatal(nil, "error:", err)
	}

	file.AttachWatcher(w, func(evn watcher.Event) {
		m.loadSingleImposter(evn.Path)
		if err = m.Close(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}

		if err = m.Start(); err != nil {
			m.log.Fatal(nil, "error:", err)
		}
	})
}

// loading single http file to imposter
func (m *Manager) loadSingleImposter(filePath string) {
	m.UnLoadAllGlobalHttp()
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

	m.AddGlobalHttp(imposter...)
}
