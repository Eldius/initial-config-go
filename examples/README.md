# Examples

This directory contains usage examples for the `initial-config-go` library.

| Example | What it shows |
|---------|---------------|
| [basic](./basic) | Minimal InitSetup with default settings and basic slog logging |
| [custom-config](./custom-config) | `WithDefaultValues`, `WithProps`, custom env prefix, config file overrides |
| [http-client](./http-client) | Instrumented HTTP client with request/response logging |
| [http-server](./http-server) | HTTP server with telemetry, request logging, and API key auth middleware |
| [redaction](./redaction) | Sensitive key redaction via `log.redacted_keys` |
| [telemetry](./telemetry) | Full OpenTelemetry (traces, metrics, logs) with config file |

## Running

Each example has a `Makefile`. Run from the example directory:

```bash
make run
```

Override config values with environment variables using the configured prefix (e.g. `APP_LOG_LEVEL=debug make run`).
