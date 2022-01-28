package proxy

import (
	"net/http"

	"alsritter.icu/middlebaby/internal/log"
)

type Proxy interface {
	IsHit(r *http.Request) bool
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
}

type Direct interface {
	IsHit(r *http.Request) bool
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
}

type mockList struct {
	// whether to proxy directly to the real interface when the mock interface is not hit
	enableDirect     bool
	reverseProxyList []Proxy
	directList       []Direct
}

func NewMockList(enableDirect bool) *mockList {
	return &mockList{enableDirect: enableDirect}
}

func (p *mockList) AddProxy(proxy Proxy) {
	p.reverseProxyList = append(p.reverseProxyList, proxy)
}

func (p *mockList) AddDirect(direct Direct) {
	p.directList = append(p.directList, direct)
}

// find request and forward requests to the proxy for processing.
func (m *mockList) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Trace("target request:", r.Host, r.Method, r.URL.Path)

	for _, h := range m.reverseProxyList {
		if h.IsHit(r) {
			h.ServeHTTP(rw, r)
		}
	}

	if !m.enableDirect {
		return
	}

	// call direct request.
	for _, h := range m.directList {
		if h.IsHit(r) {
			h.ServeHTTP(rw, r)
		}
	}
}
