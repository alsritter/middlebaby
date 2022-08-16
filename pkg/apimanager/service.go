package apimanager

import (
	"context"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
)

type Config struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`

	*ConfigCORS
}

// ConfigCORS representation of section CORS of the yaml
type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

func (c *Config) Validate() error {
	return nil
}

type Provider interface {
	ApiMockCenter

	Start() error
	Close() error

	MockHttpResponse(ctx context.Context, request *http.Request) (*interact.HttpImposter, error)
	MockGrpcResponse(ctx context.Context, request *interact.GRpcRequest) (*interact.GRpcImposter, error)
}

type Manager struct {
	ApiMockCenter
	httpApis map[string]*interact.HttpImposter
	cfg      *Config
	router   *mux.Router
	log      logger.Logger
}

func New(log logger.Logger, cfg *Config) Provider {
	return &Manager{
		ApiMockCenter: NewMockCenter(),
		router:        mux.NewRouter(),
		httpApis:      make(map[string]*interact.HttpImposter),
		log:           log,
		cfg:           cfg,
	}
}

func (m *Manager) Start() error {
	handlers.CORS(PrepareAccessControl(m.cfg.ConfigCORS)...)(m.router)
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
