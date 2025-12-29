# initial-config-go

Reusable Go starter library for application configuration, structured logging, and optional OpenTelemetry instrumentation. It helps you bootstrap new services with:
- Viper-based configuration with sane defaults and flexible sources (file + env)
- slog-based structured logging with redaction support
- Optional OpenTelemetry (traces and metrics) setup
- An instrumented HTTP client with request/response logging and OTEL propagation

This repo also provides an example app and Docker Compose to try telemetry locally.


## Stack
- Language: Go (module: `github.com/eldius/initial-config-go`, go 1.25)
- Package manager: Go modules (`go.mod`/`go.sum`)
- Key libraries/frameworks:
  - Configuration: `github.com/spf13/viper`
  - Logging: standard library `log/slog`
  - Telemetry: `go.opentelemetry.io/otel` and related exporters
  - CLI helper present in dependencies: `github.com/spf13/cobra` (not used directly in this repo’s code at the moment)


## Requirements
- Go 1.25+
- Optional developer tools for Makefile targets:
  - golangci-lint (for `make lint`) — TODO: link/install instructions
  - govulncheck (for `make vulncheck`) — https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- Docker and Docker Compose (only if you want to run the telemetry example)


## Installation
Add to your project:

- Latest
  go get github.com/eldius/initial-config-go@latest

- Or pin a version (recommended for reproducibility)
  go get github.com/eldius/initial-config-go@<version>

TODO: publish and document version tags for stable releases.


## Quick start
Initialize config, logging, and telemetry at app startup:

Example (minimal):

  package main

  import (
      "github.com/eldius/initial-config-go/setup"
  )

  func main() {
      if err := setup.InitSetup(
          "my-app", // app name used for config locations and telemetry defaults
      ); err != nil {
          panic(err)
      }

      // your app code here
  }

Example (customization similar to the provided telemetry example):

  package main

  import (
      "github.com/eldius/initial-config-go/setup"
      "github.com/eldius/initial-config-go/telemetry"
  )

  func main() {
      if err := setup.InitSetup(
          "my-app",
          setup.WithDefaultCfgFileLocations("./configs", "."),
          setup.WithEnvPrefix("myapp"),
          setup.WithDefaultCfgFileName("config"),
          setup.WithOpenTelemetryOptions(
              telemetry.WithService("my-app", "1.0.0", "dev"),
              // telemetry.WithTraceEndpoint("otlp:4317"),
              // telemetry.WithMetricEndpoint("otlp:4317"),
              // telemetry.WithOtelEnabled(true),
          ),
          setup.WithDefaultValues(map[string]any{}),
      ); err != nil {
          panic(err)
      }
  }


## Configuration

### Overview
Configuration is powered by Viper and supports multiple sources in the following precedence order:
1. Explicit values set via `setup.WithProps()` or `setup.WithDefaultValues()`
2. Environment variables (with configurable prefix)
3. YAML configuration file
4. Default values

### Configuration Sources

#### Default Search Locations
Unless overridden with `setup.WithDefaultCfgFileLocations()`:
- `~/.<appName>/` - User home directory with app-specific folder
- `~/` - User home directory
- `.` - Current working directory

Default config file name: `config.yaml`

#### Environment Variables
- Prefix defaults to `APP` (customize via `setup.WithEnvPrefix("<prefix>")`)
- Configuration keys use dot notation (e.g., `log.level`)
- Environment variables use underscores and are uppercase (e.g., `APP_LOG_LEVEL`)
- Conversion: dots (`.`) → underscores (`_`), lowercase → uppercase

**Example**: `log.level` with prefix `app` → `APP_LOG_LEVEL`

### Configuration Keys

#### Logging Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `log.format` | string | `json` | Log output format: `json` or `text` |
| `log.level` | string | `info` | Log level: `info`, `debug`, `warn`, or `error` |
| `log.output_to_file` | string | `""` | Path to log output file (empty disables file logging) |
| `log.output_to_stdout` | boolean | `false` | Enable logging to stdout |
| `log.redacted_keys` | []string | `[]` | List of attribute keys to redact from logs (e.g., `password`, `token`) |

**Note**: At least one of `log.output_to_file` or `log.output_to_stdout` must be enabled.

#### Telemetry Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `telemetry.enabled` | boolean | `false` | Enable OpenTelemetry instrumentation |
| `telemetry.traces.endpoint` | string | `""` | OTLP traces backend endpoint (e.g., `localhost:4317`) |
| `telemetry.metrics.endpoint` | string | `""` | OTLP metrics backend endpoint (e.g., `localhost:4317`) |

**Note**: Telemetry exporters only initialize when enabled and endpoints are configured.

