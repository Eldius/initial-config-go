package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/eldius/initial-config-go/logs"
)

func TelemetryForceFlush(ctx context.Context) error {
	ps := GetProviderSet()
	if ps == nil {
		return nil
	}
	if ps.MeterProvider != nil {
		if err := ps.MeterProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("failed to force flush metrics: %w", err)
		}
	}
	if ps.TracerProvider != nil {
		if err := ps.TracerProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("failed to force flush traces: %w", err)
		}
	}
	if ps.LoggerProvider != nil {
		if err := ps.LoggerProvider.ForceFlush(ctx); err != nil {
			return fmt.Errorf("failed to force flush logs: %w", err)
		}
	}
	return nil
}

func TelemetryShutdown(ctx context.Context) error {
	ps := GetProviderSet()
	if ps == nil {
		return logs.CloseLogFiles()
	}
	var errs []error
	if ps.LoggerProvider != nil {
		if err := ps.LoggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger provider: %w", err))
		}
	}
	if ps.MeterProvider != nil {
		if err := ps.MeterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter provider: %w", err))
		}
	}
	if ps.TracerProvider != nil {
		if err := ps.TracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("telemetry shutdown errors: %w", errors.Join(errs...))
	}
	if err := logs.CloseLogFiles(); err != nil {
		return fmt.Errorf("telemetry shutdown errors: %w", err)
	}
	return nil
}
