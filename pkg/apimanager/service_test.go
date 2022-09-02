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
					Headers: map[string]interface{}{
						"Accept-Encoding:": "gzip, deflate",
					},
					Params: map[string]string{},
					Body:   nil,
				},
				target: &interact.Request{
					Method: "POST",
					Host:   "localhost",
					Path:   "/get/hello/world",
					Headers: map[string]interface{}{
						"Accept-Encoding:": "gzip, deflate",
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
					Headers: map[string]interface{}{
						"Accept-Encoding:": "gzip, deflate",
					},
				},
				target: &interact.Request{
					Method: "GET",
					Host:   "localhost",
					Path:   "/path",
					Headers: map[string]interface{}{
						"Accept-Encoding:": "text/plain",
					},
				},
			},
			want: false,
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
