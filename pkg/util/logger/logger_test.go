package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		component string
		level     string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"测试使用",
			args{
				component: "redis-server",
				level:     "debug",
			},
		},
		{
			"Info 使用",
			args{
				component: "mysql-server",
				level:     "info",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlog := NewDefault(tt.args.component)
			tlog.SetLogLevel(tt.args.level)

			tlog.Debug(nil, "这是 Debug 消息")

			tlog.Info(map[string]interface{}{
				"nums":  199,
				"hello": "world",
			}, "这是Info消息")

			tlog.Warn(nil, "这是 Warn 消息")

			tlog.Error(nil, "这是 Error 消息")
			// tlog.Fatal(nil, "这是 Fatal")
		})
	}
}
