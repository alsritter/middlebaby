package http

import (
	http "net/http"
	"net/url"

	"alsritter.icu/middlebaby/internal/common"
	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var routeMatch = &mux.RouteMatch{}

type httpImposterHandler struct {
	router    *mux.Router
	imposters []common.HttpImposter
}

func NewHttpImposterHandler(imposters []common.HttpImposter, CORS config.ConfigCORS) *httpImposterHandler {
	router := mux.NewRouter()
	handlers.CORS(common.PrepareAccessControl(CORS)...)(router)

	h := &httpImposterHandler{router: router, imposters: imposters}
	h.addImposterHandler()
	h.printRouter()

	return h
}

func (h *httpImposterHandler) IsHit(r *http.Request) bool {
	return h.router.Match(r, routeMatch)
}

func (h *httpImposterHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(rw, r)
}

// Register proxy request to Router.
// It will match: "path", "host", "method", "params".
func (h *httpImposterHandler) addImposterHandler() {
	for _, imposter := range h.imposters {
		url, err := url.Parse(imposter.Request.Url)
		if err != nil {
			log.Error(err)
			continue
		}

		r := h.router.
			HandleFunc(url.Path, ImposterHandler(imposter)).
			Methods(imposter.Request.Method).
			Host(url.Host)

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
	}
}

// print all router.
func (h *httpImposterHandler) printRouter() {
	log.Debug("print all http router:")
	h.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		met, err1 := route.GetMethods()
		tpl, err2 := route.GetPathTemplate()
		host, err3 := route.GetHostTemplate()
		queries, err4 := route.GetQueriesTemplates()
		log.Debugf(`
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
