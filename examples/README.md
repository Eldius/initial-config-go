# Examples

This directory contains various usage examples for the `initial-config-go` library.

## Available Examples

### 1. [Basic](./basic)
A minimal example showing how to initialize the library with default settings and perform basic logging.

### 2. [Custom Configuration](./custom-config)
Shows how to:
- Define custom configuration keys and default values.
- Use `WithProps` for manual property setting.
- Set a custom environment variable prefix.
- Overwrite values via environment variables.

### 3. [HTTP Client](./http-client)
Demonstrates the use of the instrumented HTTP client, which provides:
- Automatic request/response logging.
- Trace propagation (when telemetry is enabled).

### 4. [Redaction](./redaction)
Shows how to configure and use the log redaction feature to protect sensitive data (e.g., passwords, API keys) from being leaked into logs.

### 5. [Telemetry](./telemetry)
A comprehensive example showing full OpenTelemetry integration (Traces, Metrics, and Logs) with a Grafana LGTM stack.

## Running the Examples

Each example has a `Makefile` for convenience. Navigate to the example directory and run:

```bash
make run
```
