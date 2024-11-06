package configs

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
//   - Log output file:       `log.output_to_file`   (output file path as a string)
func WithDefaultValues(vals map[string]any) OptionFunc {
	return func(o *SetupOptions) {
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

	return setupLogs(appName, GetLogFormat(), GetLogLevel(), GetLogOutput())
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

func setupLogs(appName, format, level, logOutputFile string) error {
	h, err := logHandler(appName, format, level, logOutputFile)
	if err != nil {
		return fmt.Errorf("failed to create log handler: %w", err)
	}
	logger := slog.New(h)
	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	slog.SetDefault(logger.With(
		slog.String("service.name", appName),
		slog.String("host", host),
	))

	return nil
}

func parseLogLevel(lvl string) slog.Level {
	switch strings.ToLower(lvl) {
	case LogLevelDEBUG:
		return slog.LevelDebug
	case LogLevelWARN:
		return slog.LevelWarn
	case LogLevelERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

const (
	LogFormatKey         = "log.format"
	LogLevelKey          = "log.level"
	LogOutputfileKey     = "log.output_to_file"
	LogOutputToStdoutKey = "log.output_to_stdout"

	LogFormatJSON = "json"
	LogFormatText = "text"
	LogLevelINFO  = "info"
	LogLevelDEBUG = "debug"
	LogLevelWARN  = "warn"
	LogLevelERROR = "error"
)

func GetLogOutput() string {
	return viper.GetString(LogOutputfileKey)
}

func GetLogLevel() string {
	return strings.ToLower(viper.GetString(LogLevelKey))
}

func GetLogFormat() string {
	return strings.ToLower(viper.GetString(LogFormatKey))
}

func logHandler(appName, format, level, outputFile string) (slog.Handler, error) {
	var w io.Writer = os.Stdout
	if out := outputFile; out != "" {
		out, err := filepath.Abs(out)
		if err != nil {
			err = fmt.Errorf("parsing log absolute path: %w", err)
			return nil, err
		}
		f, err := os.Create(out)
		if err != nil {
			err = fmt.Errorf("opening log file: %w", err)
			return nil, err
		}
		//w = io.MultiWriter(w, f)
		w = f
	}
	if strings.ToLower(format) == LogFormatJSON {
		return slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       parseLogLevel(level),
			ReplaceAttr: logAttrsReplacerFunc(appName),
		}), nil
	}
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLogLevel(level),
		ReplaceAttr: logAttrsReplacerFunc(appName),
	}), nil
}

func logAttrsReplacerFunc(appName string) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if slices.Contains(logKeys, a.Key) {
			return a
		}
		if a.Key == "msg" {
			a.Key = "message"
			return a
		}
		a.Key = fmt.Sprintf("custom.%s.%s", appName, a.Key)
		return a
	}
}
