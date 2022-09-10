package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/caseprovider"
	"github.com/alsritter/middlebaby/pkg/protomanager"
	"github.com/alsritter/middlebaby/pkg/targetprocess"
	"github.com/alsritter/middlebaby/pkg/taskserver"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	v1 "github.com/alsritter/middlebaby/web/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
	"github.com/spf13/pflag"
)

type Config struct {
	WebServicePort int32 `json:"port" yaml:"port"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if c.WebServicePort == 0 {
		return errors.New("[web server] web server listener port cannot be empty")
	}
	return nil
}

type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {}

type WebService struct {
	logger.Logger

	cfg    *Config
	api_v1 *v1.API

	apiProvider  apimanager.Provider
	caseProvider caseprovider.Provider
	protoManager protomanager.Provider
	taskService  taskserver.Provider
	target       targetprocess.Provider
}

func New(log logger.Logger,
	cfg *Config,
	apiProvider apimanager.Provider,
	caseProvider caseprovider.Provider,
	protoManager protomanager.Provider,
	taskService taskserver.Provider,
	target targetprocess.Provider) Provider {

	l := log.NewLogger("web")
	return &WebService{
		Logger:       l,
		cfg:          cfg,
		apiProvider:  apiProvider,
		caseProvider: caseProvider,
		protoManager: protoManager,
		taskService:  taskService,
		target:       target,
		api_v1:       v1.NewAPI(l, apiProvider, caseProvider, protoManager, taskService, target),
	}
}

// Start implements Provider
func (w *WebService) Start(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup) error {
	r := gin.Default()
	w.api_v1.Register(r)
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", w.cfg.WebServicePort),
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "PUT"}),
			handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"}),
			handlers.AllowedOrigins([]string{"*"}),
		)(r),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	w.Info(nil, "Web server started, Listen port: %d", w.cfg.WebServicePort)
	util.StartServiceAsync(ctx, w, cancelFunc, wg, func() error {
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	}, func() error {
		return s.Close()
	})

	return nil
}
