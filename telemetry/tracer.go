package telemetry

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

func tracerProvider(ctx context.Context, cfg OTELConfigs) error {
	l := slog.Default()
	l.Debug(fmt.Sprintf("configuring trace export for '%s'", cfg.Endpoint))

	var err error
	conn, err := grpc.NewClient(
		cfg.Endpoint,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		l.With("error", err).Error("failed to create gRPC connection to collector")
		panic(err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		l.With("error", err).Error("failed to setup exporter")
		panic(err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.Service.Name),
		semconv.ServiceVersionKey.String(cfg.Service.Version),
		attribute.String("environment", cfg.Service.Environment),
	)

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
	//tracerInstance = provider.Tracer(cfg.Service.Name)

	return nil
}

// GetTracer returns a tracer instance.
func GetTracer(tracerName string, opts ...trace.TracerOption) trace.Tracer {
	return otel.GetTracerProvider().Tracer(tracerName, opts...)
}

// GetCurrentSpan returns a span instance.
func GetCurrentSpan(ctx context.Context, tracerName string, opts ...trace.TracerOption) trace.Span {
	return trace.SpanFromContext(ctx)
}
