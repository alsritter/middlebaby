package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Proxy represent reverse proxy server.
type Proxy struct {
	server *httputil.ReverseProxy
	url    *url.URL
}

// NewProxy creates new proxy server.
func NewProxy(targetHost string) (*Proxy, error) {
	u, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}
	reverseProxy := httputil.NewSingleHostReverseProxy(u)
	return &Proxy{server: reverseProxy, url: u}, nil
}

// Handler returns handler that sends request to another server.
func (p *Proxy) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = p.url.Host
		r.URL.Scheme = p.url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = p.url.Host

		p.server.ServeHTTP(w, r)
	}
}
