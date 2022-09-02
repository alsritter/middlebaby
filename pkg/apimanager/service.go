package apimanager

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"github.com/gorilla/mux"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/spf13/pflag"
)

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	// LoadCaseEnv Initialize the environment before executing the use case
	LoadCaseEnv(itfName, caseName string)
	// MockResponse Mock Request.
	MockResponse(ctx context.Context, request *interact.Request) (*interact.Response, error)
	// ClearCaseEnv clear environment
	ClearCaseEnv()
}

type Manager struct {
	caseApis   []*interact.ImposterCase
	itfApis    []*interact.ImposterCase
	globalApis []*interact.ImposterCase
	cfg        *Config
	logger.Logger
	lock         sync.RWMutex
	caseProvider caseprovider.Provider
}

func New(log logger.Logger, cfg *Config, caseProvider caseprovider.Provider) Provider {
	return &Manager{
		cfg:          cfg,
		caseProvider: caseProvider,
		Logger:       log.NewLogger("proto"),
		caseApis:     make([]*interact.ImposterCase, 0),
		itfApis:      make([]*interact.ImposterCase, 0),
		globalApis:   make([]*interact.ImposterCase, 0),
	}
}

func (m *Manager) MockResponse(ctx context.Context, request *interact.Request) (*interact.Response, error) {
	api, isMock := m.MatchAPI(request)
	if !isMock {
		return nil, fmt.Errorf("cannot mock http request: %v", request)
	}
	return &api.Response, nil
}

// MatchAPI is used to match MockAPI
func (m *Manager) MatchAPI(req *interact.Request) (*interact.ImposterCase, bool) {
	m.lock.RLock()
	caseApis := m.caseApis
	itfApis := m.itfApis
	globalApis := m.globalApis
	m.lock.RUnlock()

	// Matching Priority: case -> interface -> global
	for _, api := range caseApis {
		if m.match(req, &api.Request) {
			return api, true
		}
	}

	for _, api := range itfApis {
		if m.match(req, &api.Request) {
			return api, true
		}
	}

	for _, api := range globalApis {
		if m.match(req, &api.Request) {
			return api, true
		}
	}

	return nil, false
}

func (m *Manager) LoadCaseEnv(itfName, caseName string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.globalApis = m.caseProvider.GetMockCasesFromGlobals()
	m.caseApis = m.caseProvider.GetMockCasesFromCase(itfName, caseName)
	m.itfApis = m.caseProvider.GetMockCasesFromItf(itfName)
}

func (m *Manager) ClearCaseEnv() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.globalApis = make([]*interact.ImposterCase, 0)
	m.caseApis = make([]*interact.ImposterCase, 0)
	m.itfApis = make([]*interact.ImposterCase, 0)
}

func (m *Manager) match(req, target *interact.Request) bool {
	// use mux match single router
	var match mux.RouteMatch
	matched := mux.NewRouter().
		Host(target.Host).
		Methods(target.Method).
		Path(target.Path).
		Match(&http.Request{
			Method: req.Method,
			URL:    &url.URL{Path: req.Path},
			Host:   req.Host,
		}, &match)
	if !matched {
		return false
	}

	if err := assert.So(m, "mock header assert", req.Headers, target.Headers); err != nil {
		m.Trace(nil, "mock head cannot hit expected:[%v] actual:[%v]", target.Headers, req.Headers)
		return false
	}

	if err := assert.So(m, "mock params assert", target.Params, req.Params); err != nil {
		m.Trace(nil, "mock params cannot hit expected:[%v] actual:[%v]", target.Params, req.Params)
		return false
	}

	if req.Body != nil && target.Body != nil {
		if err := assert.So(m, "mock body assert", target.Body.Bytes(), req.Body.Bytes()); err != nil {
			m.Trace(nil, "mock body cannot hit expected:[%s] actual:[%s]", target.Body.Bytes(), req.Body.Bytes())
			return false
		}
	}
	return true
}
