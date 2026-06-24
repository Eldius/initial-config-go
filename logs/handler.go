package logs

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/eldius/initial-config-go/configs"
)

var logFiles []*os.File

func LogHandler(format, level string, w io.Writer, keysToRedact ...string) (slog.Handler, error) {
	if strings.ToLower(format) == configs.LogFormatJSON {
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   true,
			Level:       parseLogLevel(level),
			ReplaceAttr: LogAttrsReplacerFunc(),
		})
		if len(keysToRedact) == 0 {
			return handler, nil
		}
		return NewRedactHandler(handler, keysToRedact), nil
	}
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		AddSource:   true,
		Level:       parseLogLevel(level),
		ReplaceAttr: LogAttrsReplacerFunc(),
	})
	if len(keysToRedact) == 0 {
		return handler, nil
	}
	return NewRedactHandler(handler, keysToRedact), nil
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

func GetWriter(outputFile string, logToStdout bool) (io.Writer, error) {
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
		logFiles = append(logFiles, outFile)
		if w == nil {
			return outFile, nil
		}

		w = io.MultiWriter(outFile, w)
	}

	return w, nil
}

func CloseLogFiles() error {
	var errs []error
	for _, f := range logFiles {
		if err := f.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	logFiles = nil
	if len(errs) > 0 {
		return fmt.Errorf("failed to close log files: %w", errors.Join(errs...))
	}
	return nil
}
