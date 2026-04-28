package configs

import (
	"maps"
	"strings"

	"github.com/spf13/viper"
)

// GetLogOutputFile returns the configured log output file path.
// Returns an empty string if file logging is disabled.
func GetLogOutputFile() string {
	return viper.GetString(LogOutputFileKey)
}

// GetLogToStdout returns whether logging to stdout is enabled.
func GetLogToStdout() bool {
	return viper.GetBool(LogOutputToStdoutKey)
}

// GetLogLevel returns the configured log level (info, debug, warn, or error).
func GetLogLevel() string {
	return strings.ToLower(viper.GetString(LogLevelKey))
}

// GetLogFormat returns the configured log format (JSON or text).
func GetLogFormat() string {
	return strings.ToLower(viper.GetString(LogFormatKey))
}

// GetLogKeysToRedact returns the list of log attribute keys that should be redacted.
func GetLogKeysToRedact() []string {
	val := viper.Get(LogKeysToRedactKey)
	if val == nil {
		return []string{}
	}
	if s, ok := val.(string); ok {
		if s == "" {
			return []string{}
		}
		return strings.Split(s, ",")
	}
	return viper.GetStringSlice(LogKeysToRedactKey)
}

// GetTelemetryEnabled returns whether OpenTelemetry is enabled.
func GetTelemetryEnabled() bool {
	return viper.GetBool(TelemetryEnabledKey)
}

// GetTraceBackendEndpoint returns the configured OTLP trace backend endpoint.
func GetTraceBackendEndpoint() string {
	return viper.GetString(TelemetryTracesBackendEndpointKey)
}

// GetMetricsBackendEndpoint returns the configured OTLP metrics backend endpoint.
func GetMetricsBackendEndpoint() string {
	return viper.GetString(TelemetryMetricsBackendEndpointKey)
}

// GetLogsBackendEndpoint returns the configured OTLP logs backend endpoint.
func GetLogsBackendEndpoint() string {
	return viper.GetString(TelemetryLogsBackendEndpointKey)
}

// ConfigOptionFunc is a function type for configuring default options.
type ConfigOptionFunc func(defaultOptions map[string]any)

func addAttr(m map[string]any, key string, value any) map[string]any {
	nm := make(map[string]any)
	maps.Copy(nm, m)
	nm[key] = value

	return nm
}
