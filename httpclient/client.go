package httpclient

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type HttpClient interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (*http.Response, error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	CloseIdleConnections()
}

type customClient struct {
	c   *http.Client
	log *slog.Logger
}

func NewHTTPClient() *http.Client {
	if traceProvider := otel.GetTracerProvider(); traceProvider != nil {
		return &http.Client{
			Transport: &loggingRoundTripper{
				proxied: otelhttp.NewTransport(http.DefaultTransport),
			},
		}
	}

	return &http.Client{
		Transport: &loggingRoundTripper{
			proxied: http.DefaultTransport,
		},
	}
}

func NewClient() HttpClient {
	return &customClient{
		c:   NewHTTPClient(),
		log: slog.Default(),
	}
}

func (c *customClient) Do(req *http.Request) (*http.Response, error) {
	log := c.log.With("method", req.Method, "url", req.URL.String())
	res, err := c.c.Do(req)
	if err != nil {
		log.With("error", err).Info("failed to do http request")
		return res, err
	}
	log.With("status", res.StatusCode).Info("http request succeeded")
	return res, err
}

func (c *customClient) CloseIdleConnections() {
	c.c.CloseIdleConnections()
}

func (c *customClient) Get(path string) (*http.Response, error) {
	return c.c.Get(path)
}

func (c *customClient) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	return c.c.Post(http.MethodPost, url, body)
}
func (c *customClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.c.PostForm(url, data)
}
func (c *customClient) Head(url string) (resp *http.Response, err error) {
	return c.c.Head(url)
}
