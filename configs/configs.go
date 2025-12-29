package configs

import (
	"maps"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

func GetLogOutputFile() string {
	return sync.OnceValue(func() string {
		return viper.GetString(LogOutputFileKey)
	})()
}

func GetLogToStdout() bool {
	return sync.OnceValue(func() bool {
		return viper.GetBool(LogOutputToStdoutKey)
	})()
}

func GetLogLevel() string {
	return sync.OnceValue(func() string {
		return strings.ToLower(viper.GetString(LogLevelKey))
	})()

}

func GetLogFormat() string {
	return sync.OnceValue(func() string {
		return strings.ToLower(viper.GetString(LogFormatKey))
	})()
}

func GetLogKeysToRedact() []string {
	return sync.OnceValue(func() []string {
		return viper.GetStringSlice(LogKeysToRedactKey)
	})()
}

func GetTelemetryEnabled() bool {
	return sync.OnceValue(func() bool {
		return viper.GetBool(TelemetryEnabledKey)
	})()
}

func GetTraceBackendEndpoint() string {
	return sync.OnceValue(func() string {
		return viper.GetString(TelemetryTracesBackendEndpointKey)
	})()
}

func GetMetricsBackendEndpoint() string {
	return sync.OnceValue(func() string {
		return viper.GetString(TelemetryMetricsBackendEndpointKey)
	})()
}

type ConfigOptionFunc func(defaultOptions map[string]any)

func addAttr(m map[string]any, key string, value any) map[string]any {
	nm := make(map[string]any)
	maps.Copy(nm, m)
	nm[key] = value

	return nm
}
