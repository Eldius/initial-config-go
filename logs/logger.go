package logs

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"runtime"
	"time"
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
	WithExtraDataMap(data map[string]any) Logger
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
	return &logger{
		logger: l,
		ctx:    ctx,
	}
}

func (l *logger) log(level slog.Level, msg string) {
	if !l.logger.Enabled(l.ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, l.log, l.Info]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	_ = l.logger.Handler().Handle(l.ctx, r)
}

func (l *logger) Debugf(format string, args ...any) {
	l.log(slog.LevelDebug, fmt.Sprintf(format, args...))
}
func (l *logger) Infof(format string, args ...any) {
	l.log(slog.LevelInfo, fmt.Sprintf(format, args...))
}
func (l *logger) Warnf(format string, args ...any) {
	l.log(slog.LevelWarn, fmt.Sprintf(format, args...))
}
func (l *logger) Errorf(format string, args ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, args...))
}

func (l *logger) Debug(msg string) {
	l.log(slog.LevelDebug, msg)
}
func (l *logger) Info(msg string) {
	l.log(slog.LevelInfo, msg)
}
func (l *logger) Warn(msg string) {
	l.log(slog.LevelWarn, msg)
}
func (l *logger) Error(msg string) {
	l.log(slog.LevelError, msg)
}

func (l *logger) WithError(err error) Logger {
	return &logger{
		ctx:    l.ctx,
		logger: l.logger.With("error", err),
	}
}

func (l *logger) WithExtraData(key string, value any) Logger {
	return &logger{
		ctx:    l.ctx,
		logger: l.logger.With(key, value),
	}
}

func (l *logger) WithExtraDataMap(data map[string]any) Logger {
	log := l.logger
	for k, v := range data {
		log = log.With(k, v)
	}
	return &logger{
		ctx:    l.ctx,
		logger: log,
	}
}
