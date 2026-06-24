package telemetry

import (
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type ProviderSet struct {
	MeterProvider  *metric.MeterProvider
	TracerProvider *tracesdk.TracerProvider
	LoggerProvider *otellog.LoggerProvider
}

var currentProviders *ProviderSet

func SetProviderSet(ps *ProviderSet) {
	currentProviders = ps
}

func GetProviderSet() *ProviderSet {
	return currentProviders
}

func SetTracerProvider(tp *tracesdk.TracerProvider) {
	p := getOrCreateProviderSet()
	p.TracerProvider = tp
}

func SetMeterProvider(mp *metric.MeterProvider) {
	p := getOrCreateProviderSet()
	p.MeterProvider = mp
}

func SetLoggerProvider(lp *otellog.LoggerProvider) {
	p := getOrCreateProviderSet()
	p.LoggerProvider = lp
}

func getOrCreateProviderSet() *ProviderSet {
	if currentProviders == nil {
		currentProviders = &ProviderSet{}
	}
	return currentProviders
}
