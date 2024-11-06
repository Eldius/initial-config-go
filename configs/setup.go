package configs

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"log"
	"os"
	"path/filepath"
)

const (
	LogFormatKey         = "log.format"
	LogLevelKey          = "log.level"
	LogOutputFileKey     = "log.output_to_file"
	LogOutputToStdoutKey = "log.output_to_stdout"

	LogFormatJSON = "json"
	LogFormatText = "text"
	LogLevelINFO  = "info"
	LogLevelDEBUG = "debug"
	LogLevelWARN  = "warn"
	LogLevelERROR = "error"
)

var (
	logKeys = []string{
		"host",
		"service.name",
		"level",
		"message",
		"time",
		"error",
		"source",
		"function",
		"file",
		"line",
	}
)

type SetupOptions struct {
	CfgFile       string
	DefaultValues map[string]any
}

type OptionFunc func(*SetupOptions)

// WithConfigFile defines the app configuration file
// to be used
func WithConfigFile(file string) OptionFunc {
	return func(o *SetupOptions) {
		o.CfgFile = file
	}
}

// WithDefaultValues adds default values to Viper configuration
//
// Default configuration keys:
//   - Logs configuration
//   - Log level:
//   - Log format:            `log.format` (accepts `text` or `json`)
//   - Log level:             `log.level` (accepts `info`, `debug`, `warn` or `error`)
//   - Log output file:       `log.output_to_file` (output file path as a string)
//   - Log to stdout:         `log.output_to_stdout` (accepts `true` or `false`)
func WithDefaultValues(vals map[string]any) OptionFunc {
	return func(o *SetupOptions) {
		if o.DefaultValues == nil {
			o.DefaultValues = vals
			return
		}
		maps.Copy(o.DefaultValues, vals)
	}
}

// InitSetup sets up application default configurations
// for spf13/viper and slog libraries
func InitSetup(appName string, opts ...OptionFunc) error {
	cfg := SetupOptions{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfg.CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(filepath.Join(home, fmt.Sprintf(".%s", appName)))
		viper.AddConfigPath(filepath.Join(home))
		viper.SetConfigType("yaml")
		viper.SetConfigName(appName)
	}

	setDefaults(cfg.DefaultValues)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("Could not find config file using default values: %s", err)
	}

	return setupLogs(appName, GetLogFormat(), GetLogLevel(), GetLogOutput(), GetLogToStdout())
}

func setDefaults(defaultValues map[string]any) {
	if _, ok := defaultValues[LogLevelKey]; !ok {
		defaultValues[LogLevelKey] = LogLevelINFO
	}
	if _, ok := defaultValues[LogFormatKey]; !ok {
		defaultValues[LogFormatKey] = LogFormatText
	}
	if _, ok := defaultValues[LogOutputToStdoutKey]; !ok {
		defaultValues[LogOutputToStdoutKey] = false
	}
	for k, v := range defaultValues {
		viper.SetDefault(k, v)
	}
}
