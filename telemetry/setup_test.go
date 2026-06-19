package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsEnabled(t *testing.T) {
	t.Run("disabled when Enabled is false", func(t *testing.T) {
		cfg := &OTELConfigs{
			Enabled: false,
		}
		cfg.Endpoints.Traces = "localhost:4317"
		cfg.Endpoints.Metrics = "localhost:4317"
		assert.False(t, cfg.IsEnabled())
	})

	t.Run("disabled when no endpoints are set", func(t *testing.T) {
		cfg := &OTELConfigs{Enabled: true}
		assert.False(t, cfg.IsEnabled())
	})

	t.Run("enabled with only traces endpoint", func(t *testing.T) {
		cfg := &OTELConfigs{Enabled: true}
		cfg.Endpoints.Traces = "localhost:4317"
		assert.True(t, cfg.IsEnabled())
	})

	t.Run("enabled with only metrics endpoint", func(t *testing.T) {
		cfg := &OTELConfigs{Enabled: true}
		cfg.Endpoints.Metrics = "localhost:4317"
		assert.True(t, cfg.IsEnabled())
	})

	t.Run("enabled with only logs endpoint", func(t *testing.T) {
		cfg := &OTELConfigs{Enabled: true}
		cfg.Endpoints.Logs = "localhost:4317"
		assert.True(t, cfg.IsEnabled(), "IsEnabled should return true when only logs endpoint is set")
	})

	t.Run("enabled with all endpoints set", func(t *testing.T) {
		cfg := &OTELConfigs{Enabled: true}
		cfg.Endpoints.Traces = "traces:4317"
		cfg.Endpoints.Metrics = "metrics:4317"
		cfg.Endpoints.Logs = "logs:4317"
		assert.True(t, cfg.IsEnabled())
	})
}

func TestNewDefaultCfg(t *testing.T) {
	t.Run("default config should have zero values", func(t *testing.T) {
		cfg := NewDefaultCfg()
		assert.False(t, cfg.Enabled)
		assert.Empty(t, cfg.Endpoints.Traces)
		assert.Empty(t, cfg.Endpoints.Metrics)
		assert.Empty(t, cfg.Endpoints.Logs)
		assert.Empty(t, cfg.Service.Name)
	})
}

func TestOptions(t *testing.T) {
	t.Run("WithTraceEndpoint sets the endpoint", func(t *testing.T) {
		cfg := NewDefaultCfg()
		WithTraceEndpoint("my-trace:4317")(cfg)
		assert.Equal(t, "my-trace:4317", cfg.Endpoints.Traces)
	})

	t.Run("WithMetricEndpoint sets the endpoint", func(t *testing.T) {
		cfg := NewDefaultCfg()
		WithMetricEndpoint("my-metrics:4317")(cfg)
		assert.Equal(t, "my-metrics:4317", cfg.Endpoints.Metrics)
	})

	t.Run("WithLogsEndpoint sets the endpoint", func(t *testing.T) {
		cfg := NewDefaultCfg()
		WithLogsEndpoint("my-logs:4317")(cfg)
		assert.Equal(t, "my-logs:4317", cfg.Endpoints.Logs)
	})

	t.Run("WithOtelEnabled enables telemetry", func(t *testing.T) {
		cfg := NewDefaultCfg()
		WithOtelEnabled(true)(cfg)
		assert.True(t, cfg.Enabled)
	})

	t.Run("WithService sets all service fields", func(t *testing.T) {
		cfg := NewDefaultCfg()
		WithService("my-svc", "1.0.0", "prod")(cfg)
		assert.Equal(t, "my-svc", cfg.Service.Name)
		assert.Equal(t, "1.0.0", cfg.Service.Version)
		assert.Equal(t, "prod", cfg.Service.Environment)
	})
}
