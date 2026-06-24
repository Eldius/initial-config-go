# initial-config-go

![CI/CD](https://github.com/eldius/initial-config-go/actions/workflows/ci.yml/badge.svg)


`initial-config-go` is a reusable Go library designed to simplify application bootstrapping by providing a unified way to handle configuration, structured logging, and OpenTelemetry instrumentation.

## Features

- **Configuration**: Powered by [Viper](https://github.com/spf13/viper). Supports YAML files, environment variables, and default values.
- **Structured Logging**: Built on top of Go's standard `log/slog`. Supports:
    - JSON and Text formats.
    - Output to stdout, files, or both.
    - Attribute redaction for sensitive data.
    - Automatic trace and span ID inclusion when OpenTelemetry is enabled.
    - Log shipping to OpenTelemetry collectors.
- **OpenTelemetry**: Integrated support for Traces, Metrics, and Logs.
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

> **Note:** `WithEnvPrefix("MYAPP")` lowercases the prefix to `myapp` internally, but Viper uppercases it for environment matching — use `MYAPP_*` env vars (e.g. `MYAPP_LOG_FORMAT=json`).

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

> **Note:** At least one log output must be configured — either stdout (`log.output_to_stdout: true`), a file (`log.output_to_file: /path/to/log.log`), or an OpenTelemetry logs endpoint. Otherwise `InitSetup` will return an error.

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
Sensitive keys can be automatically redacted:

```go
setup.InitSetup(ctx, "my-app",
    setup.WithDefaultValues(map[string]any{
        "log.redacted_keys": []string{"password", "api_key"},
    }),
)

// Logs containing "password" or "api_key" will have their values replaced with "***"
```

## OpenTelemetry

To enable telemetry, provide the endpoints and enable the flag. When telemetry is enabled, the library automatically:

- Sets up **Tracer**, **Meter**, and **Logger** providers.
- Starts **runtime instrumentation** (memory stats, goroutines, etc.).
- Configures the default HTTP client for trace propagation.

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

## HTTP Client Helper

The library provides an instrumented HTTP client:

```go
import "github.com/eldius/initial-config-go/http/client"

func main() {
    c := client.NewClient()
    resp, err := c.Get("https://api.example.com")
    // ...
}
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

### Combined Example

```go
import (
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
- `make lint`: Run `golangci-lint`.
- `make vulncheck`: Run `govulncheck`.
- `make validate`: Run all the above.
- `make benchmark`: Run benchmarks.
- `make telemetry-example`: Run a full OTEL stack (Grafana LGTM) with a sample app using Docker Compose.

### Local Telemetry Stack
To try the telemetry integration locally:
```bash
make telemetry-example
```
This will start:
- **Grafana**: `http://localhost:3000` (Login: `admin`/`admin`)
- **Prometheus**: `http://localhost:9090`
- **Loki**: `http://localhost:3100`
- **Tempo**: `http://localhost:3200`
- **Sample App**: Automatically sending traces, metrics, and logs.

## License
Licensed under [GPL-3.0](LICENSE).
