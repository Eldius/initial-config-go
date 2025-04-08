package configs

import (
	"github.com/spf13/viper"
	"strings"
	"sync"
)

func GetLogOutput() string {
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
	return viper.GetStringSlice(LogKeysToRedactKey)
}
