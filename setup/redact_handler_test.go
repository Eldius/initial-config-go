package setup

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"maps"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/eldius/initial-config-go/configs"
	"github.com/stretchr/testify/assert"
)

func TestRedactValues_Handler(t *testing.T) {
	t.Run("given log keys using the same casing should redact then", func(t *testing.T) {
		var buff bytes.Buffer
		h, err := logHandler(
			configs.LogFormatJSON,
			configs.LogLevelDEBUG,
			&buff,
			"authentication",
			"permission",
		)
		assert.Nil(t, err)
		l := slog.New(h)

		l.With(
			slog.String("key", "value"),
		).Info("show message")
		l.With(
			slog.String("key", "value"),
			slog.String("permission", "hidden value"),
			slog.String("authentication", "another hidden value"),
			slog.String("request.headers.authentication", "another hidden value"),
		).Info("redacted message 0")
		l.With(
			slog.String("permission", "again"),
			slog.String("key", "value"),
		).Info("redacted message 1")

		l.With(
			"test_attr", map[string]any{
				"key": "value",
				"authentication": map[string]any{
					"map_key": "map_value",
				},
			},
		).Info("redacted message 2")

		for _, line := range bytes.Split(buff.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			err := json.Unmarshal(line, &m)
			assert.NoError(t, err)

			for k := range maps.Keys(m) {
				if strings.Contains(k, "permission") || k == "authentication" {
					assert.Equal(t, "***", m[k])
				}
				if k == "key" {
					assert.Equal(t, "value", m[k])
				}
			}
		}
	})

	t.Run("running the default slog tests", func(t *testing.T) {
		handler, buf := newTestRedactHandler(t, nil)
		results := func() []map[string]any {
			var ms []map[string]any
			for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
				if len(line) == 0 {
					continue
				}
				var m map[string]any
				if err := json.Unmarshal(line, &m); err != nil {
					t.Fatal(err)
				}

				m["msg"] = m["message"]

				ms = append(ms, m)
			}
			return ms
		}
		err := slogtest.TestHandler(handler, results)
		assert.NoError(t, err)
	})

	t.Run("given a non redacted key it should not be changed", func(t *testing.T) {
		handler, buf := newTestRedactHandler(t, []string{"redacted-key"})
		l := slog.New(handler)

		l.With(slog.String("my_key", "not redacted")).Info("firstTest")

		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var logEntryMap map[string]any

			err := json.Unmarshal(line, &logEntryMap)
			assert.NoError(t, err)

			assert.Equal(t, "not redacted", logEntryMap["my_key"])
			assert.Equal(t, "firstTest", logEntryMap["message"])
		}
	})

	t.Run("given a record with redacted key in a map value it must be changed", func(t *testing.T) {
		handler, buf := newTestRedactHandler(t, []string{"authentication"})
		l := slog.New(handler)

		l.With("request", map[string]any{
			"headers": map[string]any{
				"Content-Type":   "application/json",
				"Accept":         "application/json",
				"Authentication": "My Secret Authentication Key",
			},
		}).Info("firstTest")

		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var logEntryMap map[string]any

			err := json.Unmarshal(line, &logEntryMap)
			assert.NoError(t, err)

			assert.Equal(t, "***", getLogEntryAttrValue(t, logEntryMap, "request", "headers", "Authentication"))
		}
	})

	t.Run("given a record with redacted key in a struct value it must be changed", func(t *testing.T) {
		handler, buf := newTestRedactHandler(t, []string{"authentication"})
		l := slog.New(handler)

		l.With("request", map[string]any{
			"headers": testStruct{
				ContentType:    "application/json",
				Accept:         "application/json",
				Authentication: "My Secret Authentication Key",
			},
		}).Info("firstTest")

		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var logEntryMap map[string]any

			err := json.Unmarshal(line, &logEntryMap)
			assert.NoError(t, err)

			assert.Equal(t, "***", getLogEntryAttrValue(t, logEntryMap, "request", "headers", "Authentication"))
		}
	})
}

func TestRedact_Handle_Attributes(t *testing.T) {
	var buf bytes.Buffer
	handler := newRedactHandler(slog.NewJSONHandler(&buf, nil), []string{"password"})
	logger := slog.New(handler)

	logger.Info("login attempt", "user", "admin", "password", "secret123")

	var m map[string]any
	err := json.Unmarshal(buf.Bytes(), &m)
	assert.NoError(t, err)

	assert.Equal(t, "***", m["password"], "Password should be redacted in Handle")
}

func getLogEntryAttrValue(t *testing.T, m map[string]any, keys ...string) any {
	t.Helper()

	if len(keys) == 1 {
		return m[keys[0]]
	}

	currVal := m[keys[0]]
	if v, ok := currVal.(map[string]any); ok {
		return getLogEntryAttrValue(t, v, keys[1:]...)
	}
	return nil
}

func newTestRedactHandler(t *testing.T, redactedKeys []string) (slog.Handler, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer
	return newRedactHandler(
		slog.NewJSONHandler(
			&buf,
			&slog.HandlerOptions{
				ReplaceAttr: logAttrsReplacerFunc(),
				Level:       slog.LevelDebug,
			},
		),
		redactedKeys,
	), &buf

}

type testStruct struct {
	Authentication string `json:"Authentication"`
	ContentType    string `json:"Content-type"`
	Accept         string `json:"Accept"`
}
