package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// GetMeter returns a meter instance
func GetMeter(meterName string, opts ...metric.MeterOption) metric.Meter {
	opts = append(opts, metric.WithInstrumentationAttributes(getDefaultTelemetryAttributes(cfgCache)...))
	return otel.GetMeterProvider().Meter(meterName, opts...)
}
