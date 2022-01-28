package http

import (
	"net"
	http "net/http"
	"net/http/httputil"
	"time"
)

type httpDirectHandler struct{}

func NewHttpDirectHandler() *httpDirectHandler {
	return &httpDirectHandler{}
}

func (h *httpDirectHandler) IsHit(r *http.Request) bool {
	return r.ProtoMajor == 1
}

func (h *httpDirectHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		return
	}

	p := httputil.ReverseProxy{}
	// prevent the proxy from coming back into this method again.
	p.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	p.Director = func(request *http.Request) {}
	p.ServeHTTP(rw, r)
}
