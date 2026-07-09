package setup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/contrib/bridges/otelslog"

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

	fmt.Println("setting up logs")

	cfg := telemetry.NewDefaultCfg()

	for _, o := range options.OpenTelemetryOptions {
		o(cfg)
	}

	if !cfg.Enabled {
		cfg.Enabled = configs.GetTelemetryEnabled()
	}

	if cfg.Endpoints.Logs == "" {
		cfg.Endpoints.Logs = configs.GetLogsBackendEndpoint()
	}

	if !stdout && logOutputFile == "" {
		return fmt.Errorf("%w: logOutputFile: %s / stdout: %v", ErrInvalidLogOutputConfig, logOutputFile, stdout)
	}

	for i, key := range keysToRedact {
		keysToRedact[i] = strings.ToLower(key)
	}

	if cfg.Enabled && cfg.Endpoints.Logs != "" {
		fmt.Println("setting up logs with OpenTelemetry")
		exporter, err := logShipper(ctx, cfg.Endpoints.Logs)
		if err != nil {
			return fmt.Errorf("creating log exporter: %w", err)
		}
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

		loggerProvider := otellog.NewLoggerProvider(
			otellog.WithResource(res),
			otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		)

		global.SetLoggerProvider(loggerProvider)

		telemetry.SetLoggerProvider(loggerProvider)

		handler := logs.NewRedactHandler(
			otelslog.NewHandler(
				appName,
				otelslog.WithLoggerProvider(loggerProvider),
			),
			keysToRedact,
		)
		// Set the default slog logger to use the OTel bridge handler
		slog.SetDefault(slog.New(handler))
		return nil
	}
	writer, err := logs.GetWriter(logOutputFile, stdout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLogOutputConfig, err)
	}
	h, err := logs.LogHandler(format, level, writer, keysToRedact...)
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
