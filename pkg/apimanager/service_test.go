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

package apimanager

import (
	"testing"

	"github.com/alsritter/middlebaby/pkg/interact"
	"github.com/alsritter/middlebaby/pkg/util/logger"
)

func TestManager_match(t *testing.T) {
	type args struct {
		req    *interact.Request
		target *interact.Request
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "基本测试",
			args: args{
				req: &interact.Request{
					Method: "POST",
					Host:   "localhost",
					Path:   "/get/hello/world",
					Header: map[string][]string{
						"Accept-Encoding:": {"gzip, deflate"},
					},
					Params: map[string]string{},
					Body:   nil,
				},
				target: &interact.Request{
					Method: "POST",
					Host:   "localhost",
					Path:   "/get/hello/world",
					Header: map[string][]string{
						"Accept-Encoding:": {"gzip, deflate"},
					},
					Params: map[string]string{},
					Body:   nil,
				},
			},
			want: true,
		},
		{
			name: "Host 断言失败",
			args: args{
				req: &interact.Request{
					Method: "GET",
					Host:   "127.0.0.1",
				},
				target: &interact.Request{
					Method: "GET",
					Host:   "localhost",
				},
			},
			want: false,
		},
		{
			name: "Path 断言失败",
			args: args{
				req: &interact.Request{
					Method: "GET",
					Host:   "localhost",
					Path:   "/path",
				},
				target: &interact.Request{
					Method: "GET",
					Host:   "localhost",
					Path:   "/path2",
				},
			},
			want: false,
		},
		{
			name: "Header 断言失败",
			args: args{
				req: &interact.Request{
					Method: "GET",
					Host:   "localhost",
					Path:   "/path",
					Header: map[string][]string{
						"Accept-Encoding:": {"gzip, deflate"},
					},
				},
				target: &interact.Request{
					Method: "GET",
					Host:   "localhost",
					Path:   "/path",
					Header: map[string][]string{
						"Accept-Encoding:": {"gzip, deflate"},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{Logger: logger.NewDefault("test")}
			if got := m.match(tt.args.req, tt.args.target); got != tt.want {
				t.Errorf("Manager.match() = %v, want %v", got, tt.want)
			}
		})
	}
}
