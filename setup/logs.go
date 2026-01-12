package setup

import (
	"context"
	"errors"
	"fmt"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/eldius/initial-config-go/configs"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
)

var (
	// ErrInvalidLogOutputConfig is returned when neither stdout nor file output is configured for logging.
	ErrInvalidLogOutputConfig = errors.New("invalid log output configuration: should enable stdout or define an output file")
)

func initLogs(ctx context.Context, appName string, options Options) error {
	return setupLogs(ctx, appName, configs.GetLogFormat(), configs.GetLogLevel(), configs.GetLogOutputFile(), configs.GetLogToStdout(), options, configs.GetLogKeysToRedact()...)
}

func setupLogs(ctx context.Context, appName, format, level, logOutputFile string, stdout bool, options Options, keysToRedact ...string) error {

	cfg := telemetry.NewDefaultCfg()

	for _, o := range options.OpenTelemetryOptions {
		o(cfg)
	}

	if !stdout && logOutputFile == "" {
		return fmt.Errorf("%w: logOutputFile: %s / stdout: %v", ErrInvalidLogOutputConfig, logOutputFile, stdout)
	}

	for i, key := range keysToRedact {
		keysToRedact[i] = strings.ToLower(key)
	}

	if cfg.Enabled && cfg.Endpoints.Logs != "" {
		exporter, err := logShipper(ctx, cfg.Endpoints.Logs)
		if err != nil {
			return fmt.Errorf("creating log exporter: %w", err)
		}
		// 2. Create a Resource to add service metadata to logs
		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String(cfg.Service.Name),
				semconv.ServiceVersionKey.String(cfg.Service.Version),
			),
		)
		if err != nil {
			slog.Error("failed to create resource", "error", err)
			return fmt.Errorf("creating log resource: %w", err)
		}

		// 3. Create the OTel Logger Provider
		// Use a BatchProcessor for production use to efficiently send logs in batches.
		// A simple processor can be used for debugging/testing.
		processor := otellog.NewBatchProcessor(exporter)
		loggerProvider := otellog.NewLoggerProvider(
			otellog.WithResource(res),
			otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
			otellog.WithProcessor(processor),
		)
		//defer func() {
		//	// Ensure the logger provider is shut down before exiting the application
		//	if err := loggerProvider.Shutdown(ctx); err != nil {
		//		slog.Error("failed to shutdown logger provider", "error", err)
		//	}
		//}()

		global.SetLoggerProvider(loggerProvider)

		// Set the default slog logger to use the OTel bridge handler
		slog.SetDefault(
			otelslog.NewLogger(appName, otelslog.WithLoggerProvider(loggerProvider)),
		)
		return nil
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

func logShipper(ctx context.Context, logsEndpoint string) (*otlploggrpc.Exporter, error) {
	exporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(logsEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("creating otlp log exporter: %w", err)
	}

	return exporter, nil
}
