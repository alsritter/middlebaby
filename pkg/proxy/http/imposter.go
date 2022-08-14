package http

import (
	http "net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/alsritter/middlebaby/internal/file/config"
	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var routeMatch = &mux.RouteMatch{}

type httpImposterHandler struct {
	router     *mux.Router
	mockCenter apimanager.MockCenter
	log        logger.Logger
}

func NewHttpImposterHandler(mockCenter apimanager.MockCenter, CORS config.ConfigCORS, log logger.Logger) *httpImposterHandler {
	router := mux.NewRouter()
	handlers.CORS(util.PrepareAccessControl(CORS)...)(router)
	h := &httpImposterHandler{router: router, mockCenter: mockCenter}
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
	for _, imposter := range h.mockCenter.GetAllHttp() {
		url, err := url.Parse(imposter.Request.Url)
		if err != nil {
			h.log.Error(nil, err.Error())
			continue
		}

		r := h.router.
			HandleFunc(url.Path, h.ImposterHandler(imposter)).
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
	h.log.Debug(nil, "print all http router:")
	h.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		met, err1 := route.GetMethods()
		tpl, err2 := route.GetPathTemplate()
		host, err3 := route.GetHostTemplate()
		queries, err4 := route.GetQueriesTemplates()
		h.log.Debug(nil, `
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

// ImposterHandler create specific handler for the received imposter
func (h *httpImposterHandler) ImposterHandler(imposter interact.HttpImposter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Trace(nil, "hit mock:", r.URL.String())
		dump, _ := httputil.DumpRequest(r, true)
		h.log.Trace(nil, "proxy request: %s", dump)

		if imposter.Delay() > 0 {
			time.Sleep(imposter.Delay())
		}

		h.writeHeaders(imposter, w)
		w.WriteHeader(imposter.Response.Status)
		h.writeBody(imposter, w)
	}
}

func (h *httpImposterHandler) writeHeaders(imposter interact.HttpImposter, w http.ResponseWriter) {
	if imposter.Response.Headers == nil {
		return
	}

	for key, val := range imposter.Response.Headers {
		w.Header().Set(key, val)
	}
}

func (h *httpImposterHandler) writeBody(imposter interact.HttpImposter, w http.ResponseWriter) {
	w.Write([]byte(imposter.Response.Body))
}
