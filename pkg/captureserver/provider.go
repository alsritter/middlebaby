package captureserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/alsritter/middlebaby/pkg/captureserver/grpchandler"
	"github.com/alsritter/middlebaby/pkg/captureserver/httphandler"
	"github.com/alsritter/middlebaby/pkg/messagepush"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/alsritter/middlebaby/pkg/util/mbcontext"
	"github.com/spf13/pflag"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Config struct {
	CapturePort int `json:"capturePort" yaml:"capturePort"`
}

func NewConfig() *Config {
	return &Config{
		CapturePort: 58321,
	}
}

func (c *Config) Validate() error {
	if c.CapturePort == 0 {
		return errors.New("[capture-server] capture server listener port cannot be empty")
	}
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type Provider interface {
	GetPort() int
	Start(ctx *mbcontext.Context) error
}

type captrueServer struct {
	logger.Logger
	cfg        *Config
	server     *http.Server
	httpServer http.Handler
	grpcServer http.Handler

	grpcProvider grpchandler.Provider
	httpProvider httphandler.Provider
}

func New(log logger.Logger, cfg *Config, protoManager protomanager.Provider) Provider {
	return &captrueServer{
		Logger:       log.NewLogger("capture"),
		cfg:          cfg,
		server:       &http.Server{},
		grpcProvider: grpchandler.New(log, protoManager),
		httpProvider: httphandler.New(log),
	}
}

// GetPort implements Provider
func (c *captrueServer) GetPort() int {
	return c.cfg.CapturePort
}

func (m *captrueServer) Start(ctx *mbcontext.Context) error {
	if err := m.grpcProvider.Init(ctx); err != nil {
		return err
	}

	m.grpcServer = m.grpcProvider.GetServer()
	m.httpServer = m.httpProvider.GetServer()

	util.StartServiceAsync(ctx, m, func() error {
		return m.start()
	}, func() error {
		return m.close()
	})
	return nil
}

func (m *captrueServer) start() error {
	messagepush.InitMessagePush()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", m.cfg.CapturePort))
	if err != nil {
		return fmt.Errorf("failed to listen the port: %d, err: %v", m.cfg.CapturePort, err)
	}

	// call ServeHTTP function handle request.
	// support HTTP2.0 with h2c package.
	m.server.Handler = h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(m.WithContext(r.Context()).Begin())
		m.WithContext(r.Context()).Info(nil, "====================================Begin===================================")
		if r.ProtoMajor == 2 && strings.HasPrefix(
			r.Header.Get("Content-Type"), "application/grpc") {
			m.grpcServer.ServeHTTP(w, r)
		} else {
			m.httpServer.ServeHTTP(w, r)
		}
		m.WithContext(r.Context()).Info(nil, "====================================End====================================")
		m.WithContext(r.Context()).Done()
	}), &http2.Server{})

	if err := http2.ConfigureServer(m.server, &http2.Server{}); err != nil {
		return fmt.Errorf("proxy http2 error: %v", err)
	}

	m.Info(nil, "Mock server started, Listen port: %d", m.cfg.CapturePort)
	if err := m.server.Serve(l); err != nil {
		if err.Error() != "http: Server closed" {
			return fmt.Errorf("failed to start the proxy server: %v", err)
		}
	}

	return nil
}

// Close shutdown the current http server
func (m *captrueServer) close() error {
	m.Info(nil, "stopping server...")
	if err := m.server.Shutdown(context.TODO()); err != nil {
		return fmt.Errorf("server Shutdown failed: [%v]", err)
	}
	return nil
}
