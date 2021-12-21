package proxy

import (
	"net/http"
	"net/http/httputil"

	"alsritter.icu/middlebaby/internal/log"
)

// 代理的请求
type Proxy interface {
	// IsHit 是否命中对应 url
	IsHit(r *http.Request) bool
	// ServeHTTP 具体的代理
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
}

// 真实的请求
type Direct interface {
	// IsHit 是否命中对应 url
	IsHit(r *http.Request) bool
	// ServeHTTP 具体的代理
	ServeHTTP(rw http.ResponseWriter, r *http.Request)
}

// 存储代理对象和真实请求对象
type proxyList struct {
	reverseProxyList []Proxy  // 代理服务
	directList       []Direct // 直接请求服务
}

func NewProxyList() *proxyList {
	return &proxyList{}
}

// 添加代理对象
func (p *proxyList) AddProxy(proxy Proxy) {
	p.reverseProxyList = append(p.reverseProxyList, proxy)
}

// 添加真实请求对象
func (p *proxyList) AddDirect(direct Direct) {
	p.directList = append(p.directList, direct)
}

func (p *proxyList) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	dump, _ := httputil.DumpRequest(r, true)
	log.Debugf("代理请求数据: %s\n", dump)

	// 检查是否命中代理
	for i := 0; i < len(p.reverseProxyList); i++ {
		if p.reverseProxyList[i].IsHit(r) {
			p.reverseProxyList[i].ServeHTTP(rw, r)
			dumpResp, _ := httputil.DumpRequestOut(r, true)
			log.Debugf("代理的响应数据: %s\n", dumpResp)
			return
		}
	}

	log.Debug("未命中 mock, 直接请求接口")

	/// 检查是否命中真实的请求对象
	for _, direct := range p.directList {
		if direct.IsHit(r) {
			direct.ServeHTTP(rw, r)
			return
		}
	}
}
