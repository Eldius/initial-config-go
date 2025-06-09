package telemetry

import (
	"fmt"
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

func NewDefaultCfg() *OTELConfigs {
	return &OTELConfigs{}
}

// Option defines a telemetry configuration option.
type Option func(*OTELConfigs)

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
