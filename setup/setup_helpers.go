package setup

import (
	"context"
	"time"

	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/telemetry"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PersistentPreRunE returns a Cobra PreRunE function that initializes application setup
// and telemetry tracing for the command execution.
func PersistentPreRunE(appName string, opts ...OptionFunc) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		tracing.start = time.Now()
		if err := InitSetup(appName, opts...); err != nil {
			return err
		}
		ctx := cmd.Context()
		if otel.GetTracerProvider() != nil {
			tracing.ctx, tracing.span = telemetry.NewSpan(ctx, cmd.Name(), trace.WithSpanKind(trace.SpanKindInternal))
			tracing.span.SetAttributes(
				attribute.StringSlice("args", args),
				attribute.StringSlice("aliases", cmd.Aliases),
				attribute.String("called_as", cmd.CalledAs()),
			)
		}
		cmd.SetContext(tracing.ctx)
		log := logs.NewLogger(tracing.ctx, logs.KeyValueData{
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
		if tracing.span != nil {
			tracing.span.End()
		}

		logs.NewLogger(tracing.ctx, logs.KeyValueData{
			"cmd_name":     cmd.Name(),
			"cmd_args":     args,
			"is_recording": tracing.span.IsRecording(),
			"running_time": time.Since(tracing.start).String(),
		}).Debug("stopping trace")

		time.Sleep(waitTime)

		return nil
	}
}

var (
	tracing struct {
		ctx    context.Context
		span   trace.Span
		cancel context.CancelFunc
		start  time.Time
	}
)
