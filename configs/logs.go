package configs

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func setupLogs(appName, format, level, logOutputFile string, stdout bool) error {
	h, err := logHandler(appName, format, level, logOutputFile, stdout)
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

func logHandler(appName, format, level, outputFile string, stdout bool) (slog.Handler, error) {
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