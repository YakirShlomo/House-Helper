package log

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// ZapAdapter adapts zap.Logger to Temporal's log.Logger interface
type ZapAdapter struct {
	zap *zap.Logger
}

// NewZapAdapter creates a new Temporal logger adapter for zap
func NewZapAdapter(zapLogger *zap.Logger) log.Logger {
	return &ZapAdapter{zap: zapLogger}
}

func (l *ZapAdapter) fields(keyvals []interface{}) []zap.Field {
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Any("fields", keyvals)}
	}
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		key := fmt.Sprint(keyvals[i])
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	return fields
}

func (l *ZapAdapter) Debug(msg string, keyvals ...interface{}) {
	l.zap.Debug(msg, l.fields(keyvals)...)
}

func (l *ZapAdapter) Info(msg string, keyvals ...interface{}) {
	l.zap.Info(msg, l.fields(keyvals)...)
}

func (l *ZapAdapter) Warn(msg string, keyvals ...interface{}) {
	l.zap.Warn(msg, l.fields(keyvals)...)
}

func (l *ZapAdapter) Error(msg string, keyvals ...interface{}) {
	l.zap.Error(msg, l.fields(keyvals)...)
}
