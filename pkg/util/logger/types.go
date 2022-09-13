package logger

import (
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/rs/zerolog"
)

type Message struct {
	log    Logger
	Format string
	Fields map[string]interface{}
	Args   []interface{}
	Time   time.Time
	Level  zerolog.Level
}

// CallerHook implements zerolog.Hook interface.
type CallerHook struct{}

// Run adds additional context
func (h CallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level > zerolog.InfoLevel {
		if _, file, line, ok := runtime.Caller(4); ok {
			e.Str("file", fmt.Sprintf("%s:%d", path.Base(file), line))
		}
	}
}
