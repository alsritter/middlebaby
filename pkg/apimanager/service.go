package apimanager

import (
	"context"
	"fmt"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/util/assert"
	"net/http"
	"sync"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
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
	caseApis     []*interact.ImposterCase
	itfApis      []*interact.ImposterCase
	globalApis   []*interact.ImposterCase
	cfg          *Config
	log          logger.Logger
	lock         sync.RWMutex
	caseProvider caseprovider.Provider
}

func New(log logger.Logger, cfg *Config, caseProvider caseprovider.Provider) Provider {
	return &Manager{
		caseApis:     make([]*interact.ImposterCase, 0),
		itfApis:      make([]*interact.ImposterCase, 0),
		globalApis:   make([]*interact.ImposterCase, 0),
		log:          log.NewLogger("api"),
		cfg:          cfg,
		caseProvider: caseProvider,
	}
}

func (m *Manager) Start() error {
	return nil
}

func (m *Manager) Close() error {
	return nil
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

func (m *Manager) match(req, target *interact.Request) bool {
	if err := assert.So(m.log, "mock header assert", req.Headers, target.Headers); err != nil {
		m.log.Trace(nil, "mock head cannot hit expected:[%v] actual:[%v]", target.Headers, req.Headers)
		return false
	}

	if err := assert.So(m.log, "mock params assert", target.Params, req.Params); err != nil {
		m.log.Trace(nil, "mock params cannot hit expected:[%v] actual:[%v]", target.Params, req.Params)
		return false
	}

	if err := assert.So(m.log, "mock body assert", target.Body.Bytes(), req.Body.Bytes()); err != nil {
		m.log.Trace(nil, "mock body cannot hit expected:[%s] actual:[%s]", target.Body.Bytes(), req.Body.Bytes())
		return false
	}

	return true
}

func (m *Manager) LoadCaseEnv(itfName, caseName string) {
	m.loadCaseImposter(itfName, caseName)
}

func (m *Manager) MockResponse(ctx context.Context, request *interact.Request) (*interact.ImposterCase, error) {
	api, isMock := m.MatchAPI(request)
	if !isMock {
		return nil, fmt.Errorf("cannot mock http request: %v", request)
	}
	return nil, fmt.Errorf("cannot mock http request: %v", request)
}

func (m *Manager) toHttpHeader(headers map[string]interface{}) (httpHeader http.Header) {
	httpHeader = make(http.Header)
	for k, v := range headers {
		switch vv := v.(type) {
		case string:
			httpHeader.Add(k, vv)
		case []string:
			for _, vvv := range vv {
				httpHeader.Add(k, vvv)
			}
		}
	}
	return
}
