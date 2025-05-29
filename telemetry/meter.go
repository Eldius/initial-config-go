package telemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"google.golang.org/grpc/encoding/gzip"
	"log/slog"
	"time"
)

// meterProvider sets up the metrics provider
func meterProvider(ctx context.Context, cfg OTELConfigs) error {
	l := slog.Default().With(
		slog.String("exporter_endpoint", cfg.Endpoints.Metrics),
	)
	l.Debug("configuring metric exporter")

	var opts []otlpmetricgrpc.Option

	conn, err := newGrpcConnection(cfg.Endpoints.Metrics)
	if err != nil {
		return fmt.Errorf("creating metric exporter grpc client: %w", err)
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
		return fmt.Errorf("creating metric exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second))),
		sdkmetric.WithResource(defaultResources(cfg)))

	// set global meter provider
	otel.SetMeterProvider(provider)
	return nil
}

func defaultResources(cfg OTELConfigs) *resource.Resource {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.Service.Name),
		semconv.ServiceVersionKey.String(cfg.Service.Version),
		attribute.String("environment", cfg.Service.Environment),
	)
	return res
}

// GetMeter returns a meter instance
func GetMeter(meterName string, opts ...metric.MeterOption) metric.Meter {
	return otel.GetMeterProvider().Meter(meterName, opts...)
}
