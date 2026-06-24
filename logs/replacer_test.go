package logs

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogAttrsReplacerFunc(t *testing.T) {
	replacer := LogAttrsReplacerFunc()

	t.Run("known keys are preserved", func(t *testing.T) {
		keys := []string{"host", "service.name", "level", "message", "time", "error", "source", "function", "file", "line"}
		for _, k := range keys {
			attr := replacer(nil, slog.String(k, "val"))
			assert.Equal(t, k, attr.Key, "known key %q should be preserved", k)
		}
	})

	t.Run("msg is renamed to message", func(t *testing.T) {
		attr := replacer(nil, slog.String("msg", "hello"))
		assert.Equal(t, "message", attr.Key)
		assert.Equal(t, "hello", attr.Value.String())
	})

	t.Run("request prefixed keys pass through", func(t *testing.T) {
		attr := replacer(nil, slog.String("request_id", "abc"))
		assert.Equal(t, "request_id", attr.Key)
	})

	t.Run("response prefixed keys pass through", func(t *testing.T) {
		attr := replacer(nil, slog.String("response_status", "200"))
		assert.Equal(t, "response_status", attr.Key)
	})

	t.Run("service prefixed keys pass through", func(t *testing.T) {
		attr := replacer(nil, slog.String("service_version", "1.0"))
		assert.Equal(t, "service_version", attr.Key)
	})

	t.Run("unknown keys pass through unchanged", func(t *testing.T) {
		attr := replacer(nil, slog.String("custom_key", "custom_value"))
		assert.Equal(t, "custom_key", attr.Key)
		assert.Equal(t, "custom_value", attr.Value.String())
	})
}
