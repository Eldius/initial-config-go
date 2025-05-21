package telemetry

type telemetryCfg struct {
	Endpoint string
	Enabled  bool
}

func (t *telemetryCfg) IsEnabled() bool {
	return t.Enabled && t.Endpoint != ""
}

func newDefaultCfg() *telemetryCfg {
	return &telemetryCfg{
		Endpoint: "",
		Enabled:  false,
	}
}

// Option defines a telemetry configuration option.
type Option func(*telemetryCfg)

// SetupMetricsProvider sets up the metrics provider
func SetupMetricsProvider(opts ...Option) error {
	cfg := newDefaultCfg()
	if !cfg.IsEnabled() {
		return nil
	}
	return nil
}
