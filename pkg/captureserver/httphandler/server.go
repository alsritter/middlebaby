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
	"net/http"
	"net/http/httptrace"

	"github.com/alsritter/middlebaby/pkg/util/goproxy"
	"github.com/alsritter/middlebaby/pkg/util/logger"
	"github.com/gorilla/handlers"
)

type Provider interface {
	GetServer() http.Handler
}

type captureServer struct {
	*goproxy.Proxy
	logger.Logger
}

func New(log logger.Logger) Provider {
	l := log.NewLogger("http-capture")
	return &captureServer{
		Logger: l,
		Proxy: goproxy.New(goproxy.WithDelegate(&delegateHandler{
			Logger: l,
		}),
			goproxy.WithDecryptHTTPS(&cache{}),
			goproxy.WithClientTrace(&httptrace.ClientTrace{
				DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {},
				GotConn: func(connInfo httptrace.GotConnInfo) {},
			}),
		),
	}
}

func (m *captureServer) GetServer() http.Handler {
	return handlers.CompressHandler(m)
}
