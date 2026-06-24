package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/eldius/initial-config-go/configs"
	"github.com/eldius/initial-config-go/setup"
	"github.com/eldius/initial-config-go/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var AppVersion string

/*
Telemetry configured via config.yaml.
Override with environment variables (prefix: APP):
  APP_TELEMETRY_ENABLED=true
  APP_TELEMETRY_TRACES_ENDPOINT=localhost:4317
  APP_TELEMETRY_METRICS_ENDPOINT=localhost:4317
  APP_TELEMETRY_LOGS_ENDPOINT=localhost:4317
*/
func main() {
	if err := setup.InitSetup(
		context.Background(),
		"telemetry-example-app",
		setup.WithDefaultCfgFileLocations("."),
		setup.WithEnvPrefix("APP"),
		setup.WithDefaultValues(map[string]any{
			"telemetry.enabled":          false,
			"telemetry.traces.endpoint":  "",
			"telemetry.metrics.endpoint": "",
			"telemetry.logs.endpoint":    "",
		}),
	); err != nil {
		panic(err)
	}

	if !configs.GetTelemetryEnabled() {
		slog.Warn("telemetry is not enabled — set APP_TELEMETRY_ENABLED=true and configure endpoints")
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
		span.End()
		counter.Add(ctx, 1)
	}()
	timeToSleep := time.Duration(rand.IntN(5)+1) * time.Second
	slog.Debug(fmt.Sprintf("sleeping for %s", timeToSleep.String()))
	time.Sleep(timeToSleep)
	slog.Debug("done sleeping.")
}
