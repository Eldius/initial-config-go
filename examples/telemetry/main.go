package main

import (
	"context"
	"fmt"
	"github.com/eldius/initial-config-go/setup"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"math/rand/v2"
	"time"
)

/*
Programmatically configuring telemetry components.
*/
func main() {
	if err := setup.InitSetup(
		"telemetry-example-app",
		setup.WithDefaultCfgFileLocations("./examples/telemetry/", "."),
		setup.WithEnvPrefix("telemetry"),
		setup.WithDefaultCfgFileName("config"),
		setup.WithOpenTelemetryOptions(
			telemetry.WithTraceEndpoint("otlp:55689"),
			telemetry.WithMetricEndpoint("otlp:55690"),
			telemetry.WithOtelEnabled(true),
			telemetry.WithService("telemetry-example-app", "1.0.0", "dev"),
		),
		setup.WithDefaultValues(map[string]any{}),
	); err != nil {
		panic(err)
	}

	counter, err := telemetry.GetMeter("test-meter").Int64Counter("test-counter")
	if err != nil {
		panic(err)
	}
	for {
		iterate(counter)
	}
}

func iterate(counter metric.Int64Counter) {
	ctx := context.Background()
	ctx, span := telemetry.NewSpan(ctx, "test-span", trace.WithSpanKind(trace.SpanKindInternal))
	defer func() {
		slog.Debug("span ended")
		span.End()
		counter.Add(ctx, 1)
	}()
	timeToSleep := time.Duration(rand.IntN(5) * int(time.Second))
	slog.Debug(fmt.Sprintf("sleeping for %s", timeToSleep.String()))
	time.Sleep(timeToSleep)
	slog.Debug("done sleeping.")
}
