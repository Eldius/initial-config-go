package server

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
	"strings"
)

// TelemetryMiddleware is a middleware that wraps the provided http.ServeMux with OpenTelemetry instrumentation.
// It instruments the ServeMux with the otelhttp.NewHandler function and enables logging for all requests handled by the ServeMux.
func TelemetryMiddleware(mux *http.ServeMux) http.Handler {
	return otelhttp.NewHandler(
		LoggingMiddleware(mux),
		"api-call-received",
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			pattern := r.Pattern
			if pattern == "" {
				return r.Method + " " + r.URL.Path
			}
			if strings.HasPrefix(pattern, r.Method+" ") {
				return pattern
			}
			return r.Method + " " + pattern
		}))
}
