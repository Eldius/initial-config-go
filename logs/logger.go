package logs

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
)

const (
	key contextDataKey = "logger"
)

var (
	_ Logger = (*logger)(nil)
)

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Debug(format string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	WithError(err error) Logger
	WithExtraData(key string, value any) Logger
}

type contextDataKey string

type logger struct {
	ctx    context.Context
	logger *slog.Logger
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
	//trace := telemetry.GetSpanDataFromContext(ctx)
	//l = l.With(slog.String("trace_id", trace.TraceID), slog.String("span_id", trace.SpanID))
	return &logger{
		logger: l,
		ctx:    ctx,
	}
}

func (l *logger) Debugf(format string, args ...any) {
	l.logger.DebugContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Infof(format string, args ...any) {
	l.logger.InfoContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Warnf(format string, args ...any) {
	l.logger.WarnContext(l.ctx, fmt.Sprintf(format, args...))
}
func (l *logger) Errorf(format string, args ...any) {
	l.logger.ErrorContext(l.ctx, fmt.Sprintf(format, args...))
}

func (l *logger) Debug(msg string) {
	l.logger.DebugContext(l.ctx, msg)
}
func (l *logger) Info(msg string) {
	l.logger.InfoContext(l.ctx, msg)
}
func (l *logger) Warn(msg string) {
	l.logger.WarnContext(l.ctx, msg)
}
func (l *logger) Error(msg string) {
	l.logger.ErrorContext(l.ctx, msg)
}

func (l *logger) WithError(err error) Logger {
	return &logger{
		logger: l.logger.With("error", err),
	}
}

func (l *logger) WithExtraData(key string, value any) Logger {
	return &logger{
		logger: l.logger.With(key, value),
	}
}

func (l *logger) WithExtraDataMap(data map[string]any) Logger {
	log := l.logger
	for k, v := range data {
		log = log.With(k, v)
	}
	return &logger{
		logger: log,
	}
}
