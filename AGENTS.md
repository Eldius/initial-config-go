# initial-config-go

Go library for app bootstrapping: Viper config + slog logging + OpenTelemetry. Library only — runnable demos live in `examples/` (basic, custom-config, http-client, http-server, redaction, telemetry, grafana).

Module `github.com/eldius/initial-config-go`, Go 1.27 per `go.mod` (CI pins 1.26.1 and relies on toolchain auto-switch).

## Commands

| Command | What |
|---|---|
| `make test` / `task test` | `go test ./... -cover` |
| `make lint` / `task lint` | `golangci-lint run` (no `.golangci.yml` — default linters) |
| `make vulncheck` | `go tool govulncheck ./...` |
| `make validate` | test → lint → vulncheck |
| `make benchmark` | `go test -bench=. -benchmem -count=20 ./...` |
| `make release` | validates, then tags next patch version and **pushes to origin** — don't run unless asked |
| `make telemetry-example` | Docker Compose Grafana LGTM demo stack |

Make/Task call bare `golangci-lint` / `task`; if missing from PATH use `go tool golangci-lint run` / `go tool task ...` (both declared as `go.mod` tool deps).

## Packages

- **`setup`** — entrypoint `InitSetup(ctx, appName, opts...)`: wires Viper → logs → telemetry. Cobra helpers `PersistentPreRunE(appName, opts...)` / `PersistentPostRunE(waitTime)` in `setup/setup_helpers.go` (the post-run sleep gives OTel batch processors time to flush).
- **`configs`** — `log.*` / `telemetry.*` key constants + Viper accessors.
- **`logs`** — `NewLogger(ctx, ...)` facade over slog; `NewRedactHandler`, `NewTracingHandler` (injects `trace_id`/`span_id` from ctx span), `LogHandler`, `GetWriter` / `CloseLogFiles`.
- **`telemetry`** — `InitTelemetry`, `ProviderSet` getters/setters, `TelemetryShutdown` (also closes log files), `GetSqlxDB` / `GetDB` (otelsql-instrumented), `NewSpan` / `GetTracer` / `GetMeter`.
- **`http/client`** — `NewHTTPClient()` (otelhttp transport), `NewClient()` interface wrapper.
- **`http/server`** — `TelemetryMiddleware`, `LoggingMiddleware`, `AuthenticationMiddleware` (+ single/multiple API-key auth funcs; `ApiKeyMap` buckets keys by HMAC lookup token to avoid one bcrypt compare per stored key per request).
- **`http/logging`** — shared request/response log record types.

## Gotchas

- **Env vars**: `WithEnvPrefix("MYAPP")` is lowercased internally but Viper uppercases for matching → expect `MYAPP_*`. Default prefix is `app` → `APP_*`. `.` in keys becomes `_` (`log.format` → `APP_LOG_FORMAT`).
- **Log output validation**: `InitSetup` fails only if `log.output_to_stdout=false` AND `log.output_to_file` is empty. The check runs before the OTel branch, so an OTel logs endpoint does NOT satisfy it. Default is stdout=true.
- **OTel log shipping hijacks slog**: when telemetry is enabled AND `telemetry.logs.endpoint` is set, slog's default handler is replaced by the otelslog bridge — stdout/file logging is skipped entirely, and trace correlation is native (bridge sets TraceID/SpanID fields). In the stdout/file branch, `setup` wraps the handler with `logs.NewTracingHandler` when telemetry is enabled AND a traces endpoint is set — correlation keys are `trace_id`/`span_id` (`logs.TraceIDKey`/`SpanIDKey`).
- **Redacted keys** (`log.redacted_keys`): YAML list or comma-separated env string (`APP_LOG_REDACTED_KEYS=password,token`). Case-insensitive substring match via `strings.Contains`.
- **Telemetry off by default**: `OTELConfigs.IsEnabled()` = `Enabled && ≥1 endpoint`. All OTLP exporters use insecure gRPC → endpoints are `host:port` (e.g. `localhost:4317`), no scheme/TLS.
- **Default HTTP client is NOT instrumented**: `setup.WithInstrumentHTTPClient(true)` replaces `http.DefaultClient`; otherwise call `client.NewHTTPClient()` explicitly.
- **otelhttp span names**: `TelemetryMiddleware` names spans from `r.Pattern` (Go 1.22+ mux pattern), falling back to `METHOD path`.
- **`LogAttrsReplacerFunc`** effectively only renames `msg`→`message`; all other attrs pass through unchanged (its keep-lists are no-ops).
- **Tests**: use `t.Context()`. `telemetry/shutdown_test.go` is `package telemetry_test` (external, avoids import cycle). `setup/logs_test.go` writes `my-log-file*.log` into `setup/` — those files are committed, so running tests can dirty `git status`.
- **CI** (`.github/workflows/ci.yml`): `go mod tidy`, `make test`, `make vulncheck`, golangci-lint-action — it does NOT run `make lint` / `make validate`.
- **Lint is broken on Go 1.27**: pinned golangci-lint v1.64.8's staticcheck panics repo-wide (`buildir: interface conversion ... *buildir.IR`) — pre-existing, not caused by your change. Verify with `--no-config --disable-all -E govet,errcheck,ineffassign` until the tool dep is bumped.
