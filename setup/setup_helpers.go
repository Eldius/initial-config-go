package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type tracingContextKey string

const tracingKey tracingContextKey = "tracing"

type tracingData struct {
	span  trace.Span
	start time.Time
}

// PersistentPreRunE returns a Cobra PreRunE function that initializes application setup
// and telemetry tracing for the command execution.
func PersistentPreRunE(appName string, opts ...OptionFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		if err := InitSetup(cmd.Context(), appName, opts...); err != nil {
			return err
		}
		ctx := cmd.Context()
		if otel.GetTracerProvider() != nil {
			spanCtx, span := telemetry.NewSpan(ctx, cmd.Name(), trace.WithSpanKind(trace.SpanKindInternal))
			span.SetAttributes(
				attribute.StringSlice("args", args),
				attribute.StringSlice("aliases", cmd.Aliases),
				attribute.String("called_as", cmd.CalledAs()),
			)
			ctx = spanCtx
			cmd.SetContext(context.WithValue(ctx, tracingKey, &tracingData{span: span, start: start}))
		}
		log := logs.NewLogger(ctx, logs.KeyValueData{
			"cmd_name":  cmd.Name(),
			"cmd_args":  args,
			"called_as": cmd.CalledAs(),
		})

		log.Debug("starting trace")

		return nil
	}
}

// PersistentPostRunE returns a Cobra PostRunE function that ends telemetry spans,
// logs command execution details, and waits for the specified duration before returning.
// The wait time allows telemetry data to be flushed to the backend before the process exits.
func PersistentPostRunE(waitTime time.Duration) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var isTracing bool
		var runningTime time.Duration

		if data, ok := cmd.Context().Value(tracingKey).(*tracingData); ok && data != nil {
			if data.span != nil {
				isTracing = true
				data.span.End()
			}
			runningTime = time.Since(data.start)
		}

		logs.NewLogger(cmd.Context(), logs.KeyValueData{
			"cmd_name":     cmd.Name(),
			"cmd_args":     args,
			"is_recording": isTracing,
			"running_time": runningTime.String(),
		}).Debug("stopping trace")

		if err := telemetry.TelemetryForceFlush(cmd.Context()); err != nil {
			logs.NewLogger(cmd.Context()).WithError(err).Error("failed to force flush telemetry data")
			return fmt.Errorf("force flush telemetry data: %w", err)
		}

		if err := telemetry.TelemetryShutdown(cmd.Context()); err != nil {
			logs.NewLogger(cmd.Context()).WithError(err).Error("failed to shutdown telemetry")
			return fmt.Errorf("shutdown telemetry: %w", err)
		}

		time.Sleep(waitTime)

		return nil
	}
}
