package client

import (
	"github.com/eldius/initial-config-go/logs"
	"io"
	"net/http"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

// HttpClient defines the interface for HTTP client operations with logging support.
type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	CloseIdleConnections()
}

type customClient struct {
	c *http.Client
}

// NewHTTPClient creates a new HTTP client with OpenTelemetry instrumentation
// and logging capabilities. If a tracer provider is configured, the client
// will automatically propagate trace context in requests.
func NewHTTPClient() *http.Client {
	var rt = http.DefaultTransport
	if traceProvider := otel.GetTracerProvider(); traceProvider != nil {
		return &http.Client{
			Transport: &loggingRoundTripper{
				proxied: otelhttp.NewTransport(http.DefaultTransport),
			},
		}
	}

	return newLoggingClient(rt)
}

// NewClient creates a new HttpClient implementation with default configuration
// and structured logging support.
func NewClient() HttpClient {
	return &customClient{
		c: NewHTTPClient(),
	}
}

func (c *customClient) Do(req *http.Request) (*http.Response, error) {
	log := logs.NewLogger(req.Context(), logs.KeyValueData{
		"method": req.Method,
		"url":    req.URL.String(),
	})
	res, err := c.c.Do(req)
	if err != nil {
		log.WithError(err).Info("failed to do http request")
		return res, err
	}
	log.WithExtraData("status", res.StatusCode).Info("http request succeeded")
	return res, err
}

func (c *customClient) CloseIdleConnections() {
	c.c.CloseIdleConnections()
}

func (c *customClient) Get(path string) (*http.Response, error) {
	return c.c.Get(path)
}

func (c *customClient) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return c.c.Post(url, contentType, body)
}

func (c *customClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.c.PostForm(url, data)
}

func (c *customClient) Head(url string) (resp *http.Response, err error) {
	return c.c.Head(url)
}
