package setup

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/httpclient"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.32.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	// ErrTracesInitialization is returned when trace provider initialization fails.
	ErrTracesInitialization = errors.New("initializing meter")
	// ErrTracesConnectionInitialization is returned when trace gRPC connection fails.
	ErrTracesConnectionInitialization = errors.New("initializing metrics connection")
	// ErrTracesExporterInitialization is returned when trace exporter setup fails.
	ErrTracesExporterInitialization = errors.New("initializing metric exporter")
)

// InitTelemetry initializes telemetry configuration
func InitTelemetry(ctx context.Context, telemetryOpts ...telemetry.Option) error {
	cfg := telemetry.NewDefaultCfg()
	for _, opt := range telemetryOpts {
		opt(cfg)
	}

	if cfg.Endpoints.Traces == "" {
		cfg.Endpoints.Traces = configs.GetTraceBackendEndpoint()
	}
	if cfg.Endpoints.Metrics == "" {
		cfg.Endpoints.Metrics = configs.GetMetricsBackendEndpoint()
	}

	if !cfg.Enabled {
		cfg.Enabled = configs.GetTelemetryEnabled()
	}

	l := slog.With(
		"component", "telemetry",
		"enabled", cfg.IsEnabled())

	l.Debug("configuring telemetry")

	if !cfg.IsEnabled() {
		return nil
	}

	otel.SetLogger(logr.FromSlogHandler(slog.Default().Handler()))

	if err := meterProvider(ctx, *cfg); err != nil {
		return err
	}
	if err := tracerProvider(ctx, *cfg); err != nil {
		return err
	}

	http.DefaultClient = httpclient.NewHTTPClient()

	// Start the runtime instrumentation
	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(5 * time.Second),
	); err != nil {
		return fmt.Errorf("failed to start runtime instrumentation: %w", err)
	}
	return nil
}

func tracerProvider(ctx context.Context, cfg telemetry.OTELConfigs) error {
	l := slog.Default()
	l.Debug(fmt.Sprintf("configuring trace export for '%s'", cfg.Endpoints.Traces))

	conn, err := newGrpcConnection(cfg.Endpoints.Traces)
	if err != nil {
		err = fmt.Errorf("%w: %w: %w", ErrTracesInitialization, ErrTracesConnectionInitialization, err)
		l.With("error", err).Error("failed to create gRPC connection")
		return err
	}

	l.With(
		"tracer_grpc_conn_status",
		conn.GetState().String(),
	).Debug("gRPC connection to collector established")

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithCompressor(gzip.Name),
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlptracegrpc.WithTimeout(10*time.Second))
	if err != nil {
		err = fmt.Errorf("%w: %w: %w", ErrTracesInitialization, ErrTracesExporterInitialization, err)
		l.With("error", err).Error("failed to setup exporter")
		return err
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.Service.Name),
		semconv.ServiceVersionKey.String(cfg.Service.Version),
		attribute.String("environment", cfg.Service.Environment),
	)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	// set global tracer provider & text propagators
	otel.SetTracerProvider(provider)

	return nil
}

var (
	// ErrMeterInitialization is returned when meter provider initialization fails.
	ErrMeterInitialization = errors.New("initializing meter")
	// ErrMetricsConnectionInitialization is returned when metrics gRPC connection fails.
	ErrMetricsConnectionInitialization = errors.New("initializing metrics connection")
	// ErrMetricsExporterInitialization is returned when metrics exporter setup fails.
	ErrMetricsExporterInitialization = errors.New("initializing metric exporter")
)

// meterProvider sets up the metrics provider
func meterProvider(ctx context.Context, cfg telemetry.OTELConfigs) error {
	l := slog.Default().With(
		slog.String("exporter_endpoint", cfg.Endpoints.Metrics),
	)
	l.Debug("configuring metric exporter")

	var opts []otlpmetricgrpc.Option

	conn, err := newGrpcConnection(cfg.Endpoints.Metrics)
	if err != nil {
		return fmt.Errorf("%w: %w: %w", ErrMeterInitialization, ErrMetricsConnectionInitialization, err)
	}

	opts = append(opts,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithCompressor(gzip.Name),
		otlpmetricgrpc.WithGRPCConn(conn),
		otlpmetricgrpc.WithTimeout(10*time.Second))

	exporter, err := otlpmetricgrpc.New(
		ctx,
		opts...,
	)
	if err != nil {
		return fmt.Errorf("%w: %w: %w", ErrMeterInitialization, ErrMetricsExporterInitialization, err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(defaultResources(cfg)))

	// set global meter provider
	otel.SetMeterProvider(provider)

	return nil
}

func defaultResources(cfg telemetry.OTELConfigs) *resource.Resource {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.Service.Name),
		semconv.ServiceVersionKey.String(cfg.Service.Version),
		attribute.String("environment", cfg.Service.Environment),
	)
	return res
}

func newGrpcConnection(endpoint string) (*grpc.ClientConn, error) {
	return grpc.NewClient(
		endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
