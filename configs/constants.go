package configs

const (
	LogFormatKey         = "log.format"
	LogLevelKey          = "log.level"
	LogOutputFileKey     = "log.output_to_file"
	LogOutputToStdoutKey = "log.output_to_stdout"
	LogKeysToRedactKey   = "log.redacted_keys"

	LogFormatJSON = "json"
	LogFormatText = "text"
	LogLevelINFO  = "info"
	LogLevelDEBUG = "debug"
	LogLevelWARN  = "warn"
	LogLevelERROR = "error"

	TelemetryEnabledKey                = "telemetry.enabled"
	TelemetryTracesBackendEndpointKey  = "telemetry.traces.endpoint"
	TelemetryMetricsBackendEndpointKey = "telemetry.metrics.endpoint"
)

var (
	// DefaultConfigValuesMap contains the default values for the initial
	// configuration keys
	DefaultConfigValuesMap = map[string]any{
		LogFormatKey:                       LogFormatJSON,
		LogLevelKey:                        LogLevelINFO,
		LogOutputFileKey:                   "",
		LogOutputToStdoutKey:               false,
		LogKeysToRedactKey:                 []string{},
		TelemetryEnabledKey:                false,
		TelemetryTracesBackendEndpointKey:  "",
		TelemetryMetricsBackendEndpointKey: "",
	}
)
