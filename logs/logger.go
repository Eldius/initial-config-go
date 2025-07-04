package logs

import (
	"context"
	"fmt"
	"github.com/eldius/initial-config-go/telemetry"
	"log/slog"
	"maps"
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

type KeyValueData map[string]any

func addDataToContext(ctx context.Context, data ...KeyValueData) context.Context {
	if values := ctx.Value(key); values != nil {
		if m, ok := values.(KeyValueData); ok {
			for _, d := range data {
				maps.Copy(m, d)
			}
			return context.WithValue(ctx, contextDataKey("logger"), m)
		}
	}
	return context.WithValue(ctx, contextDataKey("logger"), data)
}

// NewLogger creates a new Logger instance
func NewLogger(ctx context.Context, fields ...KeyValueData) Logger {
	l := slog.Default()
	ctx = addDataToContext(ctx, fields...)
	for _, d := range fields {
		for k, v := range d {
			l = l.With(k, v)
		}
	}
	trace := telemetry.GetSpanDataFromContext(ctx)
	l = l.With(slog.String("trace_id", trace.TraceID), slog.String("span_id", trace.SpanID))
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
func (l *logger) Info(msg string) {
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

func (l *logger) WithExtraDataMap(data map[string]interface{}) Logger {
	log := l.l
	for k, v := range data {
		log = log.With(k, v)
	}
	return &logger{
		l: log,
	}
}
