package telemetry

type OTELConfigs struct {
	Service struct {
		Name        string
		Version     string
		Environment string
	}
	Endpoints struct {
		Traces  string
		Metrics string
		Logs    string
	}
	Enabled bool
}

func (t *OTELConfigs) IsEnabled() bool {
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

// WithLogsEndpoint sets the endpoint for the logs exporter.
func WithLogsEndpoint(endpoint string) Option {
	return func(cfg *OTELConfigs) {
		cfg.Endpoints.Logs = endpoint
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
