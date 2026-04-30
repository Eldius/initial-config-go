package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerSource(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{AddSource: true})
	l := &logger{
		ctx:    context.Background(),
		logger: slog.New(handler),
	}

	l.Info("test message")

	var data map[string]any
	err := json.Unmarshal(buf.Bytes(), &data)
	assert.NoError(t, err)

	source, ok := data["source"].(map[string]any)
	assert.True(t, ok, "source should be present in log output")
	assert.Contains(t, source, "function", "source should contain function")
	assert.Contains(t, source, "file", "source should contain file")
	assert.Contains(t, source, "line", "source should contain line")
	assert.Contains(t, source["file"], "logger_test.go", "source file should be logger_test.go")
	assert.Contains(t, source["function"], "TestLoggerSource", "source function should contain TestLoggerSource")
}

func TestLoggerSource_f(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{AddSource: true})
	l := &logger{
		ctx:    context.Background(),
		logger: slog.New(handler),
	}

	l.Infof("test message %s", "formatted")

	var data map[string]any
	err := json.Unmarshal(buf.Bytes(), &data)
	assert.NoError(t, err)

	source, ok := data["source"].(map[string]any)
	assert.True(t, ok, "source should be present in log output")
	assert.Contains(t, source, "function", "source should contain function")
	assert.Contains(t, source, "file", "source should contain file")
	assert.Contains(t, source, "line", "source should contain line")
	assert.Contains(t, source["file"], "logger_test.go", "source file should be logger_test.go")
	assert.Contains(t, source["function"], "TestLoggerSource_f", "source function should contain TestLoggerSource_f")
}

func TestLoggerSource_WithError(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{AddSource: true})
	l := &logger{
		ctx:    context.Background(),
		logger: slog.New(handler),
	}

	l.WithError(fmt.Errorf("test error")).Info("test message")

	var data map[string]any
	err := json.Unmarshal(buf.Bytes(), &data)
	assert.NoError(t, err)

	source, ok := data["source"].(map[string]any)
	assert.True(t, ok, "source should be present in log output")
	assert.Contains(t, source, "function", "source should contain function")
	assert.Contains(t, source, "file", "source should contain file")
	assert.Contains(t, source, "line", "source should contain line")
	assert.Contains(t, source["file"], "logger_test.go", "source file should be logger_test.go")
	assert.Contains(t, source["function"], "TestLoggerSource_WithError", "source function should contain TestLoggerSource_WithError")
	assert.Equal(t, "test error", data["error"], "error message should be present")
}
