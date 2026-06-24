# initial-config-go

Go library for app bootstrapping (Viper config + slog logging + OpenTelemetry).

Module: `github.com/eldius/initial-config-go` / Go 1.26.4

## Commands

| Command | What |
|---|---|
| `make test` / `task test` | `go test ./... -cover` |
| `make lint` / `task lint` | `golangci-lint run` |
| `make vulncheck` / `task vulncheck` | `go tool govulncheck ./...` |
| `make validate` / `task validate` | test + lint + vulncheck (order matters) |
| `make benchmark` | `go test -bench=. -benchmem -count=20 ./...` |
| `make release` | tags next version and pushes |
| `make telemetry-example` | Docker Compose Grafana LGTM stack |

Tools are declared in `go.mod` `tool` directives — use `go tool <name>`.

## Key packages

- **`setup`** — public entrypoint: `InitSetup(ctx, appName, opts...)`. Cobra helpers: `PersistentPreRunE`, `PersistentPostRunE` in `setup/setup_helpers.go`. Options: `WithInstrumentHTTPClient`, `WithOpenTelemetryOptions`, etc.
- **`configs`** — config key constants, accessor funcs via Viper.
- **`logs`** — `logs.NewLogger(ctx)` returns a `Logger` interface. Also: `NewRedactHandler`, `LogHandler` (JSON/text factory), `LogAttrsReplacerFunc` (msg→message), `GetWriter`, `CloseLogFiles`.
- **`telemetry`** — `InitTelemetry(ctx, opts...)`, `ProviderSet` (getter/setter for meter/tracer/logger providers), `TelemetryShutdown(ctx)`, `TelemetryForceFlush(ctx)`. Also exports `GetSqlxDB`/`GetDB` for instrumented SQL via `otelsql`, and `NewSpan`/`GetTracer`/`GetMeter`.
- **`http/client`** — `NewHTTPClient()` returns `*http.Client`; `NewClient()` returns `HttpClient` interface.
- **`http/server`** — `TelemetryMiddleware(mux)`, `LoggingMiddleware`, `AuthenticationMiddleware`, `SingleUserApiKeyAuthenticationFunc`, `MultipleUserApiKeyAuthenticationFunc` (bcrypt + lookup token).
- **`http/logging`** — shared `HTTPRequestLogRecord`, `HTTPRequestData`, `HTTPResponseData` types.

## Gotchas

- **Env prefix**: `WithEnvPrefix("MYAPP")` lowercases to `myapp` internally, but Viper uppercases for env matching → expect `MYAPP_*` env vars (not `myapp_*`).
- **Config key dots**: `log.format` in code → nested YAML: `log: { format: json }`. Viper's `SetEnvKeyReplacer` replaces `.` with `_` for env vars → `APP_LOG_FORMAT`.
- **Log output required**: At least one of `log.output_to_stdout`, `log.output_to_file`, or OpenTelemetry logs endpoint must be configured, or `InitSetup` returns an error.
- **Redacted keys**: Accept YAML lists **or** comma-separated env var strings (e.g. `APP_LOG_REDACTED_KEYS=password,token,authorization`). Matching is case-insensitive partial via `strings.Contains`.
- **Default log output**: `Options.GetDefaultValues()` sets `log.output_to_stdout = true` unless explicitly overridden.
- **Telemetry**: Not enabled by default. `OTELConfigs.IsEnabled()` returns true only if `Enabled == true` **and** at least one endpoint is non-empty.
- **Tests**: Use `t.Context()` (Go 1.26+). External test packages to avoid import cycles (e.g. `telemetry/shutdown_test.go` uses `package telemetry_test`).
- **CI**: Runs `go mod tidy`, `make test`, `make vulncheck`, and `golangci-lint-action` — does NOT run `make lint`.
- **otelhttp span naming**: `TelemetryMiddleware` uses custom `SpanNameFormatter` that prefers `r.Pattern` over path.
- **Auth**: `AuthenticationMiddleware` returns `func(http.Handler) http.Handler`. bcrypt-based `ApiKeyMap` uses HMAC lookup token to reduce bcrypt comparisons.
- **Log attrib replacer**: `logs.LogAttrsReplacerFunc()` keeps `host`, `service.*`, `error`, `source`, `request*`, `response*`, maps `msg`→`message`; all other attrs pass through.
- **Default HTTP client**: NOT instrumented by default. Use `setup.WithInstrumentHTTPClient(true)` to replace `http.DefaultClient` with OTel-instrumented client, or call `client.NewHTTPClient()` explicitly.
- **Telemetry shutdown**: `telemetry.TelemetryShutdown(ctx)` also closes log files. Called automatically by `PersistentPostRunE`.
- **Benchmarks**: in `logs/redact_handler_benchmark_test.go`.