### Configuration Examples

#### YAML Configuration File
```yaml
log:
  level: debug
  format: json
  output_to_file: app.log
  output_to_stdout: true
  redacted_keys:
    - password
    - token
    - api_key

telemetry:
  enabled: true
  traces:
    endpoint: localhost:4317
  metrics:
    endpoint: localhost:4317
```

#### Environment Variables
```bash
# Logging configuration
export APP_LOG_FORMAT=json
export APP_LOG_LEVEL=debug
export APP_LOG_OUTPUT_TO_FILE=app.log
export APP_LOG_OUTPUT_TO_STDOUT=true
export APP_LOG_REDACTED_KEYS=password,token,api_key

# Telemetry configuration
export APP_TELEMETRY_ENABLED=true
export APP_TELEMETRY_TRACES_ENDPOINT=localhost:4317
export APP_TELEMETRY_METRICS_ENDPOINT=localhost:4317
```

#### Programmatic Configuration
```go
import (
    "github.com/eldius/initial-config-go/setup"
    "github.com/eldius/initial-config-go/configs"
)

func main() {
    err := setup.InitSetup(
        "my-app",
        setup.WithDefaultValues(map[string]any{
            configs.LogLevelKey:          configs.LogLevelDEBUG,
            configs.LogFormatKey:         configs.LogFormatJSON,
            configs.LogOutputFileKey:     "app.log",
            configs.LogOutputToStdoutKey: true,
            configs.LogKeysToRedactKey:   []string{"password", "token"},
        }),
        setup.WithEnvPrefix("myapp"), // Use MYAPP_* env vars
    )
    if err != nil {
        panic(err)
    }
}
```

### Helper Constants and Default Maps

The `configs` package provides helper constants and pre-configured default value maps:

```go
// Use defaults with file logging enabled
setup.WithDefaultValues(configs.DefaultConfigValuesLogFileMap)

// Use defaults with stdout logging enabled
setup.WithDefaultValues(configs.DefaultConfigValuesLogStdoutMap)

// Use defaults with both file and stdout logging enabled
setup.WithDefaultValues(configs.DefaultConfigValuesLogAllMap)
```


## HTTP client helper
Package `httpclient` provides:
- NewHTTPClient(): returns an *http.Client with:
  - OpenTelemetry transport when a tracer provider is present
  - Request/response logging via a custom RoundTripper
- NewClient(): wraps http.Client with convenience methods and slog logging of outcomes


## Example app with telemetry
An example application is available under `examples/telemetry`.

- Compose stack to run Grafana OTEL LGTM and the example app:
  make telemetry-example

- Tear down:
  make telemetry-example-down

The compose file exposes:
- Grafana UI at http://localhost:3000
- OTLP gRPC at 4317 and HTTP at 4318

Dockerfiles used: `dockerfiles/telemetry.Dockerfile`
Healthcheck script: `scripts/otel_lgtm_healthcheck.sh`


## Makefile scripts
- make test — Run tests with coverage
- make lint — Run golangci-lint (requires installation)
- make vulncheck — Run govulncheck
- make validate — test + lint + vulncheck
- make benchmark — Run benchmarks across packages
- make telemetry-example — Build and run the telemetry compose example
- make telemetry-example-down — Stop the telemetry compose example
- make release — Tag and push next version (assumes git tags). TODO: document release flow.


## Tests
Tests live mainly under `./setup`. Run all tests:

- With Go:
  go test ./...

- Or via Makefile:
  make test


## Project structure
High-level overview:

- configs/ — config keys and default values
- httpclient/ — http client with OTEL + logging
- logs/ — logger wrapper around slog with context and trace/span fields
- setup/ — initialization of Viper, logging, and telemetry
- telemetry/ — telemetry setup, tracer and meter helpers
- examples/telemetry/ — example app and config
- dockerfiles/ and docker-compose-telemetry.yml — example runtime
- scripts/ — helper shell scripts
- Makefile — common developer tasks
- LICENSE — project license (GPL-3.0)


## License
This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the LICENSE file for details.


## Previous README content (retained and updated)
Default configuration keys (illustrative YAML):

```yaml
log:
  level: info
  format: json
  output_to_file: execution.log
  output_to_stdout: false
  redacted_keys: []
telemetry:
  enabled: false
  traces:
    endpoint: ""
  metrics:
    endpoint: ""
```


## TODOs
- Add installation instructions for golangci-lint (or provide config)
- Publish and document versioned releases and CHANGELOG
- Add more usage examples (e.g., logging with context, redact handler, HTTP client usage)
- Contribution guidelines and code of conduct
