package httpclient

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"net/http"
)

func NewClient() *http.Client {
	if traceProvider := otel.GetTracerProvider(); traceProvider != nil {
		return &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		}
	}

	return &http.Client{}
}
