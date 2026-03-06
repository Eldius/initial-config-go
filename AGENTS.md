# Specialized Agents for initial-config-go

This project defines specific domains that benefit from specialized expertise. When interacting with an AI agent or delegating tasks, consider these specialized roles:

## 1. Core Architect
**Focus**: System initialization, configuration management, and library orchestration.

- **Responsibilities**:
    - Managing the `setup` package and `InitSetup` logic.
    - Configuring [Viper](https://github.com/spf13/viper) for file, environment, and default sources.
    - Maintaining the library's public API and extension points (`OptionFunc`).
- **Key Files**:
    - `setup/setup.go`
    - `setup/setup_helpers.go`
    - `configs/configs.go`
    - `configs/constants.go`

## 2. Observability Specialist
**Focus**: OpenTelemetry (OTEL) integration, instrumentation, and exporter configuration.

- **Responsibilities**:
    - Configuring Tracer, Meter, and Logger providers.
    - Managing OTLP gRPC exporters and connections.
    - Implementing runtime instrumentation and resource attributes.
    - Maintaining the `telemetry-example` infrastructure (Docker Compose, Grafana LGTM).
- **Key Files**:
    - `telemetry/setup.go`
    - `telemetry/tracer.go`
    - `telemetry/meter.go`
    - `setup/telemetry.go`
    - `docker-compose-telemetry.yml`

## 3. Log Security Specialist
**Focus**: Structured logging, attribute redaction, and log handler implementation.

- **Responsibilities**:
    - Maintaining the `logs` package and the `Logger` interface.
    - Implementing and optimizing the `RedactHandler` for sensitive data protection.
    - Integrating `slog` with OpenTelemetry via bridge handlers.
- **Key Files**:
    - `logs/logger.go`
    - `setup/logs.go`
    - `setup/redact_handler.go`
    - `setup/redact_handler_test.go`

## 4. Networking Expert
**Focus**: Instrumented HTTP clients, middleware, and trace propagation.

- **Responsibilities**:
    - Maintaining the `http/client` package.
    - Implementing `RoundTripper` logic for request/response logging.
    - Ensuring trace context propagation across service boundaries using `otelhttp`.
- **Key Files**:
    - `http/client/client.go`
    - `http/client/logging.go`

## Task Delegation Guide

| Task Category | Recommended Agent |
|---------------|-------------------|
| Adding a new configuration source | Core Architect |
| Fixing a trace propagation issue | Networking Expert |
| Implementing a new metrics exporter | Observability Specialist |
| Adding keys to the redaction list | Log Security Specialist |
| Troubleshooting Docker Compose stack | Observability Specialist |
| Benchmarking log redaction performance | Log Security Specialist |
