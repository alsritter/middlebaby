package httphandler

import (
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/util/goproxy"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"net/http"
	"net/http/httptrace"
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
	cfg        *Config
	apiManager apimanager.Provider
}

func New(log logger.Logger, cfg *Config, apiManager apimanager.Provider) Provider {
	return &mockServer{
		Logger: log.NewLogger("mit-proxy"),
		Proxy: goproxy.New(goproxy.WithDelegate(&delegateHandler{
			apiManager:   apiManager,
			enableDirect: cfg.EnableDirect,
		}),
			goproxy.WithDecryptHTTPS(&cache{}),
			goproxy.WithClientTrace(&httptrace.ClientTrace{
				DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
					log.Trace(nil, "DNS Info: %+v.", dnsInfo)
				},
				GotConn: func(connInfo httptrace.GotConnInfo) {
					log.Trace(nil, "Got Conn: %+v.", connInfo)
				},
			}),
		),
	}
}

func (m *mockServer) GetServer() http.Handler {
	return m
}
