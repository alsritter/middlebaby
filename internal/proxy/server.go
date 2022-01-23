package proxy

import (
	"context"
	"net/http"
	"net/url"

	"alsritter.icu/middlebaby/internal/log"

	"alsritter.icu/middlebaby/internal/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	defaultCORSMethods        = []string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE", "PATCH", "TRACE", "CONNECT"}
	defaultCORSHeaders        = []string{"X-Requested-With", "Content-Type", "Authorization"}
	defaultCORSExposedHeaders = []string{"Cache-Control", "Content-Language", "Content-Type", "Expires", "Last-Modified", "Pragma"}
)

// Server definition of mock server
type Server struct {
	router     *mux.Router
	httpServer *http.Server
	// proxy      *Proxy
	imposters []Imposter
}

// NewServer initialize the mock server
func NewServer(r *mux.Router, httpServer *http.Server, imposters []Imposter) Server {
	return Server{
		router:     r,
		httpServer: httpServer,
		imposters:  imposters,
	}
}

// PrepareAccessControl Return options to initialize the mock server with default access control
func PrepareAccessControl(config config.ConfigCORS) (h []handlers.CORSOption) {
	h = append(h, handlers.AllowedMethods(defaultCORSMethods))
	h = append(h, handlers.AllowedHeaders(defaultCORSHeaders))
	h = append(h, handlers.ExposedHeaders(defaultCORSExposedHeaders))

	if len(config.Methods) > 0 {
		h = append(h, handlers.AllowedMethods(config.Methods))
	}

	if len(config.Origins) > 0 {
		h = append(h, handlers.AllowedOrigins(config.Origins))
	}

	if len(config.Headers) > 0 {
		h = append(h, handlers.AllowedHeaders(config.Headers))
	}

	if len(config.ExposedHeaders) > 0 {
		h = append(h, handlers.ExposedHeaders(config.ExposedHeaders))
	}

	if config.AllowCredentials {
		h = append(h, handlers.AllowCredentials())
	}

	return
}

// Run run launch a previous configured http server if any error happens while the starting process
// application will be crashed
func (s *Server) Run() {
	go func() {
		log.Infof("The fake server is on tap now: %s\n", s.httpServer.Addr)
		err := s.run()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
}

func (s *Server) run() error {
	s.addImposterHandler()
	s.printRouter()
	return s.httpServer.ListenAndServe()
}

// Shutdown shutdown the current http server
func (s *Server) Shutdown() error {
	log.Info("stopping server...")
	if err := s.httpServer.Shutdown(context.TODO()); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	return nil
}

// Register proxy request to Router
func (s *Server) addImposterHandler() {
	for _, imposter := range s.imposters {
		url, err := url.Parse(imposter.Request.Url)
		if err != nil {
			log.Error(err)
			continue
		}

		r := s.router.HandleFunc(url.Path, ImposterHandler(imposter)).
			Methods(imposter.Request.Method)

		if imposter.Request.Headers != nil {
			for k, v := range *imposter.Request.Headers {
				r.HeadersRegexp(k, v)
			}
		}

		log.Info(imposter.Request.Params)

		if imposter.Request.Params != nil {
			for k, v := range *imposter.Request.Params {
				r.Queries(k, v)
			}
		}
	}
}

func (s *Server) printRouter() {
	s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err1 := route.GetPathTemplate()
		met, err2 := route.GetMethods()
		log.Debugf("path: %s, err: %v,  Method: %v, err2: %v", tpl, err1, met, err2)
		return nil
	})
}
