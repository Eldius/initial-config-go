package telemetry

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// GetTracer returns a tracer instance.
func GetTracer(tracerName string, opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(tracerName, opts...)
}

// GetCurrentSpan returns a span instance.
func GetCurrentSpan(ctx context.Context, tracerName string, opts ...trace.TracerOption) trace.Span {
	return trace.SpanFromContext(ctx)
}

// NewSpan creates and starts a new trace span with the provided context, trace name, and optional span start options.
func NewSpan(ctx context.Context, traceName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, traceName, opts...)
}

type TracingIDs struct {
	TraceID string
	SpanID  string
}

func GetSpanDataFromContext(ctx context.Context) TracingIDs {
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	spanID := trace.SpanFromContext(ctx).SpanContext().SpanID().String()
	return TracingIDs{
		TraceID: traceID,
		SpanID:  spanID,
	}
}
