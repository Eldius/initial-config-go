package telemetry

import "context"

type OTELConfigs struct {
	Service struct {
		Name        string
		Version     string
		Environment string
	}
	Endpoint string
	Enabled  bool
}

func (t *OTELConfigs) IsEnabled() bool {
	return t.Enabled && t.Endpoint != ""
}

func newDefaultCfg() *OTELConfigs {
	return &OTELConfigs{
		Endpoint: "",
		Enabled:  false,
	}
}

// Option defines a telemetry configuration option.
type Option func(*OTELConfigs)

// InitTelemetry initializes telemetry configuration
func InitTelemetry(ctx context.Context, telemetryOpts ...Option) error {
	cfg := newDefaultCfg()
	for _, opt := range telemetryOpts {
		opt(cfg)
	}
	if !cfg.IsEnabled() {
		return nil
	}

	if err := meterProvider(ctx, *cfg); err != nil {
		return err
	}
	if err := tracerProvider(ctx, *cfg); err != nil {
		return err
	}
	return nil
}

// WithEndpoint sets the endpoint for the telemetry exporter.
func WithEndpoint(endpoint string) Option {
	return func(cfg *OTELConfigs) {
		cfg.Endpoint = endpoint
	}
}

// WithEnabled enables or disables telemetry.
func WithEnabled(enabled bool) Option {
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
