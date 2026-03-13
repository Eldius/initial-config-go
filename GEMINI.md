# Project Overview: initial-config-go

`initial-config-go` is a reusable Go library designed to simplify application bootstrapping. It provides a unified interface for configuration management (Viper), structured logging (slog), and OpenTelemetry (OTEL) instrumentation (traces, metrics, and logs).

## Main Technologies
- **Language**: Go 1.26
- **Configuration**: [Viper](https://github.com/spf13/viper)
- **Logging**: Go standard library [log/slog](https://pkg.go.dev/log/slog)
- **Telemetry**: [OpenTelemetry Go SDK](https://go.opentelemetry.io/otel)
- **HTTP Client Instrumentation**: [otelhttp](https://go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp)

## Key Components

- **`setup`**: The core package that orchestrates the initialization of configuration, logging, and telemetry. Use `InitSetup(ctx, appName, opts...)` to initialize.
- **`configs`**: Defines configuration keys and default values in `constants.go`.
- **`logs`**: A wrapper around `slog` that provides a context-aware `Logger` interface with automatic trace/span ID inclusion. Use `logs.NewLogger(ctx)` to create loggers.
- **`telemetry`**: Helpers for setting up OpenTelemetry tracer, meter, and log providers.
- **`http/client`**: An instrumented HTTP client (`NewClient()`) that supports trace propagation and request/response logging.

## Initialization & Customization

The library is initialized via `setup.InitSetup`. Customization is achieved through `OptionFunc`s:

- `WithDefaultValues(map[string]any)`: Set default configuration values.
- `WithProps(...Prop)`: Set specific properties.
- `WithEnvPrefix(string)`: Set prefix for environment variables (defaults to `APP_`).
- `WithDefaultCfgFileName(string)`: Set the name of the config file (defaults to `config`).
- `WithDefaultCfgFileLocations(...string)`: Set search paths for the config file.
- `WithConfigFileToBeUsed(string)`: Force a specific config file.
- `WithOpenTelemetryOptions(...telemetry.Option)`: Pass options to the telemetry setup.

## Building and Running

Common development tasks are managed via the `Makefile`:

- **Test**: `make test` (Runs tests with coverage)
- **Lint**: `make lint` (Runs `golangci-lint`)
- **Vulnerability Check**: `make vulncheck` (Runs `govulncheck`)
- **Validate**: `make validate` (Runs test, lint, and vulncheck)
- **Benchmark**: `make benchmark` (Runs benchmarks)
- **Example Stack**: `make telemetry-example` (Starts a Grafana LGTM stack and a sample app using Docker Compose)

## Development Conventions

### Coding Style
- Follow standard Go idiomatic practices.
- Use `slog` for all logging.
- Ensure that `context.Context` is passed through for trace propagation.

### Testing
- Tests are located alongside the source code or in specific `_test.go` files (e.g., in the `setup/` directory).
- Use `make test` to verify changes.
- Benchmarks are maintained for performance-critical parts like the `redact_handler`.

### Telemetry
- Always use `InitSetup` to ensure telemetry is correctly initialized.
- Use the `logs.NewLogger(ctx)` to ensure logs are linked to active spans.
- Use the instrumented HTTP client for external service calls to maintain trace continuity.

### Configuration
- New configuration keys should be added to `configs/constants.go`.
- Default values should be added to `configs/constants.go` and `setup/setup.go` if they are core defaults.
- Environment variables use the `APP_` prefix by default (configurable).
- Config files are expected in YAML format.

## Troubleshooting Telemetry

If telemetry is not working:
1. Ensure `telemetry.enabled` is `true`.
2. Check that `telemetry.traces.endpoint` and `telemetry.metrics.endpoint` are correctly set.
3. Use the `telemetry-example` to verify the setup against a known-good stack.
4. Verify that the application name passed to `InitSetup` is not empty.
