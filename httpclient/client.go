package httpclient

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"net/http"
)

func NewClient() *http.Client {
	if otel.GetTracerProvider() == nil {
		return &http.Client{}
	}

	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
