package setup

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/eldius/initial-config-go/configs"
)

var (
	// ErrInvalidLogOutputConfig is returned when neither stdout nor file output is configured for logging.
	ErrInvalidLogOutputConfig = errors.New("invalid log output configuration: should enable stdout or define an output file")
)

func setupLogs(appName, format, level, logOutputFile string, stdout bool, keysToRedact ...string) error {
	if !stdout && logOutputFile == "" {
		return fmt.Errorf("%w: logOutputFile: %s / stdout: %v", ErrInvalidLogOutputConfig, logOutputFile, stdout)
	}

	for i, key := range keysToRedact {
		keysToRedact[i] = strings.ToLower(key)
	}

	writer, err := getWriter(logOutputFile, stdout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLogOutputConfig, err)
	}
	h, err := logHandler(format, level, writer, keysToRedact...)
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

func getWriter(outputFile string, logToStdout bool) (io.Writer, error) {
	var w io.Writer
	if logToStdout {
		w = os.Stdout
	}
	if outputFile != "" {
		outputFile, err := filepath.Abs(outputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path to log file: %w", err)
		}
		outFile, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open output file %s: %w", outputFile, err)
		}
		if w == nil {
			return outFile, nil
		}

		w = io.MultiWriter(outFile, w)
	}

	return w, nil
}

func logHandler(format, level string, w io.Writer, keysToRedact ...string) (slog.Handler, error) {
	if strings.ToLower(format) == configs.LogFormatJSON {
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       parseLogLevel(level),
			ReplaceAttr: logAttrsReplacerFunc(),
		})
		if len(keysToRedact) == 0 {
			return handler, nil
		}
		return newRedactHandler(handler, keysToRedact), nil
	}
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLogLevel(level),
		ReplaceAttr: logAttrsReplacerFunc(),
	})
	if len(keysToRedact) == 0 {
		return handler, nil
	}

	if len(keysToRedact) == 0 {
		return handler, nil
	}
	return newRedactHandler(handler, keysToRedact), nil
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

func logAttrsReplacerFunc() func(groups []string, a slog.Attr) slog.Attr {
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
		return a
	}
}
