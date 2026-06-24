package telemetry_test

import (
	"testing"

	"github.com/eldius/initial-config-go/setup"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryShutdown(t *testing.T) {
	t.Run("TelemetryShutdown with no providers should not error", func(t *testing.T) {
		err := telemetry.TelemetryShutdown(t.Context())
		assert.NoError(t, err)
	})

	t.Run("TelemetryShutdown after InitSetup should not panic", func(t *testing.T) {
		err := setup.InitSetup(t.Context(), "test-app-shutdown")
		require.NoError(t, err)
		err = telemetry.TelemetryShutdown(t.Context())
		// Should not panic; may or may not error depending on state
		t.Logf("TelemetryShutdown result: %v", err)
	})
}
