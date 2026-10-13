package configs

import (
	"reflect"
	"testing"
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
