# initial-config-go

![CI/CD](https://github.com/eldius/initial-config-go/actions/workflows/ci.yml/badge.svg)


`initial-config-go` is a reusable Go library designed to simplify application bootstrapping by providing a unified way to handle configuration, structured logging, and OpenTelemetry instrumentation.

## Features

- **Configuration**: Powered by [Viper](https://github.com/spf13/viper). Supports YAML files, environment variables, and default values.
- **Structured Logging**: Built on top of Go's standard `log/slog`. Supports:
    - JSON and Text formats.
    - Output to stdout, files, or both.
    - Attribute redaction for sensitive data.
    - Log shipping to OpenTelemetry collectors.
- **OpenTelemetry**: Integrated support for Traces, Metrics, and Logs, including runtime instrumentation and instrumented SQL connections (`telemetry.GetDB` / `telemetry.GetSqlxDB` via `otelsql`).
- **HTTP Client**: Instrumented HTTP client with automatic trace propagation and request/response logging.
- **HTTP Server**: Middleware for request/response logging, OpenTelemetry instrumentation, and API key authentication.

## Installation

```bash
go get github.com/eldius/initial-config-go
```

## Quick Start

Initialize the library at the beginning of your `main` function:

```go
package main

import (
	"context"
	"github.com/eldius/initial-config-go/setup"
	"log/slog"
)

func main() {
	ctx := context.Background()
	
	// Initialize configuration, logging, and telemetry
	if err := setup.InitSetup(ctx, "my-app"); err != nil {
		panic(err)
	}

	slog.Info("Application started!")
}
```

## Configuration

The library uses a hierarchical configuration approach (Viper precedence):

1. **Defaults** (lowest) — set via `WithDefaultValues` or `WithProps`.
2. **Config file** — YAML from `~/.<appName>/`, `~/`, or `.`.
3. **Environment variables** (highest) — override everything above.

### Config File Search Locations
The library searches for `<name>.yaml` (default name: `config`) in:
- `~/.<appName>/`
- `~/`
- `.` (current working directory)

### Customizing Initialization

You can customize the setup using `OptionFunc`s:

```go
setup.InitSetup(ctx, "my-app",
    setup.WithEnvPrefix("MYAPP"),
    setup.WithDefaultCfgFileName("settings"),
    setup.WithDefaultCfgFileLocations("./configs"),
    setup.WithDefaultValues(map[string]any{
        "server.port": 8080,
    }),
    setup.WithProps(
        setup.Prop{Key: "custom.key", Value: "custom-value"},
    ),
)
```

> **Note:** `WithEnvPrefix("MYAPP")` lowercases the prefix internally, but Viper uppercases it for environment matching — use `MYAPP_*` env vars (e.g. `MYAPP_LOG_FORMAT=json`). Without `WithEnvPrefix`, the default prefix is `app` → use `APP_*` env vars. Dots in keys become underscores (`log.format` → `APP_LOG_FORMAT`).

### Configuration Keys

When using `InitSetup` the effective defaults are:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `log.format` | string | `text` | `json` or `text` |
| `log.level` | string | `info` | `debug`, `info`, `warn`, `error` |
| `log.output_to_file` | string | `""` | Path to log file (empty to disable) |
| `log.output_to_stdout` | bool | `true` | Enable/disable stdout logging |
| `log.redacted_keys` | []string | `[]` | Keys to redact from logs |
| `telemetry.enabled` | bool | `false` | Enable OpenTelemetry |
| `telemetry.traces.endpoint` | string | `""` | OTLP Traces gRPC endpoint |
| `telemetry.metrics.endpoint` | string | `""` | OTLP Metrics gRPC endpoint |
| `telemetry.logs.endpoint` | string | `""` | OTLP Logs gRPC endpoint |
| `telemetry.debug` | bool | `false` | Write OTel SDK debug logs to `otel.log` |

Endpoints are `host:port` (e.g. `localhost:4317`) — all OTLP exporters use insecure gRPC, so no scheme or TLS. Telemetry only activates when `telemetry.enabled` is `true` **and** at least one endpoint is set.

Config keys use dots as separators (`log.format`). In YAML this maps to nested structure:

```yaml
log:
  format: json
  level: debug
  output_to_stdout: true
telemetry:
  enabled: true
  traces:
    endpoint: "localhost:4317"
```

## Logging

### Structured Logging with Context

> **Note:** At least one of stdout (`log.output_to_stdout: true`) or a file (`log.output_to_file: /path/to/log.log`) must be configured, otherwise `InitSetup` returns an error. An OpenTelemetry logs endpoint does **not** satisfy this check (stdout is on by default, so you usually don't need to care).
>
> When telemetry is enabled **and** `telemetry.logs.endpoint` is set, slog's default handler is replaced by the OTel bridge — stdout/file logging is skipped entirely and logs go to the collector instead.

Use the `logs` package to create loggers that automatically include trace information:

```go
import "github.com/eldius/initial-config-go/logs"

func process(ctx context.Context) {
    log := logs.NewLogger(ctx, logs.KeyValueData{"user_id": 123})
    log.Info("Processing request")
    
    if err := doSomething(); err != nil {
        log.WithError(err).Error("Failed to do something")
    }
}
```

### Redaction
Sensitive keys can be automatically redacted — configured via config file or programmatically:

```go
setup.InitSetup(ctx, "my-app",
    setup.WithDefaultValues(map[string]any{
        "log.redacted_keys": []string{"password", "api_key"},
    }),
)

// Logs containing "password" or "api_key" will have their values replaced with "***"
```

The redaction handler is also available directly in the `logs` package:

```go
import "github.com/eldius/initial-config-go/logs"

handler := logs.NewRedactHandler(slog.NewJSONHandler(w, nil), []string{"password"})
logger := slog.New(handler)
```

## Cobra Integration

If your app uses [Cobra](https://github.com/spf13/cobra), the `setup` package provides `PersistentPreRunE` / `PersistentPostRunE` helpers that handle init, per-command spans, and graceful telemetry shutdown:

```go
rootCmd.PersistentPreRunE = setup.PersistentPreRunE("my-app")
rootCmd.PersistentPostRunE = setup.PersistentPostRunE(2 * time.Second)
```

`PersistentPostRunE` flushes and shuts down the telemetry providers, then sleeps for `waitTime` so the OTel batch processors have time to export before the process exits. `TelemetryShutdown` also closes any open log files.

## OpenTelemetry

To enable telemetry, provide the endpoints and enable the flag. When telemetry is enabled, the library automatically:

- Sets up **Tracer**, **Meter**, and **Logger** providers.
- Starts **runtime instrumentation** (memory stats, goroutines, etc.).

### Through InitSetup (recommended)

```go
import "github.com/eldius/initial-config-go/telemetry"

setup.InitSetup(ctx, "my-app",
    setup.WithOpenTelemetryOptions(
        telemetry.WithOtelEnabled(true),
        telemetry.WithTraceEndpoint("localhost:4317"),
        telemetry.WithMetricEndpoint("localhost:4317"),
        telemetry.WithLogsEndpoint("localhost:4317"),
        telemetry.WithService("my-app", "1.0.0", "production"),
    ),
)
```

### Standalone

The `telemetry` package also exports `InitTelemetry` directly for use outside `InitSetup`:

```go
import "github.com/eldius/initial-config-go/telemetry"

telemetry.InitTelemetry(ctx,
    telemetry.WithOtelEnabled(true),
    telemetry.WithTraceEndpoint("localhost:4317"),
    telemetry.WithService("my-app", "1.0.0", "production"),
)
defer telemetry.TelemetryShutdown(ctx)
```

> **Note:** The `http.DefaultClient` is NOT automatically instrumented. See [HTTP Client Helper](#http-client-helper) for options.

### Instrumented SQL

The `telemetry` package provides `GetDB` / `GetSqlxDB`, which open connections through an `otelsql`-registered driver — queries are automatically traced and metered:

```go
db, err := telemetry.GetSqlxDB("postgres", connStr)
```

### Custom Spans

Use `telemetry.NewSpan` (or `telemetry.GetTracer`) to add spans anywhere:

```go
ctx, span := telemetry.NewSpan(ctx, "my-operation")
defer span.End()
```

## HTTP Client Helper

The library provides an instrumented HTTP client with automatic trace propagation and request/response logging:

```go
import "github.com/eldius/initial-config-go/http/client"

func main() {
    c := client.NewClient()
    resp, err := c.Get("https://api.example.com")
    // ...
}
```

The `http.DefaultClient` is NOT automatically instrumented. To replace it, use `WithInstrumentHTTPClient`:

```go
setup.InitSetup(ctx, "my-app",
    setup.WithInstrumentHTTPClient(true),
)
```

Or use the explicit `NewHTTPClient()` function:

```go
import "github.com/eldius/initial-config-go/http/client"

http.DefaultClient = client.NewHTTPClient()
```

Features:
- Automatic Trace Propagation.
- Request/Response logging.
- Integration with `slog`.

## HTTP Server Middleware

The library provides middleware for instrumenting HTTP servers:

### Logging & Telemetry Middleware

```go
import (
    "net/http"
    "github.com/eldius/initial-config-go/http/server"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /api/health", healthHandler)

    // Wrap the mux with telemetry and logging middleware
    handler := server.TelemetryMiddleware(mux)

    http.ListenAndServe(":8080", handler)
}
```

Features:
- Automatic trace propagation via OpenTelemetry.
- Detailed request/response logging (method, URL, headers, body, status code, duration).
- Span naming based on HTTP method and route pattern.

### Authentication Middleware

```go
import "github.com/eldius/initial-config-go/http/server"

// Single API key authentication
authFunc := server.SingleUserApiKeyAuthenticationFunc(
    "my-secret-key", "", user,
)

// Multi-user API key authentication (bcrypt-hashed keys with HMAC lookup token)
apiKeyMap, _ := server.NewApiKeyMapFromPlainMap(map[string]server.User{
    "key-for-user-a": userA,
    "key-for-user-b": userB,
})
authFunc := server.MultipleUserApiKeyAuthenticationFunc(apiKeyMap, "")

// Protect routes with authentication
protected := server.AuthenticationMiddleware(authFunc)
mux.Handle("GET /api/protected", protected(http.HandlerFunc(handler)))
```

The `headerName` argument (empty string above) defaults to `X-Api-Key`. The authenticated user is available inside handlers via `server.AuthenticatedUserFromContext(r.Context())`. `NewApiKeyMapFromPlainMap` hashes the keys with bcrypt — there is also `NewApiKeyMapFromHashedMap` for pre-hashed keys.

### Combined Example

```go
import (
    "context"
    "net/http"
    "github.com/eldius/initial-config-go/setup"
    "github.com/eldius/initial-config-go/http/server"
)

func main() {
    setup.InitSetup(context.Background(), "my-server")
    mux := http.NewServeMux()
    
    // Public route
    mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })
    
    // Protected route with single API key
    auth := server.AuthenticationMiddleware(
        server.SingleUserApiKeyAuthenticationFunc("my-key", "", myUser{}),
    )
    mux.Handle("GET /api/protected", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("secret data"))
    })))
    
    http.ListenAndServe(":8080", server.TelemetryMiddleware(mux))
}
```

## Development

### Makefile Targets
- `make test`: Run tests with coverage.
- `make lint`: Run `golangci-lint` (default linters, no config file).
- `make vulncheck`: Run `govulncheck`.
- `make validate`: Run all the above.
- `make benchmark`: Run benchmarks.
- `make telemetry-example`: Run a full OTEL stack (Grafana LGTM) with a sample app using Docker Compose.

All targets also exist as [Task](https://taskfile.dev) tasks (`task test`, `task lint`, ...). Both runners call bare `golangci-lint` / `task` — if they're not on your PATH, use `go tool golangci-lint run` / `go tool task ...` (declared as `go.mod` tool dependencies).

### Local Telemetry Stack
To try the telemetry integration locally:
```bash
make telemetry-example
```
This starts the all-in-one [Grafana LGTM](https://github.com/grafana/docker-otel-lgtm) container (Grafana + Prometheus + Loki + Tempo) plus a sample app sending traces, metrics, and logs:
- **Grafana UI**: `http://localhost:3000` (anonymous admin)
- **OTLP endpoints**: `localhost:4317` (gRPC), `localhost:4318` (HTTP)

More runnable demos live in [`examples/`](examples/): basic, custom-config, http-client, http-server, redaction, telemetry, grafana.

## License
Licensed under [GPL-3.0](LICENSE).
