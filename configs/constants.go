package configs

const (
	// Configuration keys for logging
	LogFormatKey         = "log.format"
	LogLevelKey          = "log.level"
	LogOutputFileKey     = "log.output_to_file"
	LogOutputToStdoutKey = "log.output_to_stdout"
	LogKeysToRedactKey   = "log.redacted_keys"

	// Log format constants
	LogFormatJSON = "json"
	LogFormatText = "text"

	// Log level constants
	LogLevelINFO  = "info"
	LogLevelDEBUG = "debug"
	LogLevelWARN  = "warn"
	LogLevelERROR = "error"

	// Configuration keys for telemetry
	TelemetryEnabledKey                = "telemetry.enabled"
	TelemetryTracesBackendEndpointKey  = "telemetry.traces.endpoint"
	TelemetryMetricsBackendEndpointKey = "telemetry.metrics.endpoint"
	TelemetryLogsBackendEndpointKey    = "telemetry.logs.endpoint"
)

var (
	// DefaultConfigValuesMap contains the default values for all configuration keys.
	DefaultConfigValuesMap = map[string]any{
		LogFormatKey:                       LogFormatJSON,
		LogLevelKey:                        LogLevelINFO,
		LogOutputFileKey:                   "",
		LogOutputToStdoutKey:               false,
		LogKeysToRedactKey:                 []string{},
		TelemetryEnabledKey:                false,
		TelemetryTracesBackendEndpointKey:  "",
		TelemetryMetricsBackendEndpointKey: "",
		TelemetryLogsBackendEndpointKey:    "",
	}

	// DefaultConfigValuesLogFileMap provides defaults with file logging enabled.
	DefaultConfigValuesLogFileMap = addAttr(DefaultConfigValuesMap, LogOutputFileKey, "execution.log")

	// DefaultConfigValuesLogStdoutMap provides defaults with stdout logging enabled.
	DefaultConfigValuesLogStdoutMap = addAttr(DefaultConfigValuesMap, LogOutputToStdoutKey, true)

	// DefaultConfigValuesLogAllMap provides defaults with both file and stdout logging enabled.
	DefaultConfigValuesLogAllMap = addAttr(addAttr(DefaultConfigValuesMap, LogOutputToStdoutKey, true), LogOutputFileKey, "execution.log")
)
