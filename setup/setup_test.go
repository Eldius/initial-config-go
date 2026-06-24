package setup

import (
	"testing"

	"github.com/eldius/initial-config-go/logs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitSetup_DefaultValues(t *testing.T) {
	t.Run("InitSetup with no options should succeed and log to stdout by default", func(t *testing.T) {
		err := InitSetup(t.Context(), "test-app-defaults")
		assert.NoError(t, err, "InitSetup with no options should not return an error")
	})

	t.Run("InitSetup with empty app name should return an error", func(t *testing.T) {
		err := InitSetup(t.Context(), "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyAppName)
	})

	t.Run("InitSetup with custom stdout setting should succeed", func(t *testing.T) {
		err := InitSetup(t.Context(), "test-app-custom",
			WithDefaultValues(map[string]any{
				"log.output_to_stdout": true,
				"log.format":           "text",
				"log.level":            "debug",
			}),
		)
		assert.NoError(t, err)
	})

	t.Run("InitSetup with file output should succeed", func(t *testing.T) {
		err := InitSetup(t.Context(), "test-app-file",
			WithDefaultValues(map[string]any{
				"log.output_to_file": "/tmp/test-init-setup.log",
				"log.format":         "json",
			}),
		)
		assert.NoError(t, err)
	})
}

func TestGetDefaultValues(t *testing.T) {
	t.Run("GetDefaultValues returns expected defaults", func(t *testing.T) {
		opts := Options{}
		defaults := opts.GetDefaultValues()

		assert.Equal(t, "info", defaults["log.level"])
		assert.Equal(t, "text", defaults["log.format"])
		assert.Equal(t, true, defaults["log.output_to_stdout"], "log.output_to_stdout should default to true")
	})

	t.Run("GetDefaultValues does not override explicitly set values", func(t *testing.T) {
		opts := Options{
			DefaultValues: map[string]any{
				"log.level":            "debug",
				"log.format":           "json",
				"log.output_to_stdout": false,
			},
		}
		defaults := opts.GetDefaultValues()

		assert.Equal(t, "debug", defaults["log.level"])
		assert.Equal(t, "json", defaults["log.format"])
		assert.Equal(t, false, defaults["log.output_to_stdout"], "explicitly set false should not be overridden")
	})

	t.Run("GetDefaultValues does not overwrite nil map when provided", func(t *testing.T) {
		opts := Options{
			DefaultValues: map[string]any{
				"custom.key": "custom-value",
			},
		}
		defaults := opts.GetDefaultValues()

		assert.Equal(t, "custom-value", defaults["custom.key"])
		assert.Equal(t, true, defaults["log.output_to_stdout"])
	})
}

func TestWithOpenTelemetryOptions(t *testing.T) {
	t.Run("WithOpenTelemetryOptions should append options", func(t *testing.T) {
		opts := Options{}
		optFunc := WithOpenTelemetryOptions()
		optFunc(&opts)
		assert.Empty(t, opts.OpenTelemetryOptions)

		optFunc2 := WithOpenTelemetryOptions(nil)
		optFunc2(&opts)
		assert.Len(t, opts.OpenTelemetryOptions, 1)
	})
}

func TestWithDefaultValues(t *testing.T) {
	t.Run("WithDefaultValues should merge values correctly", func(t *testing.T) {
		opts := Options{
			DefaultValues: map[string]any{
				"existing.key": "existing-value",
			},
		}
		optFunc := WithDefaultValues(map[string]any{
			"new.key": "new-value",
		})
		optFunc(&opts)

		assert.Equal(t, "existing-value", opts.DefaultValues["existing.key"])
		assert.Equal(t, "new-value", opts.DefaultValues["new.key"])
	})
}

func TestWithProps(t *testing.T) {
	t.Run("WithProps should add properties to DefaultValues", func(t *testing.T) {
		opts := Options{}
		optFunc := WithProps(
			Prop{Key: "prop.key1", Value: "value1"},
			Prop{Key: "prop.key2", Value: 42},
		)
		optFunc(&opts)

		assert.Equal(t, "value1", opts.DefaultValues["prop.key1"])
		assert.Equal(t, 42, opts.DefaultValues["prop.key2"])
	})
}

func TestGetEnvPrefix(t *testing.T) {
	t.Run("default env prefix should be lowercase 'app'", func(t *testing.T) {
		opts := Options{}
		assert.Equal(t, "app", opts.GetEnvPrefix())
	})

	t.Run("custom env prefix should be lowercased", func(t *testing.T) {
		opts := Options{}
		WithEnvPrefix("MYAPP")(&opts)
		assert.Equal(t, "myapp", opts.GetEnvPrefix())
	})
}

func TestGetDefaultCfgFileName(t *testing.T) {
	t.Run("default config file name should be 'config'", func(t *testing.T) {
		opts := Options{}
		assert.Equal(t, "config", opts.GetDefaultCfgFileName())
	})

	t.Run("custom config file name should be returned as-is", func(t *testing.T) {
		opts := Options{DefaultCfgFileName: "settings"}
		assert.Equal(t, "settings", opts.GetDefaultCfgFileName())
	})
}

func TestGetDefaultCfgFileLocations(t *testing.T) {
	t.Run("default locations should include home dir and current dir", func(t *testing.T) {
		opts := Options{}
		locs := opts.GetDefaultCfgFileLocations("test-app")
		require.GreaterOrEqual(t, len(locs), 2)
		assert.Contains(t, locs[0], "test-app")
		assert.Equal(t, ".", locs[1])
	})
}

func TestWithInstrumentHTTPClient(t *testing.T) {
	t.Run("option sets the field to true", func(t *testing.T) {
		opts := Options{}
		WithInstrumentHTTPClient(true)(&opts)
		assert.True(t, opts.InstrumentHTTPClient)
	})

	t.Run("option sets the field to false", func(t *testing.T) {
		opts := Options{}
		WithInstrumentHTTPClient(false)(&opts)
		assert.False(t, opts.InstrumentHTTPClient)
	})

	t.Run("default value is false", func(t *testing.T) {
		opts := Options{}
		assert.False(t, opts.InstrumentHTTPClient)
	})
}

func TestCloseLogFiles(t *testing.T) {
	t.Run("CloseLogFiles with no files should not error", func(t *testing.T) {
		err := logs.CloseLogFiles()
		assert.NoError(t, err)
	})

	t.Run("CloseLogFiles after file-based InitSetup should succeed", func(t *testing.T) {
		err := InitSetup(t.Context(), "test-app-close-files",
			WithDefaultValues(map[string]any{
				"log.output_to_file": "/tmp/test-close-files.log",
			}),
		)
		require.NoError(t, err)
		err = logs.CloseLogFiles()
		assert.NoError(t, err)
		// Calling twice should be safe (files already closed, slice cleared)
		err = logs.CloseLogFiles()
		assert.NoError(t, err)
	})
}
