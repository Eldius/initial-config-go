package logs

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDKey is the log attribute key used for the OpenTelemetry trace ID.
	TraceIDKey = "trace_id"
	// SpanIDKey is the log attribute key used for the OpenTelemetry span ID.
	SpanIDKey = "span_id"
)

type tracingHandler struct {
	h slog.Handler
}

// NewTracingHandler wraps h so that, when the context of a log record carries
// a valid OpenTelemetry span context (i.e. there is an active span), the
// record is enriched with trace_id and span_id attributes before being
// handled. Records logged without an active span are left untouched.
func NewTracingHandler(h slog.Handler) slog.Handler {
	return &tracingHandler{h: h}
}

func (t *tracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return t.h.Enabled(ctx, level)
}

func (t *tracingHandler) Handle(ctx context.Context, record slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return t.h.Handle(ctx, record)
	}

	newRecord := record.Clone()
	newRecord.AddAttrs(
		slog.String(TraceIDKey, sc.TraceID().String()),
		slog.String(SpanIDKey, sc.SpanID().String()),
	)
	return t.h.Handle(ctx, newRecord)
}

func (t *tracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &tracingHandler{h: t.h.WithAttrs(attrs)}
}

func (t *tracingHandler) WithGroup(name string) slog.Handler {
	return &tracingHandler{h: t.h.WithGroup(name)}
}
