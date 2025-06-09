package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// GetMeter returns a meter instance
func GetMeter(meterName string, opts ...metric.MeterOption) metric.Meter {
	return otel.GetMeterProvider().Meter(meterName, opts...)
}
