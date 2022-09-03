package httphandler

import (
	"net/http"
	"net/http/httptrace"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/util/goproxy"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

// Config defines the config structure
type Config struct {
	EnableDirect bool
}

type Provider interface {
	GetServer() http.Handler
}

type mockServer struct {
	*goproxy.Proxy
	logger.Logger
}

func New(log logger.Logger, cfg *Config, apiManager apimanager.Provider) Provider {
	l := log.NewLogger("http")
	return &mockServer{
		Logger: log.NewLogger("http"),
		Proxy: goproxy.New(goproxy.WithDelegate(&delegateHandler{
			Logger:       l,
			apiManager:   apiManager,
			enableDirect: cfg.EnableDirect,
		}),
			goproxy.WithDecryptHTTPS(&cache{}),
			goproxy.WithClientTrace(&httptrace.ClientTrace{
				DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {},
				GotConn: func(connInfo httptrace.GotConnInfo) {},
			}),
		),
	}
}

func (m *mockServer) GetServer() http.Handler {
	return m
}
