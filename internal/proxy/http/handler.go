package http

import (
	http "net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"alsritter.icu/middlebaby/internal/common"
	"alsritter.icu/middlebaby/internal/config"
	"alsritter.icu/middlebaby/internal/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var routeMatch = &mux.RouteMatch{}

type httpRequestHandler struct {
	router    *mux.Router
	imposters []common.HttpImposter
}

func NewHttpRequestHandler(imposters []common.HttpImposter, CORS config.ConfigCORS) *httpRequestHandler {
	router := mux.NewRouter()
	handlers.CORS(common.PrepareAccessControl(CORS)...)(router)

	h := &httpRequestHandler{router: router, imposters: imposters}
	h.addImposterHandler()
	h.printRouter()

	return h
}

func (h *httpRequestHandler) IsHit(r *http.Request) bool {
	return h.router.Match(r, routeMatch)
}

func (h *httpRequestHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(rw, r)
}

// Register proxy request to Router
func (h *httpRequestHandler) addImposterHandler() {
	for _, imposter := range h.imposters {
		url, err := url.Parse(imposter.Request.Url)
		if err != nil {
			log.Error(err)
			continue
		}

		r := h.router.HandleFunc(url.Path, ImposterHandler(imposter)).
			Methods(imposter.Request.Method)

		if imposter.Request.Headers != nil {
			for k, v := range imposter.Request.Headers {
				r.HeadersRegexp(k, v)
			}
		}

		log.Info(imposter.Request.Params)

		if imposter.Request.Params != nil {
			for k, v := range imposter.Request.Params {
				r.Queries(k, v)
			}
		}
	}
}

// print all router.
func (h *httpRequestHandler) printRouter() {
	h.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err1 := route.GetPathTemplate()
		met, err2 := route.GetMethods()
		log.Debugf("path: %s, err: %v,  Method: %v, err2: %v", tpl, err1, met, err2)
		return nil
	})
}

// ImposterHandler create specific handler for the received imposter
func ImposterHandler(imposter common.HttpImposter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dump, _ := httputil.DumpRequest(r, true)
		log.Debugf("proxy request: %s", dump)

		if imposter.Delay() > 0 {
			time.Sleep(imposter.Delay())
		}
		writeHeaders(imposter, w)
		w.WriteHeader(imposter.Response.Status)
		writeBody(imposter, w)
	}
}

func writeHeaders(imposter common.HttpImposter, w http.ResponseWriter) {
	if imposter.Response.Headers == nil {
		return
	}

	for key, val := range imposter.Response.Headers {
		w.Header().Set(key, val)
	}
}

func writeBody(imposter common.HttpImposter, w http.ResponseWriter) {
	w.Write([]byte(imposter.Response.Body))
}
