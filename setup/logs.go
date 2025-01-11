package setup

import (
	"errors"
	"fmt"
	"github.com/eldius/initial-config-go/configs"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	InvalidLogOutputConfigErr = errors.New("invalid log output configuration: should enable stdout or define an output file")
)

func setupLogs(appName, format, level, logOutputFile string, stdout bool, keysToRedact ...string) error {
	if !stdout && logOutputFile == "" {
		return fmt.Errorf("%w: logOutputFile: %s / stdout: %v", InvalidLogOutputConfigErr, logOutputFile, stdout)
	}
	h, err := logHandler(appName, format, level, logOutputFile, stdout, keysToRedact...)
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
	case configs.LogLevelDEBUG:
		return slog.LevelDebug
	case configs.LogLevelWARN:
		return slog.LevelWarn
	case configs.LogLevelERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func logHandler(appName, format, level, outputFile string, stdout bool, keysToRedact ...string) (slog.Handler, error) {
	var w io.Writer
	if stdout {
		w = os.Stdout
	}
	if outputFile != "" {
		out, err := filepath.Abs(outputFile)
		if err != nil {
			err = fmt.Errorf("parsing log absolute path: %w", err)
			return nil, err
		}
		f, err := os.Create(out)
		if err != nil {
			err = fmt.Errorf("opening log file: %w", err)
			return nil, err
		}
		if w != nil {
			w = io.MultiWriter(w, f)
		} else {
			w = f
		}
	}
	//keysToRedact := make([]string, 0)

	if strings.ToLower(format) == configs.LogFormatJSON {
		return newRedactHandler(slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       parseLogLevel(level),
			ReplaceAttr: logAttrsReplacerFunc(appName),
		}),
			keysToRedact,
		), nil
	}
	return newRedactHandler(slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLogLevel(level),
		ReplaceAttr: logAttrsReplacerFunc(appName),
	}), keysToRedact), nil
}

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

func logAttrsReplacerFunc(appName string) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if slices.Contains(logKeys, a.Key) {
			return a
		}
		if strings.HasPrefix(a.Key, "request") ||
			strings.HasPrefix(a.Key, "response") ||
			strings.HasPrefix(a.Key, "service") {
			return a
		}

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
