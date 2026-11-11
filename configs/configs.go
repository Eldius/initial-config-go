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

// GetTelemetryDebugEnabled returns whether OpenTelemetry debug mode is enabled.
func GetTelemetryDebugEnabled() bool {
	return viper.GetBool(TelemetryDebugKey)
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

// GetTelemetryHeaders returns the configured OTLP headers, sent with each
// exporter request (e.g. for authentication). In config files it is a map:
//
//	telemetry:
//	  headers:
//	    x-api-key: secret
//
// As an environment variable it is a comma-separated list of key=value pairs:
//
//	APP_TELEMETRY_HEADERS="Authorization=Bearer token,X-Api-Key=secret"
func GetTelemetryHeaders() map[string]string {
	val := viper.Get(TelemetryHeadersKey)
	if val == nil {
		return map[string]string{}
	}
	if s, ok := val.(string); ok {
		return parseHeadersString(s)
	}
	return viper.GetStringMapString(TelemetryHeadersKey)
}

func parseHeadersString(s string) map[string]string {
	headers := make(map[string]string)
	if s == "" {
		return headers
	}
	for pair := range strings.SplitSeq(s, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return headers
}

// ConfigOptionFunc is a function type for configuring default options.
type ConfigOptionFunc func(defaultOptions map[string]any)

func addAttr(m map[string]any, key string, value any) map[string]any {
	nm := make(map[string]any)
	maps.Copy(nm, m)
	nm[key] = value

	return nm
}

// MergeConfigMaps merges two config maps
func MergeConfigMaps(a map[string]any, b map[string]any) map[string]any {
	nm := make(map[string]any)
	maps.Copy(nm, b)
	maps.Copy(nm, a)

	return nm
}
