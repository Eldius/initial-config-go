package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

var (
	testTraceID = trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	testSpanID  = trace.SpanID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
)

func testSpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    testTraceID,
		SpanID:     testSpanID,
		TraceFlags: trace.FlagsSampled,
	})
}

func TestTracingHandler_Handle(t *testing.T) {
	t.Run("given a context with a valid span context it should add trace_id and span_id attributes", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		l := slog.New(handler)

		ctx := trace.ContextWithSpanContext(t.Context(), testSpanContext())
		l.InfoContext(ctx, "message with span")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		assert.Equal(t, testTraceID.String(), m[TraceIDKey])
		assert.Equal(t, testSpanID.String(), m[SpanIDKey])
	})

	t.Run("given a context without a span it should not add tracing attributes", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		l := slog.New(handler)

		l.InfoContext(t.Context(), "message without span")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		_, hasTraceID := m[TraceIDKey]
		_, hasSpanID := m[SpanIDKey]
		assert.False(t, hasTraceID)
		assert.False(t, hasSpanID)
	})

	t.Run("given an invalid span context it should not add tracing attributes", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		l := slog.New(handler)

		ctx := trace.ContextWithSpanContext(t.Context(), trace.SpanContext{})
		l.InfoContext(ctx, "message with invalid span")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		_, hasTraceID := m[TraceIDKey]
		_, hasSpanID := m[SpanIDKey]
		assert.False(t, hasTraceID)
		assert.False(t, hasSpanID)
	})

	t.Run("given a logger built with With it should keep tracing attributes", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		l := slog.New(handler).With("app", "test")

		ctx := trace.ContextWithSpanContext(t.Context(), testSpanContext())
		l.InfoContext(ctx, "message with attrs", "key", "value")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		assert.Equal(t, "test", m["app"])
		assert.Equal(t, testTraceID.String(), m[TraceIDKey])
		assert.Equal(t, testSpanID.String(), m[SpanIDKey])
	})

	t.Run("given a logger built with WithGroup tracing attributes should be added inside the group", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		l := slog.New(handler).WithGroup("group")

		ctx := trace.ContextWithSpanContext(t.Context(), testSpanContext())
		l.InfoContext(ctx, "message with group", "key", "value")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		group, ok := m["group"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, testTraceID.String(), group[TraceIDKey])
		assert.Equal(t, testSpanID.String(), group[SpanIDKey])
	})

	t.Run("running the default slog tests", func(t *testing.T) {
		var buf bytes.Buffer
		handler := NewTracingHandler(slog.NewJSONHandler(&buf, nil))
		results := func() []map[string]any {
			var ms []map[string]any
			for line := range bytes.SplitSeq(buf.Bytes(), []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				var m map[string]any
				if err := json.Unmarshal(line, &m); err != nil {
					t.Fatal(err)
				}
				ms = append(ms, m)
			}
			return ms
		}
		err := slogtest.TestHandler(handler, results)
		assert.NoError(t, err)
	})
}

func TestTracingHandler_WithNewLogger(t *testing.T) {
	t.Run("given a logger created via NewLogger it should include tracing ids from its context", func(t *testing.T) {
		var buf bytes.Buffer
		defaultLogger := slog.Default()
		slog.SetDefault(slog.New(NewTracingHandler(slog.NewJSONHandler(&buf, nil))))
		t.Cleanup(func() { slog.SetDefault(defaultLogger) })

		ctx := trace.ContextWithSpanContext(t.Context(), testSpanContext())
		NewLogger(ctx).Info("facade message")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		assert.Equal(t, testTraceID.String(), m[TraceIDKey])
		assert.Equal(t, testSpanID.String(), m[SpanIDKey])
	})

	t.Run("given a logger created from a context without span it should not include tracing ids", func(t *testing.T) {
		var buf bytes.Buffer
		defaultLogger := slog.Default()
		slog.SetDefault(slog.New(NewTracingHandler(slog.NewJSONHandler(&buf, nil))))
		t.Cleanup(func() { slog.SetDefault(defaultLogger) })

		NewLogger(context.Background()).Info("facade message without span")

		var m map[string]any
		err := json.Unmarshal(buf.Bytes(), &m)
		require.NoError(t, err)

		_, hasTraceID := m[TraceIDKey]
		_, hasSpanID := m[SpanIDKey]
		assert.False(t, hasTraceID)
		assert.False(t, hasSpanID)
	})
}
