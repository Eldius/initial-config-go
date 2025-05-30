package telemetry

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
	"log/slog"
	"time"
)

var (
	ErrTraceExporterInitialization = errors.New("failed to start traces exporter")
)

func tracerProvider(ctx context.Context, cfg OTELConfigs) error {
	l := slog.Default()
	l.Debug(fmt.Sprintf("configuring trace export for '%s'", cfg.Endpoints.Traces))
	fmt.Println(fmt.Sprintf("configuring trace export for '%s'", cfg.Endpoints.Traces))

	conn, err := newGrpcConnection(cfg.Endpoints.Traces)
	if err != nil {
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
	//tracerInstance = provider.Tracer(cfg.Service.Name)

	fmt.Println("trace exporter configured")
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

// NewSpan creates and starts a new trace span with the provided context, trace name, and optional span start options.
func NewSpan(ctx context.Context, traceName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, traceName, opts...)
}

type TracingIDs struct {
	TraceID string
	SpanID  string
}
type tracingDataKey struct{}

func GetSpanDataFromContext(ctx context.Context) TracingIDs {
	traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()
	spanID := trace.SpanFromContext(ctx).SpanContext().SpanID().String()
	return TracingIDs{
		TraceID: traceID,
		SpanID:  spanID,
	}
}
