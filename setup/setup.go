package setup

import (
	"context"
	"errors"
	"fmt"
	"github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var (
	ErrEmptyAppName = errors.New("appName is empty")
)

type Options struct {
	CfgFilePathToBeUsed     string
	DefaultCfgFileLocations []string
	DefaultCfgFileName      string
	DefaultValues           map[string]any
	OpenTelemetryOptions    []telemetry.Option
}

func (o *Options) GetDefaultValues() map[string]any {
	if o.DefaultValues == nil {
		o.DefaultValues = make(map[string]any)
	}

	if _, ok := o.DefaultValues[configs.LogLevelKey]; !ok {
		o.DefaultValues[configs.LogLevelKey] = configs.LogLevelINFO
	}
	if _, ok := o.DefaultValues[configs.LogFormatKey]; !ok {
		o.DefaultValues[configs.LogFormatKey] = configs.LogFormatText
	}
	if _, ok := o.DefaultValues[configs.LogOutputToStdoutKey]; !ok {
		o.DefaultValues[configs.LogOutputToStdoutKey] = false
	}
	return o.DefaultValues
}

// WithOpenTelemetryOptions sets OpenTelemetry options
func WithOpenTelemetryOptions(opts ...telemetry.Option) OptionFunc {
	return func(o *Options) {
		o.OpenTelemetryOptions = opts
	}
}

// GetDefaultCfgFileName returns default config file name
func (o *Options) GetDefaultCfgFileName() string {
	if o.DefaultCfgFileName == "" {
		o.DefaultCfgFileName = "config"
	}

	return o.DefaultCfgFileName
}

// GetDefaultCfgFileLocations returns default config file locations
func (o *Options) GetDefaultCfgFileLocations(appName string) []string {
	if o.DefaultCfgFileLocations == nil {
		o.DefaultCfgFileLocations = []string{
			fmt.Sprintf("~/.%s", appName),
			".",
		}
	}
	return o.DefaultCfgFileLocations
}

// OptionFunc customization option
type OptionFunc func(*Options)

// WithDefaultCfgFileLocations defines locations to search for config files
func WithDefaultCfgFileLocations(f ...string) OptionFunc {
	return func(o *Options) {
		o.DefaultCfgFileLocations = f
	}
}

// WithDefaultCfgFileName defines default config file name
func WithDefaultCfgFileName(f string) OptionFunc {
	return func(o *Options) {
		o.DefaultCfgFileName = f
	}
}

// WithConfigFileToBeUsed defines the app configuration file
// to be used
func WithConfigFileToBeUsed(file string) OptionFunc {
	return func(o *Options) {
		o.CfgFilePathToBeUsed = file
	}
}

// WithDefaultValues adds default values to Viper configuration
//
// Default configuration keys:
//   - Logs configuration
//   - Log level:
//   - Log format: `log.format` (accepts `text` or `json`)
//   - Log level: `log.level` (accepts `info`, `debug`, `warn` or `error`)
//   - Log output file: `log.output_to_file` (output file path as a string)
//   - Log to stdout: `log.output_to_stdout` (accepts `true` or `false`)
func WithDefaultValues(vals map[string]any) OptionFunc {
	return func(o *Options) {
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
	if appName == "" {
		return fmt.Errorf("invalid app name: %ww", ErrEmptyAppName)
	}

	cfg := Options{}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.CfgFilePathToBeUsed != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfg.CfgFilePathToBeUsed)
	} else {
		// Find a home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(filepath.Join(home, fmt.Sprintf(".%s", appName)))
		viper.AddConfigPath(filepath.Join(home))
		for _, f := range cfg.GetDefaultCfgFileLocations(appName) {
			abs, err := absolutePath(f)
			if err != nil {
				err = fmt.Errorf("cannot resolve absolute path of config file '%s': %v", f, err)
				log.Printf("failed to get absolute path for config file location: %s", err)
			}
			viper.AddConfigPath(abs)
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName(cfg.GetDefaultCfgFileName())
	}

	setDefaults(cfg.GetDefaultValues())

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Printf("Could not find config file using default values: %s", err)
	}

	if err := telemetry.InitTelemetry(context.Background(), cfg.OpenTelemetryOptions...); err != nil {
		return err
	}

	if err := setupLogs(
		appName,
		configs.GetLogFormat(),
		configs.GetLogLevel(),
		configs.GetLogOutput(),
		configs.GetLogToStdout(),
		configs.GetLogKeysToRedact()...,
	); err != nil {
		return fmt.Errorf("setupLogs: %w", err)
	}

	return nil
}

func setDefaults(defaultValues map[string]any) {
	for k, v := range defaultValues {
		viper.SetDefault(k, v)
	}
}

func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}

func absolutePath(path string) (string, error) {
	path, err := expandPath(path)
	if err != nil {
		return "", fmt.Errorf("expanded path: %w", err)
	}
	return filepath.Abs(path)
}
