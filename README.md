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
Configuration is powered by Viper and supports:
- Defaults defined in code
- YAML config file (config name and search paths configurable)
- Environment variables (with configurable prefix and dot-to-underscore mapping)

Default search locations (unless overridden):
- ~/.<appName>
- . (current directory)

Default config file name: `config` with type `yaml`.

Environment variables:
- Env prefix defaults to `app` (can be changed via `setup.WithEnvPrefix("<prefix>")`)
- Dots in keys are replaced with underscores, and the prefix is uppercased
  Example: key `log.level` with prefix `app` => env var `APP_LOG_LEVEL`

Supported keys (see configs/constants.go):
- log.format: one of `json`, `text`
- log.level: one of `info`, `debug`, `warn`, `error`
- log.output_to_file: string path to output file ("" means disabled)
- log.output_to_stdout: boolean (default false)
- log.redacted_keys: string slice of keys to be redacted from logs
- telemetry.enabled: boolean (default false)
- telemetry.traces.endpoint: string (OTLP gRPC/HTTP endpoint)
- telemetry.metrics.endpoint: string (OTLP gRPC/HTTP endpoint)

Example environment variables (with default `app` prefix):
- APP_LOG_FORMAT=json
- APP_LOG_LEVEL=info
- APP_LOG_OUTPUT_TO_FILE=execution.log
- APP_LOG_OUTPUT_TO_STDOUT=true
- APP_LOG_REDACTED_KEYS=token,password  (Viper also supports list in YAML)
- APP_TELEMETRY_ENABLED=true
- APP_TELEMETRY_TRACES_ENDPOINT=otlp:4317
- APP_TELEMETRY_METRICS_ENDPOINT=otlp:4317

Note: Telemetry will only initialize exporters when enabled and endpoints are provided. See telemetry/setup.go for details.


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
