package http

import (
	http "net/http"
	"net/http/httputil"
	"time"

	"alsritter.icu/middlebaby/internal/file/common"
	"alsritter.icu/middlebaby/internal/log"
)

// ImposterHandler create specific handler for the received imposter
func ImposterHandler(imposter common.HttpImposter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Trace("hit mock:", r.URL.String())
		dump, _ := httputil.DumpRequest(r, true)
		log.Tracef("proxy request: %s", dump)

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
