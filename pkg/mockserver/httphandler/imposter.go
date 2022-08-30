package httphandler

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type httpImposterHandler struct {
	log logger.Logger
}

func NewHttpImposterHandler(log logger.Logger) *httpImposterHandler {
	return &httpImposterHandler{log: log}
}

func (h *httpImposterHandler) ImposterHandler(imposter interact.ImposterCase) http.HandlerFunc {
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

func (h *httpImposterHandler) writeHeaders(imposter interact.ImposterCase, w http.ResponseWriter) {
	if imposter.Response.Headers == nil {
		return
	}

	for key, val := range imposter.Response.Headers {
		w.Header().Set(key, val)
	}
}

func (h *httpImposterHandler) writeBody(imposter interact.ImposterCase, w http.ResponseWriter) {
	_, _ = w.Write(imposter.Response.Body.Bytes())
}
