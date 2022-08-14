package http

import (
	"io"
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
	// forwarding the HTTPS
	if r.Method == http.MethodConnect {
		handleTunneling(rw, r)
		return
	}

	// forwarding the HTTP
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

// reference: https://gist.github.com/wwek/41790cbef2e33b6065eaea688ea54760
func handleTunneling(w http.ResponseWriter, r *http.Request) {
	// Setting timeout prevents server resources from being occupied due to a large number of timeouts
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	// type conversion
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}

// forward connected data
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
