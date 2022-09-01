// Copyright 2018 ouqiang authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package goproxy HTTP(S)代理, 支持中间人代理解密HTTPS数据
package goproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alsritter/middlebaby/pkg/util/goproxy/cert"
	"github.com/viki-org/dnscache"
)

const (
	// 连接目标服务器超时时间
	defaultTargetConnectTimeout = 5 * time.Second
	// 目标服务器读写超时时间
	defaultTargetReadWriteTimeout = 10 * time.Second
)

type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

// 隧道连接成功响应行
var tunnelEstablishedResponseLine = []byte("HTTP/1.1 200 Connection established\r\n\r\n")

var badGateway = []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadGateway, http.StatusText(http.StatusBadGateway)))

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}

	ctxPool = sync.Pool{
		New: func() interface{} {
			return new(Context)
		},
	}
	headerPool  = NewHeaderPool()
	requestPool = newRequestPool()
)

type RequestPool struct {
	pool sync.Pool
}

func newRequestPool() *RequestPool {
	return &RequestPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(http.Request)
			},
		},
	}
}

func (p *RequestPool) Get() *http.Request {
	req := p.pool.Get().(*http.Request)

	req.Method = ""
	req.URL = nil
	req.Proto = ""
	req.ProtoMajor = 0
	req.ProtoMinor = 0
	req.Header = nil
	req.Body = nil
	req.GetBody = nil
	req.ContentLength = 0
	req.TransferEncoding = nil
	req.Close = false
	req.Host = ""
	req.Form = nil
	req.PostForm = nil
	req.MultipartForm = nil
	req.Trailer = nil
	req.RemoteAddr = ""
	req.RequestURI = ""
	req.TLS = nil
	req.Cancel = nil
	req.Response = nil

	return req
}

func (p *RequestPool) Put(req *http.Request) {
	if req != nil {
		p.pool.Put(req)
	}
}

type HeaderPool struct {
	pool sync.Pool
}

func NewHeaderPool() *HeaderPool {
	return &HeaderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return http.Header{}
			},
		},
	}
}

func (p *HeaderPool) Get() http.Header {
	header := p.pool.Get().(http.Header)
	for k := range header {
		delete(header, k)
	}

	return header
}

func (p *HeaderPool) Put(header http.Header) {
	if header != nil {
		p.pool.Put(header)
	}
}

// 生成隧道建立请求行
func makeTunnelRequestLine(addr string) string {
	return fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", addr)
}

type options struct {
	disableKeepAlive bool
	delegate         Delegate

	decryptHTTPS bool
	certCache    cert.Cache
	transport    *http.Transport
	clientTrace  *httptrace.ClientTrace
}

type Option func(*options)

// WithDisableKeepAlive 连接是否重用
func WithDisableKeepAlive(disableKeepAlive bool) Option {
	return func(opt *options) {
		opt.disableKeepAlive = disableKeepAlive
	}
}

func WithClientTrace(t *httptrace.ClientTrace) Option {
	return func(opt *options) {
		opt.clientTrace = t
	}
}

// WithDelegate 设置委托类
func WithDelegate(delegate Delegate) Option {
	return func(opt *options) {
		opt.delegate = delegate
	}
}

// WithTransport 自定义http transport
func WithTransport(t *http.Transport) Option {
	return func(opt *options) {
		opt.transport = t
	}
}

// WithDecryptHTTPS 中间人代理, 解密HTTPS, 需实现证书缓存接口
func WithDecryptHTTPS(c cert.Cache) Option {
	return func(opt *options) {
		opt.decryptHTTPS = true
		opt.certCache = c
	}
}

// New 创建proxy实例
func New(opt ...Option) *Proxy {
	opts := &options{}
	for _, o := range opt {
		o(opts)
	}
	if opts.delegate == nil {
		opts.delegate = &DefaultDelegate{}
	}
	if opts.transport == nil {
		opts.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConns:          100,
			MaxConnsPerHost:       10,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	p := &Proxy{}
	p.delegate = opts.delegate
	p.decryptHTTPS = opts.decryptHTTPS
	if p.decryptHTTPS {
		p.cert = cert.NewCertificate(opts.certCache, true)
	}
	p.transport = opts.transport
	p.transport.DialContext = p.dialContext()
	p.dnsCache = dnscache.New(5 * time.Minute)
	p.transport.DisableKeepAlives = opts.disableKeepAlive
	p.transport.Proxy = p.delegate.ParentProxy
	p.clientTrace = opts.clientTrace

	return p
}

// Proxy 实现了http.Handler接口
type Proxy struct {
	delegate      Delegate
	clientConnNum int32
	decryptHTTPS  bool
	cert          *cert.Certificate
	transport     *http.Transport
	clientTrace   *httptrace.ClientTrace
	dnsCache      *dnscache.Resolver
}

var _ http.Handler = &Proxy{}

// ServeHTTP 实现了http.Handler接口
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	atomic.AddInt32(&p.clientConnNum, 1)
	ctx := ctxPool.Get().(*Context)
	ctx.Reset(req)

	defer func() {
		p.delegate.Finish(ctx)
		ctxPool.Put(ctx)
		atomic.AddInt32(&p.clientConnNum, -1)
	}()
	p.delegate.Connect(ctx, rw)
	if ctx.abort {
		return
	}
	p.delegate.Auth(ctx, rw)
	if ctx.abort {
		return
	}

	switch {
	case ctx.IsHTTPS() || (ctx.Req.Method == http.MethodConnect && ctx.Req.URL.Port() == "443"):
		p.tunnelProxy(ctx, rw)
	default:
		p.httpProxy(ctx, rw)
	}
}

