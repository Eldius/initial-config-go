package httpclient

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// loggingRoundTripper is a struct that implements the http.RoundTripper interface.
// It wraps an existing http.RoundTripper to add logging functionality.
type loggingRoundTripper struct {
	proxied http.RoundTripper
}

// RoundTrip is the core of the interceptor. It's called for each HTTP request.
// It logs request and response details and measures the request duration.
func (lrt *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	log := slog.With("pkg", "http_logging")
	logData := map[string]any{
		"url":          req.URL.String(),
		"method":       req.Method,
		"headers":      req.Header,
		"request_body": extractRequestBody(req),
	}
	log.With("request", logData).Debug("HTTPRequestStarting")

	start := time.Now()

	resp, err := lrt.proxied.RoundTrip(req)

	logData["duration"] = time.Since(start)
	logData["response_body"] = extractResponseBody(resp)

	if err != nil {
		log.With("error", err, "request", logData).Error("HTTPRequestFailed")
		return nil, err
	}

	log.With("request", logData).Debug("HTTPRequestFinished")

	return resp, nil
}

func extractRequestBody(req *http.Request) string {
	if req.Body == nil {
		return ""
	}
	reader := req.Body
	defer func() {
		_ = reader.Close()
	}()
	body, _ := io.ReadAll(reader)
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

func extractResponseBody(res *http.Response) string {
	if res.Body == nil {
		return ""
	}
	reader := res.Body
	defer func() {
		_ = reader.Close()
	}()
	body, _ := io.ReadAll(reader)
	res.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

// newLoggingClient creates an *http.Client with the logging interceptor.
func newLoggingClient(rt http.RoundTripper) *http.Client {
	// If http.DefaultTransport is used, it may be shared among different clients,
	// which can be fine. However, creating a new transport provides isolation.
	if rt == nil {
		rt = http.DefaultTransport
	}

	return &http.Client{
		// Set the Transport to our custom loggingRoundTripper.
		// We pass the original default transport to be "proxied" or wrapped.
		Transport: &loggingRoundTripper{
			proxied: rt,
		},
	}
}
