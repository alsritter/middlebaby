package mockserver

import (
	"crypto/tls"
	"github.com/alsritter/middlebaby/pkg/util/goproxy"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
)

type eventHandler struct {
	enableDirect bool
}

type mitmproxy struct {
	px  *goproxy.Proxy
	log logger.Logger
}

func NewProxy(enableDirect bool, log logger.Logger) *mitmproxy {
	return &mitmproxy{
		log: log.NewLogger("mitmproxy"),
		px: goproxy.New(goproxy.WithDelegate(&eventHandler{
			enableDirect: enableDirect,
		}), goproxy.WithDecryptHTTPS(&cache{}),
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
func (e *eventHandler) Connect(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *eventHandler) Auth(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *eventHandler) BeforeRequest(ctx *goproxy.Context) {}

func (e *eventHandler) BeforeResponse(ctx *goproxy.Context, resp *http.Response, err error) {}

func (e *eventHandler) ParentProxy(request *http.Request) (*url.URL, error) {
	def := goproxy.DefaultDelegate{}
	return def.ParentProxy(request)
}

func (e *eventHandler) Finish(ctx *goproxy.Context) {}

func (e *eventHandler) ErrorLog(err error) {
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
