package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitTelemetry(t *testing.T) {
	t.Run("disabled telemetry returns nil", func(t *testing.T) {
		err := InitTelemetry(t.Context(), WithOtelEnabled(false))
		assert.NoError(t, err)
	})

	t.Run("default config does not enable telemetry", func(t *testing.T) {
		err := InitTelemetry(t.Context())
		assert.NoError(t, err)
	})
}

func TestTraceErrorMessages(t *testing.T) {
	t.Run("trace error messages should use correct terminology", func(t *testing.T) {
		assert.Contains(t, ErrTracesInitialization.Error(), "tracer",
			"ErrTracesInitialization should mention 'tracer', not 'meter'")
		assert.NotContains(t, ErrTracesInitialization.Error(), "meter")

		assert.Contains(t, ErrTracesConnectionInitialization.Error(), "traces",
			"ErrTracesConnectionInitialization should mention 'traces', not 'metrics'")
		assert.NotContains(t, ErrTracesConnectionInitialization.Error(), "metrics")

		assert.Contains(t, ErrTracesExporterInitialization.Error(), "trace",
			"ErrTracesExporterInitialization should mention 'trace', not 'metric'")
		assert.NotContains(t, ErrTracesExporterInitialization.Error(), "metric")
	})

	t.Run("meter error messages are correct", func(t *testing.T) {
		assert.Contains(t, ErrMeterInitialization.Error(), "meter")
		assert.Contains(t, ErrMetricsConnectionInitialization.Error(), "metrics")
		assert.Contains(t, ErrMetricsExporterInitialization.Error(), "metric")
	})
}
