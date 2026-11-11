package configs

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestMergeConfigMaps(t *testing.T) {
	t.Run("should merge two config maps", func(t *testing.T) {
		a := map[string]any{"a": "b"}
		b := map[string]any{"c": "d"}
		expected := map[string]any{"a": "b", "c": "d"}

		result := MergeConfigMaps(a, b)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("first element should have priority for same key", func(t *testing.T) {
		a := map[string]any{"a": "b"}
		b := map[string]any{"a": "c"}
		expected := map[string]any{"a": "b"}

		result := MergeConfigMaps(a, b)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})
}

func TestGetTelemetryHeaders(t *testing.T) {
	t.Run("returns empty map when not set", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		headers := GetTelemetryHeaders()
		if len(headers) != 0 {
			t.Errorf("expected empty map, got %v", headers)
		}
	})

	t.Run("returns map from config file value", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		expected := map[string]string{
			"Authorization": "Bearer token",
			"X-Api-Key":     "secret",
		}
		viper.Set(TelemetryHeadersKey, expected)

		headers := GetTelemetryHeaders()
		if !reflect.DeepEqual(headers, expected) {
			t.Errorf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("returns map from YAML config", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		viper.SetConfigType("yaml")
		yamlCfg := `
telemetry:
  enabled: true
  headers:
    authorization: Bearer token
    x-api-key: secret
`
		if err := viper.ReadConfig(strings.NewReader(yamlCfg)); err != nil {
			t.Fatalf("failed to read yaml config: %v", err)
		}

		expected := map[string]string{
			"authorization": "Bearer token",
			"x-api-key":     "secret",
		}
		headers := GetTelemetryHeaders()
		if !reflect.DeepEqual(headers, expected) {
			t.Errorf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("parses comma-separated key=value pairs from env string", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		viper.Set(TelemetryHeadersKey, "Authorization=Bearer token, X-Api-Key=secret")

		expected := map[string]string{
			"Authorization": "Bearer token",
			"X-Api-Key":     "secret",
		}
		headers := GetTelemetryHeaders()
		if !reflect.DeepEqual(headers, expected) {
			t.Errorf("expected %v, got %v", expected, headers)
		}
	})

	t.Run("ignores malformed pairs and keeps values containing '='", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		viper.Set(TelemetryHeadersKey, "Authorization=Bearer=dabc,invalid-pair")

		expected := map[string]string{
			"Authorization": "Bearer=dabc",
		}
		headers := GetTelemetryHeaders()
		if !reflect.DeepEqual(headers, expected) {
			t.Errorf("expected %v, got %v", expected, headers)
		}
	})
}
