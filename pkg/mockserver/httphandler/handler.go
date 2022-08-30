package httphandler

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"

	"github.com/alsritter/middlebaby/pkg/util/goproxy"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

type delegateHandler struct {
	enableDirect bool
}

type MiddlemanProxy struct {
	*goproxy.Proxy
	log logger.Logger
}

func NewProxy(enableDirect bool, log logger.Logger) *MiddlemanProxy {
	return &MiddlemanProxy{
		log: log.NewLogger("mit-proxy"),
		Proxy: goproxy.New(goproxy.WithDelegate(&delegateHandler{
			enableDirect: enableDirect,
		}),
			goproxy.WithDecryptHTTPS(&cache{}),
			goproxy.WithClientTrace(&httptrace.ClientTrace{
				DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
					log.Trace(nil, "DNS Info: %+v.", dnsInfo)
				},
				GotConn: func(connInfo httptrace.GotConnInfo) {
					log.Trace(nil, "Got Conn: %+v.", connInfo)
				},
			}),
		),
	}
}

// Connect check the request type.
func (e *delegateHandler) Connect(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) Auth(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) BeforeRequest(ctx *goproxy.Context) {}

func (e *delegateHandler) BeforeResponse(ctx *goproxy.Context, resp *http.Response, err error) {}

func (e *delegateHandler) ParentProxy(request *http.Request) (*url.URL, error) {
	def := goproxy.DefaultDelegate{}
	return def.ParentProxy(request)
}

func (e *delegateHandler) Finish(ctx *goproxy.Context) {}

func (e *delegateHandler) ErrorLog(err error) {
}

// 实现证书缓存接口
type cache struct {
	m sync.Map
}

func (c *cache) Set(host string, cert *tls.Certificate) {
	c.m.Store(host, cert)
}

func (c *cache) Get(host string) *tls.Certificate {
	v, ok := c.m.Load(host)
	if !ok {
		return nil
	}

	return v.(*tls.Certificate)
}
