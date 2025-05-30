package telemetry

import (
	"context"
	"fmt"
	"github.com/eldius/initial-config-go/configs"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"log/slog"
	"time"
)

type OTELConfigs struct {
	Service struct {
		Name        string
		Version     string
		Environment string
	}
	Endpoints struct {
		Traces  string
		Metrics string
	}
	Enabled bool
}

func (t *OTELConfigs) IsEnabled() bool {
	fmt.Printf("t.Enabled: %v\nt.Endpoints.Traces: %s\nt.Endpoints.Metrics: %s\n", t.Enabled, t.Endpoints.Traces, t.Endpoints.Metrics)
	return t.Enabled && t.Endpoints.Traces != "" && t.Endpoints.Metrics != ""
}

func newDefaultCfg() *OTELConfigs {
	return &OTELConfigs{}
}

// Option defines a telemetry configuration option.
type Option func(*OTELConfigs)

// InitTelemetry initializes telemetry configuration
func InitTelemetry(ctx context.Context, telemetryOpts ...Option) error {
	cfg := newDefaultCfg()
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

	// Start the runtime instrumentation
	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(5 * time.Second),
	); err != nil {
		return fmt.Errorf("failed to start runtime instrumentation: %w", err)
	}
	return nil
}

// WithTraceEndpoint sets the endpoint for the traces exporter.
func WithTraceEndpoint(endpoint string) Option {
	return func(cfg *OTELConfigs) {
		cfg.Endpoints.Traces = endpoint
	}
}

// WithMetricEndpoint sets the endpoint for the metrics exporter.
func WithMetricEndpoint(endpoint string) Option {
	return func(cfg *OTELConfigs) {
		cfg.Endpoints.Metrics = endpoint
	}
}

// WithOtelEnabled enables or disables telemetry.
func WithOtelEnabled(enabled bool) Option {
	return func(cfg *OTELConfigs) {
		cfg.Enabled = enabled
	}
}

// WithService sets the service name, version and environment.
func WithService(name, version, env string) Option {
	return func(cfg *OTELConfigs) {
		cfg.Service.Name = name
		cfg.Service.Version = version
		cfg.Service.Environment = env
	}
}