// ClientConnNum 获取客户端连接数
func (p *Proxy) ClientConnNum() int32 {
	return atomic.LoadInt32(&p.clientConnNum)
}

// DoRequest 执行HTTP请求，并调用responseFunc处理response
func (p *Proxy) DoRequest(ctx *Context, responseFunc func(*http.Response, error)) {
	if ctx.Data == nil {
		ctx.Data = make(map[interface{}]interface{})
	}
	p.delegate.BeforeRequest(ctx)
	if ctx.abort {
		return
	}

	if ctx.failFast || ctx.needMock {
		responseFunc(ctx.Resp, nil)
		return
	}

	newReq := requestPool.Get()
	*newReq = *ctx.Req
	newHeader := headerPool.Get()
	CloneHeader(newReq.Header, newHeader)
	newReq.Header = newHeader
	for _, item := range hopHeaders {
		if newReq.Header.Get(item) != "" {
			newReq.Header.Del(item)
		}
	}
	if p.clientTrace != nil {
		newReq = newReq.WithContext(httptrace.WithClientTrace(newReq.Context(), p.clientTrace))
	}

	resp, err := p.transport.RoundTrip(newReq)
	p.delegate.BeforeResponse(ctx, resp, err)
	if ctx.abort {
		return
	}

	if err == nil {
		for _, h := range hopHeaders {
			resp.Header.Del(h)
		}
	}
	responseFunc(resp, err)
	headerPool.Put(newHeader)
	requestPool.Put(newReq)
}

// HTTP代理
func (p *Proxy) httpProxy(ctx *Context, rw http.ResponseWriter) {
	ctx.Req.URL.Scheme = "http"
	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTP请求错误: %s", ctx.Req.URL, err))
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		CopyHeader(rw.Header(), resp.Header)
		rw.WriteHeader(resp.StatusCode)
		buf := bufPool.Get().([]byte)
		_, _ = io.CopyBuffer(rw, resp.Body, buf)
		bufPool.Put(buf)
	})
}

// HTTPS代理
func (p *Proxy) httpsProxy(ctx *Context, tlsClientConn *tls.Conn) {
	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 请求错误: %s", ctx.Req.URL, err))
			_, _ = tlsClientConn.Write(badGateway)
			return
		}
		err = resp.Write(tlsClientConn)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, response写入客户端失败, %s", ctx.Req.URL, err))
		}
		_ = resp.Body.Close()
	})
}

