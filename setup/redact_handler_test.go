package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eldius/initial-config-go/configs"
	"github.com/stretchr/testify/assert"
	"log"
	"log/slog"
	"maps"
	"reflect"
	"strings"
	"testing"
	"testing/slogtest"
)

func TestRedactValues_Handler(t *testing.T) {
	t.Run("given log keys using the same casing should redact then", func(t *testing.T) {
		var buff bytes.Buffer
		h, err := logHandler(
			"my-app",
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
			t.Logf(" ==> raw: %s", line)
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			err := json.Unmarshal(line, &m)
			assert.NoError(t, err)
			t.Logf("     parsed: %+v", m)

			for k := range maps.Keys(m) {
				t.Logf("       key: %s", k)
				if strings.Contains(k, "permission") || strings.Contains(k, "authentication") {
					t.Logf("       m[k]: %s", m[k])
					assert.Equal(t, "***", m[k], fmt.Sprintf("actual value for key '%s': %s => expected value: ***", k, m[k]))
				}
				if strings.Contains(k, "key") {
					t.Logf("       m[k]: %s", m[k])
					assert.Equal(t, "value", m[k], fmt.Sprintf("actual value for key '%s': %s => expected value: value", k, m[k]))
				}
			}
		}
	})

	t.Run("running the default slog tests", func(t *testing.T) {
		handler, buf := newTestRedactedHandler(t, nil)
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
				//logMapKeys := maps.Keys(m)
				//assert.NotContains(t, logMapKeys, "msg")

				m["msg"] = m["message"]

				ms = append(ms, m)
			}
			return ms
		}
		err := slogtest.TestHandler(handler, results)
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("given a non redacted key it should not be changed", func(t *testing.T) {
		handler, buf := newTestRedactedHandler(t, []string{"redacted-key"})
		l := slog.New(handler)

		l.With(slog.String("my_key", "not redacted")).Info("firstTest")

		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var logEntryMap map[string]any

			err := json.Unmarshal(line, &logEntryMap)
			assert.NoError(t, err)

			t.Logf("logRecord: %s", line)
			t.Logf("logEntryMap: %+v", logEntryMap)

			assert.Equal(t, "not redacted", logEntryMap["my_key"])

			assert.Equal(t, "firstTest", logEntryMap["message"])
		}
	})

	t.Run("given a record with redacted key in a map value it must be changed", func(t *testing.T) {
		handler, buf := newTestRedactedHandler(t, []string{"authentication"})
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

			t.Logf("logRecord: %s", line)
			t.Logf("logEntryMap: %+v", logEntryMap)

			assert.Equal(t, "***", getLogEntryAttrValue(t, logEntryMap, "request", "headers", "Authentication"))

			assert.Equal(t, "firstTest", logEntryMap["message"])
		}
	})

	t.Run("given a record with redacted key in a struct value it must be changed", func(t *testing.T) {
		handler, buf := newTestRedactedHandler(t, []string{"authentication"})
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

			t.Logf("logRecord: %s", line)
			t.Logf("logEntryMap: %+v", logEntryMap)

			assert.Equal(t, "***", getLogEntryAttrValue(t, logEntryMap, "request", "headers", "Authentication"))

			assert.Equal(t, "firstTest", logEntryMap["message"])
		}
	})
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

func newTestRedactedHandler(t *testing.T, redactedKeys []string) (slog.Handler, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer
	return newRedactHandler(
		slog.NewJSONHandler(
			&buf,
			&slog.HandlerOptions{ReplaceAttr: logAttrsReplacerFunc()},
		),
		redactedKeys,
	), &buf

}

func TestZeroValue(t *testing.T) {
	t.Run("given a simple string must return ***", func(t *testing.T) {
		val := reflect.ValueOf("My Secret String")
		newVal := zeroValue(val)

		assert.NotNilf(t, newVal, "zeroValue should not be nil")
		assert.Equal(t, "***", newVal.Elem().String())

		t.Logf("%#v", val)
	})

	t.Run("given an int64 must return ***", func(t *testing.T) {
		intVal := int64(123)
		val := reflect.ValueOf(intVal)
		newVal := zeroValue(val)

		assert.NotNilf(t, newVal, "zeroValue should not be nil")
		assert.Equal(t, int64(-1), newVal.Elem().Int())

		t.Logf("%#v", val)
	})

	t.Run("given an int must return ***", func(t *testing.T) {
		intVal := int(123)
		val := reflect.ValueOf(intVal)
		newVal := zeroValue(val)

		assert.NotNilf(t, newVal, "zeroValue should not be nil")
		assert.Equal(t, int64(-1), newVal.Elem().Int())

		t.Logf("%#v", val)
	})
}

func TestParseStruct(t *testing.T) {
	t.Run("given a simple struct must return authentication attribute", func(t *testing.T) {
		in := testStruct{
			Authentication: "Secret value",
			ContentType:    "application/json",
			Accept:         "application/json",
		}
		out := parseStruct(in, []string{"authentication"})
		assert.Equal(t, reflect.TypeOf(in), reflect.TypeOf(out))
		assert.Equal(t, "***", out.Authentication)
		assert.Equal(t, in.ContentType, out.ContentType)
		assert.Equal(t, in.Accept, out.Accept)
	})
}

type testStruct struct {
	Authentication string `json:"Authentication"`
	ContentType    string `json:"Content-type"`
	Accept         string `json:"Accept"`
}
