package logs

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/eldius/initial-config-go/configs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogHandler(t *testing.T) {
	t.Run("json format produces valid JSON", func(t *testing.T) {
		var buf bytes.Buffer
		h, err := LogHandler(configs.LogFormatJSON, configs.LogLevelDEBUG, &buf)
		require.NoError(t, err)
		l := slog.New(h)
		l.Info("test")
		var m map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &m))
		assert.Equal(t, "test", m["message"].(string))
	})

	t.Run("text format outputs text", func(t *testing.T) {
		var buf bytes.Buffer
		h, err := LogHandler(configs.LogFormatText, configs.LogLevelDEBUG, &buf)
		require.NoError(t, err)
		l := slog.New(h)
		l.Info("test")
		assert.Contains(t, buf.String(), "test")
	})

	t.Run("with redacted keys replaces their values", func(t *testing.T) {
		var buf bytes.Buffer
		h, err := LogHandler(configs.LogFormatJSON, configs.LogLevelDEBUG, &buf, "secret")
		require.NoError(t, err)
		l := slog.New(h)
		l.With(slog.String("secret", "hidden-value")).Info("test")
		var m map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &m))
		assert.Equal(t, "***", m["secret"])
	})

	t.Run("without redacted keys passes values through", func(t *testing.T) {
		var buf bytes.Buffer
		h, err := LogHandler(configs.LogFormatJSON, configs.LogLevelDEBUG, &buf)
		require.NoError(t, err)
		l := slog.New(h)
		l.With(slog.String("visible", "value")).Info("test")
		var m map[string]any
		assert.NoError(t, json.Unmarshal(buf.Bytes(), &m))
		assert.Equal(t, "value", m["visible"])
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo},
		{"", slog.LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, parseLogLevel(tt.input))
		})
	}
}

func TestGetWriter(t *testing.T) {
	t.Run("stdout only", func(t *testing.T) {
		w, err := GetWriter("", true)
		require.NoError(t, err)
		assert.Same(t, os.Stdout, w)
	})

	t.Run("file only", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.log")
		w, err := GetWriter(path, false)
		require.NoError(t, err)
		_, ok := w.(*os.File)
		assert.True(t, ok, "expected *os.File")
		_ = CloseLogFiles()
		_, err = os.Stat(path)
		assert.NoError(t, err, "file should exist")
	})

	t.Run("stdout and file returns MultiWriter", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.log")
		w, err := GetWriter(path, true)
		require.NoError(t, err)
		assert.NotNil(t, w)
		_ = CloseLogFiles()
	})
}

func TestCloseLogFiles(t *testing.T) {
	t.Run("with no files should not error", func(t *testing.T) {
		err := CloseLogFiles()
		assert.NoError(t, err)
	})

	t.Run("after open should close and clear", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.log")
		_, err := GetWriter(path, false)
		require.NoError(t, err)
		err = CloseLogFiles()
		assert.NoError(t, err)
		err = CloseLogFiles()
		assert.NoError(t, err, "double close should be safe")
	})
}

func TestRedactHandler_Simplified(t *testing.T) {
	t.Run("direct NewRedactHandler redacts keys", func(t *testing.T) {
		var buf bytes.Buffer
		h := NewRedactHandler(slog.NewJSONHandler(&buf, nil), []string{"password"})
		l := slog.New(h)
		l.Info("login", "password", "secret", "user", "admin")
		var m map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
		assert.Equal(t, "***", m["password"])
		assert.Equal(t, "admin", m["user"])
	})

	t.Run("NewRedactHandler passes non-redacted keys through", func(t *testing.T) {
		var buf bytes.Buffer
		h := NewRedactHandler(slog.NewJSONHandler(&buf, nil), []string{"password"})
		l := slog.New(h)
		l.Info("test", "user", "admin", "token", "abc123")
		var m map[string]any
		require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
		assert.Equal(t, "admin", m["user"])
		assert.Equal(t, "abc123", m["token"])
	})

	t.Run("NewRedactHandler handles empty redact list", func(t *testing.T) {
		var buf bytes.Buffer
		h := NewRedactHandler(slog.NewJSONHandler(&buf, nil), nil)
		l := slog.New(h)
		l.Info("test", "key", "value")
		assert.Contains(t, buf.String(), "value")
	})
}
