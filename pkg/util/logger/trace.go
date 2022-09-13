package logger

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

var (
	logBucket   sync.Map
	messagePool = sync.Pool{
		New: func() interface{} {
			return new(Message)
		},
	}
	taskQueue = make(chan []*Message, 10)
)

func init() {
	go func() {
		for logs := range taskQueue {
			for _, msg := range logs {
				printLogWithLevel(msg)
				messagePool.Put(msg)
			}
		}
	}()
}

func addJobs(logs []*Message) {
	taskQueue <- logs
}

func printLogWithLevel(msg *Message) {
	switch msg.Level {
	case zerolog.TraceLevel:
		msg.log.TraceWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	case zerolog.DebugLevel:
		msg.log.DebugWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	case zerolog.InfoLevel:
		msg.log.InfoWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	case zerolog.WarnLevel:
		msg.log.WarnWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	case zerolog.ErrorLevel:
		msg.log.ErrorWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	case zerolog.FatalLevel:
		msg.log.FatalWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	default:
		msg.log.InfoWithTime(msg.Time, msg.Fields, msg.Format, msg.Args...)
	}
}

func genTraceId() string {
	return uuid.New().String()
}

type TraceLogger interface {
	Logger
	// Begin begin build trace id.
	Begin() context.Context
	// output trace log.
	Done()
}

type basicTraceLogger struct {
	log Logger
	ctx context.Context
}

// Debug implements TraceLogger
func (b *basicTraceLogger) Debug(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.DebugLevel, fields, format, args...)
}

// DebugWithTime implements TraceLogger
func (b *basicTraceLogger) DebugWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.TraceLevel, fields, format, args...)
}

// Error implements TraceLogger
func (b *basicTraceLogger) Error(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.ErrorLevel, fields, format, args...)
}

// ErrorWithTime implements TraceLogger
func (b *basicTraceLogger) ErrorWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.ErrorLevel, fields, format, args...)
}

// Fatal implements TraceLogger
func (b *basicTraceLogger) Fatal(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.FatalLevel, fields, format, args...)
}

// FatalWithTime implements TraceLogger
func (b *basicTraceLogger) FatalWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.FatalLevel, fields, format, args...)
}

// GetCurrentLevel implements TraceLogger
func (b *basicTraceLogger) GetCurrentLevel() string {
	return b.GetCurrentLevel()
}

// Info implements TraceLogger
func (b *basicTraceLogger) Info(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.InfoLevel, fields, format, args...)
}

// InfoWithTime implements TraceLogger
func (b *basicTraceLogger) InfoWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.InfoLevel, fields, format, args...)
}

// NewLogger implements TraceLogger
func (b *basicTraceLogger) NewLogger(component string) Logger {
	return b.NewLogger(component)
}

// SetLogLevel implements TraceLogger
func (b *basicTraceLogger) SetLogLevel(verbosity string) {
	b.SetLogLevel(verbosity)
}

// Trace implements TraceLogger
func (b *basicTraceLogger) Trace(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.TraceLevel, fields, format, args...)
}

// TraceWithTime implements TraceLogger
func (b *basicTraceLogger) TraceWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.TraceLevel, fields, format, args...)
}

// Warn implements TraceLogger
func (b *basicTraceLogger) Warn(fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.WarnLevel, fields, format, args...)
}

// WarnWithTime implements TraceLogger
func (b *basicTraceLogger) WarnWithTime(_ time.Time, fields map[string]interface{}, format string, args ...interface{}) {
	b.storeMessage(zerolog.WarnLevel, fields, format, args...)
}

// WithContext implements TraceLogger
func (b *basicTraceLogger) WithContext(ctx context.Context) TraceLogger {
	return b
}

// Begin implements Logger
func (b *basicTraceLogger) Begin() context.Context {
	ctx := context.WithValue(b.ctx, "trace-id", genTraceId())
	return ctx
}

// Done implements Logger
func (b *basicTraceLogger) Done() {
	if val, ok := logBucket.Load(b.getTraceId()); ok {
		addJobs(val.([]*Message))
		logBucket.Delete(b.getTraceId())
	}
}

func (b *basicTraceLogger) storeMessage(level zerolog.Level, fields map[string]interface{}, format string, args ...interface{}) {
	var (
		msgs []*Message
		ob   = messagePool.Get().(*Message)
	)

	ob.Format = format
	ob.Args = args
	ob.Level = level
	ob.Time = time.Now()
	ob.log = b.log

	if val, ok := logBucket.Load(b.getTraceId()); ok {
		msgs = val.([]*Message)
		msgs = append(msgs, ob)
	} else {
		msgs = []*Message{ob}
	}
	logBucket.Store(b.getTraceId(), msgs)
}

func (b *basicTraceLogger) getTraceId() string {
	if b.ctx.Value("trace-id") == nil {
		return ""
	}
	return b.ctx.Value("trace-id").(string)
}
