package httphandler

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/alsritter/middlebaby/pkg/apimanager"
	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"

	"github.com/alsritter/middlebaby/pkg/util/goproxy"
)

type delegateHandler struct {
	logger.Logger
	apiManager   apimanager.Provider
	enableDirect bool
}

// Connect check the request type.
func (e *delegateHandler) Connect(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) Auth(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) BeforeRequest(ctx *goproxy.Context) {
	body, err := ioutil.ReadAll(ctx.Req.Body)
	ctx.Req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		e.Error(nil, "read request body error: %v", err)
		ctx.Abort()
		return
	}

	resp, err := e.apiManager.MockResponse(context.TODO(), &interact.Request{
		Protocol: interact.ProtocolHTTP,
		Method:   ctx.Req.Method,
		Host:     ctx.Req.Host,
		Path:     ctx.Req.URL.Path,
		Header:   ctx.Req.Header,
		Body:     body,
	})

	if err != nil {
		e.Warn(nil, "%v", err)
		if !e.enableDirect {
			ctx.Resp = &http.Response{
				Status:     http.StatusText(http.StatusInternalServerError),
				StatusCode: http.StatusInternalServerError,
				Proto:      ctx.Req.Proto,
				ProtoMajor: ctx.Req.ProtoMajor,
				ProtoMinor: ctx.Req.ProtoMinor,
				Header:     http.Header{},
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
			}
			ctx.IsFailFast()
		}
		return
	}

	var b io.ReadCloser
	if resp.Body != nil {
		b = ioutil.NopCloser(bytes.NewReader([]byte(resp.GetBodyString())))
	} else {
		b = ioutil.NopCloser(bytes.NewReader([]byte("")))
	}

	e.Debug(nil, "mock [%v] request successful", ctx.Req.URL)
	ctx.IsNeedMock()
	ctx.Resp = &http.Response{
		Status:     http.StatusText(resp.Status),
		StatusCode: resp.Status,
		Proto:      ctx.Req.Proto,
		ProtoMajor: ctx.Req.ProtoMajor,
		ProtoMinor: ctx.Req.ProtoMinor,
		Header:     resp.Header,
		Body:       b,
	}
}

func (e *delegateHandler) BeforeResponse(ctx *goproxy.Context, resp *http.Response, err error) {}

func (e *delegateHandler) ParentProxy(request *http.Request) (*url.URL, error) {
	def := goproxy.DefaultDelegate{}
	return def.ParentProxy(request)
}

func (e *delegateHandler) Finish(ctx *goproxy.Context) {}

func (e *delegateHandler) ErrorLog(err error) {
	e.Error(nil, "request failed %v", err)
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
