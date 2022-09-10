/*
 Copyright (C) 2022 alsritter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package goproxy

import (
	"net/http"
	"net/url"
	"strings"
)

// Context 代理上下文
type Context struct {
	Req         *http.Request
	Data        map[interface{}]interface{}
	TunnelProxy bool
	abort       bool

	needMock bool
	failFast bool
	Resp     *http.Response
}

func (c *Context) IsFailFast() {
	c.failFast = true
}

func (c *Context) IsNeedMock() {
	c.needMock = true
}

func (c *Context) IsHTTPS() bool {
	return c.Req.URL.Scheme == "https"
}

var defaultPorts = map[string]string{
	"https": "443",
	"http":  "80",
	"":      "80",
}

func (c *Context) Addr() string {
	addr := c.Req.Host

	if !strings.Contains(c.Req.URL.Host, ":") {
		addr += ":" + defaultPorts[c.Req.URL.Scheme]
	}

	return addr
}

// Abort 中断执行
func (c *Context) Abort() {
	c.abort = true
}

// IsAborted 是否已中断执行
func (c *Context) IsAborted() bool {
	return c.abort
}

// Reset 重置
func (c *Context) Reset(req *http.Request) {
	c.Req = req
	c.Data = make(map[interface{}]interface{})
	c.abort = false
	c.TunnelProxy = false
}

type Delegate interface {
	// Connect 收到客户端连接
	Connect(ctx *Context, rw http.ResponseWriter)
	// Auth 代理身份认证
	Auth(ctx *Context, rw http.ResponseWriter)
	// BeforeRequest HTTP请求前 设置X-Forwarded-For, 修改Header、Body
	BeforeRequest(ctx *Context)
	// BeforeResponse 响应发送到客户端前, 修改Header、Body、Status Code
	BeforeResponse(ctx *Context, resp *http.Response, err error)
	// ParentProxy 上级代理
	ParentProxy(*http.Request) (*url.URL, error)
	// Finish 本次请求结束
	Finish(ctx *Context)
	// ErrorLog 记录错误信息
	ErrorLog(err error)
}

var _ Delegate = &DefaultDelegate{}

// DefaultDelegate 默认Handler什么也不做
type DefaultDelegate struct {
	Delegate
}

func (h *DefaultDelegate) Connect(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) Auth(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) BeforeRequest(ctx *Context) {}

func (h *DefaultDelegate) BeforeResponse(ctx *Context, resp *http.Response, err error) {}

func (h *DefaultDelegate) ParentProxy(req *http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment(req)
}

func (h *DefaultDelegate) Finish(ctx *Context) {}

func (h *DefaultDelegate) ErrorLog(err error) {}
