package httphandler

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
		e.Error(nil, "read request body error: %w", err)
		ctx.Abort()
		return
	}

	resp, err := e.apiManager.MockResponse(context.TODO(), &interact.Request{
		Protocol: interact.ProtocolHTTP,
		Method:   ctx.Req.Method,
		Host:     ctx.Req.Host,
		Path:     ctx.Req.URL.Path,
		Headers:  getHeadersFromHttpHeaders(ctx.Req.Header),
		Body:     interact.NewBytesMessage(body),
	})

	if err != nil {
		e.Warn(nil, "%w", err)

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

	ctx.IsNeedMock()
	ctx.Resp = &http.Response{
		Status:     http.StatusText(resp.Status),
		StatusCode: resp.Status,
		Proto:      ctx.Req.Proto,
		ProtoMajor: ctx.Req.ProtoMajor,
		ProtoMinor: ctx.Req.ProtoMinor,
		Header:     resp.Headers,
		Body:       ioutil.NopCloser(bytes.NewReader(resp.Body.Bytes())),
	}
}

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

// getHeadersFromHttpHeaders is used to get map[string]string from http.Header
func getHeadersFromHttpHeaders(input http.Header) map[string]interface{} {
	headers := map[string]interface{}{}
	for key, values := range input {
		key = strings.ToLower(key)
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