// 隧道代理
func (p *Proxy) tunnelProxy(ctx *Context, rw http.ResponseWriter) {
	clientConn, err := hijacker(rw)
	if err != nil {
		p.delegate.ErrorLog(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() {
		_ = clientConn.Close()
	}()

	parentProxyURL, err := p.delegate.ParentProxy(ctx.Req)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - Failed to resolve the proxy address: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	if parentProxyURL == nil {
		_, err = clientConn.Write(tunnelEstablishedResponseLine)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - 隧道连接成功,通知客户端错误: %s", ctx.Req.URL.Host, err))
			return
		}
	}

	var tlsClientConn *tls.Conn
	if p.decryptHTTPS {
		tlsConfig, err := p.cert.GenerateTlsConfig(ctx.Req.URL.Host)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 生成证书失败: %s", ctx.Req.URL.Host, err))
			return
		}
		tlsClientConn = tls.Server(clientConn, tlsConfig)
		defer func() {
			_ = tlsClientConn.Close()
		}()
		if err := tlsClientConn.Handshake(); err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 握手失败: %s", ctx.Req.URL.Host, err))
			return
		}

		buf := bufio.NewReader(tlsClientConn)
		tlsReq, err := http.ReadRequest(buf)
		if err != nil {
			if err != io.EOF {
				p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 读取客户端请求失败: %s", ctx.Req.URL.Host, err))
			}
			return
		}
		tlsReq.RemoteAddr = ctx.Req.RemoteAddr
		tlsReq.URL.Scheme = "https"
		tlsReq.URL.Host = tlsReq.Host
		ctx.Req = tlsReq
	}

	targetAddr := ctx.Req.URL.Host
	if parentProxyURL != nil {
		targetAddr = parentProxyURL.Host
	}
	if !strings.Contains(targetAddr, ":") {
		targetAddr += ":443"
	}

	if p.decryptHTTPS {
		p.httpsProxy(ctx, tlsClientConn)
	} else {
		targetConn, err := net.DialTimeout("tcp", targetAddr, defaultTargetConnectTimeout)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - 隧道转发连接目标服务器失败: %s", ctx.Req.URL.Host, err))
			return
		}
		defer func() {
			_ = targetConn.Close()
		}()
		if parentProxyURL != nil {
			tunnelRequestLine := makeTunnelRequestLine(ctx.Req.URL.Host)
			_, _ = targetConn.Write([]byte(tunnelRequestLine))
		}
		p.tunnelConnected(ctx, nil)
		p.transfer(clientConn, targetConn)
	}
}

// Bidirectional forwarding
func (p *Proxy) transfer(src net.Conn, dst net.Conn) {
	go func() {
		buf := bufPool.Get().([]byte)
		_, err := io.CopyBuffer(src, dst, buf)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("隧道双向转发错误: [%s -> %s] %s", dst.RemoteAddr().String(), src.RemoteAddr().String(), err))
		}
		bufPool.Put(buf)
		_ = src.Close()
		_ = dst.Close()
	}()

	buf := bufPool.Get().([]byte)
	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("隧道双向转发错误: [%s -> %s] %s", src.RemoteAddr().String(), dst.RemoteAddr().String(), err))
	}
	bufPool.Put(buf)
	_ = dst.Close()
	_ = src.Close()
}

func (p *Proxy) tunnelConnected(ctx *Context, err error) {
	ctx.TunnelProxy = true
	p.delegate.BeforeRequest(ctx)
	if err != nil {
		p.delegate.BeforeResponse(ctx, nil, err)
		return
	}

	if ctx.failFast || ctx.needMock {
		p.delegate.BeforeResponse(ctx, ctx.Resp, nil)
		return
	}

	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	p.delegate.BeforeResponse(ctx, resp, nil)
}

func (p *Proxy) dialContext() DialContext {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout: defaultTargetConnectTimeout,
		}
		separator := strings.LastIndex(addr, ":")
		ips, err := p.dnsCache.Fetch(addr[:separator])
		if err != nil {
			return nil, err
		}
		var ip string
		for _, item := range ips {
			ip = item.String()
			if !strings.Contains(ip, ":") {
				break
			}
		}

		addr = ip + addr[separator:]

		return dialer.DialContext(ctx, network, addr)
	}
}

// 获取底层连接
func hijacker(rw http.ResponseWriter) (*ConnBuffer, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("http server不支持Hijacker")
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijacker错误: %s", err)
	}

	return NewConnBuffer(conn, buf), nil
}

// CopyHeader 浅拷贝Header
func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// CloneHeader 深拷贝Header
func CloneHeader(h http.Header, h2 http.Header) {
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
}

var hopHeaders = []string{
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	//"Te",
	//"Trailer",
	//"Transfer-Encoding",
}

type ConnBuffer struct {
	net.Conn
	buf *bufio.ReadWriter
}

func NewConnBuffer(conn net.Conn, buf *bufio.ReadWriter) *ConnBuffer {
	if buf == nil {
		buf = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	}
	return &ConnBuffer{
		Conn: conn,
		buf:  buf,
	}
}

func (cb *ConnBuffer) BufferReader() *bufio.Reader {
	return cb.buf.Reader
}

func (cb *ConnBuffer) Read(b []byte) (n int, err error) {
	return cb.buf.Read(b)
}

func (cb *ConnBuffer) Peek(n int) ([]byte, error) {
	return cb.buf.Peek(n)
}

func (cb *ConnBuffer) Write(p []byte) (n int, err error) {
	n, err = cb.buf.Write(p)
	if err != nil {
		return 0, err
	}

	return n, cb.buf.Flush()
}

func (cb *ConnBuffer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return cb.Conn, cb.buf, nil
}

func (cb *ConnBuffer) WriteHeader(_ int) {}

func (cb *ConnBuffer) Header() http.Header { return nil }
