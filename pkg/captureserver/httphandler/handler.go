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

package httphandler

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"github.com/alsritter/middlebaby/pkg/messagepush"
	"github.com/alsritter/middlebaby/pkg/util/logger"

	"github.com/alsritter/middlebaby/pkg/util/goproxy"
)

type delegateHandler struct {
	logger.Logger
}

// Connect check the request type.
func (e *delegateHandler) Connect(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) Auth(ctx *goproxy.Context, rw http.ResponseWriter) {}

func (e *delegateHandler) BeforeRequest(ctx *goproxy.Context) {
	body, err := ioutil.ReadAll(ctx.Req.Body)
	ctx.Req.Body = ioutil.NopCloser(bytes.NewReader(body))
	if err != nil {
		e.WithContext(ctx.Req.Context()).Error(nil, "read request body error: %v", err)
		ctx.Abort()
		return
	}

	messagepush.SendMessage(fmt.Sprintf("%+v", ctx.Req))
	e.WithContext(ctx.Req.Context()).Debug(nil, "capture [%s] request [%+v]", ctx.Req.URL, ctx.Req)
}

func (e *delegateHandler) BeforeResponse(ctx *goproxy.Context, resp *http.Response, err error) {

	messagepush.SendMessage(fmt.Sprintf("%+v", resp))
	e.WithContext(ctx.Req.Context()).Debug(nil, "capture [%v] response [%+v]", ctx.Req.URL, resp)
}

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
