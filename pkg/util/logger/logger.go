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

package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config defines the config structure
type Config struct {
	Pretty bool   `yaml:"prefix"`
	Level  string `yaml:"level"`
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Pretty: true,
		Level:  "debug",
	}
}

func (c *Config) Validate() error {
	return nil
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.Level, prefix+"log.level", c.Level, "log level(debug, info, warn, error, fatal)")
	f.BoolVar(&c.Pretty, prefix+"log.pretty", c.Pretty, "log in a pretty format")
}

// Logger defines the basic log library implementation
type Logger interface {
	// Trace print a message with trace level.
	Trace(fields map[string]interface{}, format string, args ...interface{})
	// Debug print a message with debug level.
	Debug(fields map[string]interface{}, format string, args ...interface{})
	// Info print a message with info level.
	Info(fields map[string]interface{}, format string, args ...interface{})
	// Warn print a message with warn level.
	Warn(fields map[string]interface{}, format string, args ...interface{})
	// Error print a message with error level.
	Error(fields map[string]interface{}, format string, args ...interface{})
	// Fatal print a message with fatal level.
	Fatal(fields map[string]interface{}, format string, args ...interface{})

	// Trace print a message with trace level.
	TraceWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})
	// Debug print a message with debug level.
	DebugWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})
	// Info print a message with info level.
	InfoWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})
	// Warn print a message with warn level.
	WarnWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})
	// Error print a message with error level.
	ErrorWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})
	// Fatal print a message with fatal level.
	FatalWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{})

	// WithContext trade log.
	WithContext(ctx context.Context) TraceLogger

	// NewLogger is used to derive a new child Logger
	NewLogger(component string) Logger
	// SetLogLevel is used to set log level
	SetLogLevel(verbosity string)

	GetCurrentLevel() string
}

// New is used to init service
func New(cfg *Config, component string) (Logger, error) {
	if cfg == nil {
		cfg = NewConfig()
	}
	service := &BasicLogger{
		cfg:       cfg,
		component: component,
	}
	service.setup()
	return service, nil
}

// NewDefault is used to initialize a simple Logger
func NewDefault(component string) Logger {
	logger, err := New(NewConfig(), component)
	if err != nil {
		panic(err)
	}
	return logger
}

// BasicLogger simply implements Logger
type BasicLogger struct {
	cfg *Config

	component string
	logger    zerolog.Logger
}

// WithContext implements Logger
func (b *BasicLogger) WithContext(ctx context.Context) TraceLogger {
	return &basicTraceLogger{
		log: b,
		ctx: ctx,
	}
}

// NewLogger is used to derive a new child Logger
func (b *BasicLogger) NewLogger(component string) Logger {
	name := strings.Join([]string{b.component, component}, ".")
	logger, err := New(b.cfg, name)
	if err != nil {
		b.Warn(map[string]interface{}{
			"name": name,
		}, "failed to extend logger: %s", err)
		return b
	}
	return logger
}

func (b *BasicLogger) setup() {
	b.logger = log.With().Str("comp", b.component).Logger().Hook(CallerHook{})
	if b.cfg != nil {
		if b.cfg.Pretty {
			b.logger = b.logger.Output(zerolog.ConsoleWriter{
				Out: os.Stdout,
			})
		}
		b.SetLogLevel(b.cfg.Level)
	}
}

// Trace Log print a message with debug level.
func (b *BasicLogger) Trace(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Trace().Fields(fields).Msgf(format, args...)
}

// TraceWithTime implements Logger
func (b *BasicLogger) TraceWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Trace().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// Debug Log print a message with debug level.
func (b *BasicLogger) Debug(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Debug().Fields(fields).Msgf(format, args...)
}

// DebugWithTime implements Logger
func (b *BasicLogger) DebugWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Debug().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// Info Log print a message with info level.
func (b *BasicLogger) Info(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Info().Fields(fields).Msgf(format, args...)
}

// InfoWithTime implements Logger
func (b *BasicLogger) InfoWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Info().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// Warn Log print a message with warn level.
func (b *BasicLogger) Warn(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Warn().Fields(fields).Msgf(format, args...)
}

// WarnWithTime implements Logger
func (b *BasicLogger) WarnWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Warn().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// Error Log print a message with error level.
func (b *BasicLogger) Error(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Error().Fields(fields).Msgf(format, args...)
}

// ErrorWithTime implements Logger
func (b *BasicLogger) ErrorWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Error().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// Fatal Log print a message with fatal level.
func (b *BasicLogger) Fatal(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Fatal().Fields(fields).Msgf(format, args...)
}

// FatalWithTime implements Logger
func (b *BasicLogger) FatalWithTime(logTime time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Fatal().Time("log-time", logTime).Fields(fields).Msgf(format, args...)
}

// SetLogLevel is used to set log level
func (b *BasicLogger) SetLogLevel(verbosity string) {
	var l zerolog.Logger
	switch verbosity {
	case "trace":
		l = b.logger.Level(zerolog.TraceLevel)
	case "debug":
		l = b.logger.Level(zerolog.DebugLevel)
	case "info":
		l = b.logger.Level(zerolog.InfoLevel)
	case "warn":
		l = b.logger.Level(zerolog.WarnLevel)
	case "error":
		l = b.logger.Level(zerolog.ErrorLevel)
	case "fatal":
		l = b.logger.Level(zerolog.FatalLevel)
	default:
		l = b.logger.Level(zerolog.InfoLevel)
	}
	b.logger = l
}

func (b *BasicLogger) GetCurrentLevel() string {
	return b.cfg.Level
}
