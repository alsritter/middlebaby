package apimanager

import (
	"context"
	"fmt"
	"net/url"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
)

type Config struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

func NewConfig() *Config {
	return &Config{
		Methods:          []string{},
		Headers:          []string{},
		Origins:          []string{},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
	}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	Start() error
	Close() error

	MockResponse(ctx context.Context, request *interact.Request) (*interact.ImposterCase, error)
}

type Manager struct {
	caseApis map[string]*interact.ImposterCase
	cfg      *Config
	router   *mux.Router
	log      logger.Logger
}

func New(log logger.Logger, cfg *Config) Provider {
	return &Manager{
		router:   mux.NewRouter(),
		caseApis: make(map[string]*interact.ImposterCase),
		log:      log.NewLogger("api"),
		cfg:      cfg,
	}
}

func (m *Manager) Start() error {
	handlers.CORS(PrepareAccessControl(m.cfg)...)(m.router)
	m.addHttpImposterHandler()
	m.printRouter()
	return nil
}

func (m *Manager) Close() error {
	//TODO: Avoid full load.
	m.caseApis = make(map[string]*interact.ImposterCase)
	m.router = mux.NewRouter()
	return nil
}

func (m *Manager) MockResponse(ctx context.Context, request *interact.Request) (*interact.ImposterCase, error) {
	// TODO: add a lock, Cannot use when loading mock !!!!
	var match mux.RouteMatch
	if m.router.Match(request, &match) {
		return m.caseApis[match.Route.GetName()], nil
	}
	return nil, fmt.Errorf("cannot mock http request: %v", request)
}

// Register proxy request to Router.
// It will match: "path", "host", "method", "params".
func (m *Manager) addHttpImposterHandler() {
	for _, imposter := range m.GetAllCase() {
		u, err := url.Parse(imposter.Request.Path)
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
		m.caseApis[imposter.Id] = &imposter
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
