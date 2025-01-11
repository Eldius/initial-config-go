package configs

import (
	"github.com/spf13/viper"
	"strings"
)

func GetLogOutput() string {
	return viper.GetString(LogOutputFileKey)
}

func GetLogToStdout() bool {
	return viper.GetBool(LogOutputToStdoutKey)
}

func GetLogLevel() string {
	return strings.ToLower(viper.GetString(LogLevelKey))
}

func GetLogFormat() string {
	return strings.ToLower(viper.GetString(LogFormatKey))
}

func GetLogKeysToRedact() []string {
	return viper.GetStringSlice(LogKeysToRedactKey)
}
