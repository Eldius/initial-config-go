package setup

import (
	"context"
	"log/slog"
	"slices"
	"strings"
)

// redactHandler is a custom handler that redacts specific attributes.
type redactHandler struct {
	slog.Handler
	handler    slog.Handler
	redactKeys []string
}

func newRedactHandler(handler slog.Handler, redactKeys []string) *redactHandler {
	return &redactHandler{
		handler:    handler,
		redactKeys: redactKeys,
	}
}

// Handle method processes the log record and redacts the specified attribute.
func (rh *redactHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Attrs(func(attr slog.Attr) bool {
		var redact bool
		for _, key := range strings.Split(attr.Key, ".") {
			redact = slices.Contains(rh.redactKeys, key)
		}
		if redact {
			attr.Value = slog.StringValue("***")
		}

		return true
	})
	return rh.handler.Handle(ctx, record)
}

func (rh *redactHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return rh.handler.Enabled(ctx, level)
}

func (rh *redactHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return rh.handler.WithAttrs(attrs)
}

func (rh *redactHandler) WithGroup(name string) slog.Handler {
	return rh.handler.WithGroup(name)
}
