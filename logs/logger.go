package logs

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	key contextDataKey = "logger"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debug(format string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	WithError(err error) Logger
	WithExtraData(key string, value interface{}) Logger
}

type contextDataKey string

type logger struct {
	Logger
	ctx context.Context
	l   *slog.Logger
}

type KeyValueData struct {
	Key   string
	Value interface{}
}

func addDataToContext(ctx context.Context, field KeyValueData) context.Context {
	if values := ctx.Value(key); values != nil {
		if m, ok := values.(map[string]interface{}); ok {
			m[field.Key] = field.Value
			return context.WithValue(ctx, contextDataKey("logger"), m)
		}
	}
	return context.WithValue(ctx, contextDataKey("logger"), map[string]interface{}{field.Key: field.Value})
}

// NewLogger creates a new Logger instance
func NewLogger(ctx context.Context, fields ...KeyValueData) Logger {
	l := slog.Default()
	for _, field := range fields {
		l = l.With(field.Key, field.Value)
		ctx = addDataToContext(ctx, field)
	}

	return &logger{
		l:   l,
		ctx: ctx,
	}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.l.DebugContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Infof(format string, args ...interface{}) {
	l.l.InfoContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Warnf(format string, args ...interface{}) {
	l.l.WarnContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Errorf(format string, args ...interface{}) {
	l.l.ErrorContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l *logger) Debug(msg string) {
	l.l.DebugContext(l.ctx, msg)
}
func (l *logger) Infor(msg string) {
	l.l.InfoContext(l.ctx, msg)
}
func (l *logger) Warn(msg string) {
	l.l.WarnContext(l.ctx, msg)
}
func (l *logger) Error(msg string) {
	l.l.ErrorContext(l.ctx, msg)
}

func (l *logger) WithError(err error) Logger {
	return &logger{
		l: l.l.With("error", err),
	}
}

func (l *logger) WithExtraData(key string, value interface{}) Logger {
	return &logger{
		l: l.l.With(key, value),
	}
}
