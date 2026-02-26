package client

import (
	"bytes"
	"github.com/eldius/initial-config-go/http/logging"
	"github.com/eldius/initial-config-go/logs"
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
	log := logs.NewLogger(req.Context(), logs.KeyValueData{
		"pkg": "http_client_logging",
	})
	logData := logging.HTTPRequestLogRecord{
		URL:    req.URL.String(),
		Method: req.Method,
		Request: logging.HTTPRequestData{
			Headers: req.Header,
			Body:    logging.ExtractRequestBody(req),
		},
	}
	log.WithExtraData("request", logData).Debug("HTTPRemoteRequestStarting")

	start := time.Now()

	resp, err := lrt.proxied.RoundTrip(req)

	if err != nil {
		log.WithExtraData("error", err).WithExtraData("request", logData).Error("HTTPRequestFailed")
		return nil, err
	}
	logData.Duration = time.Since(start)
	if resp != nil {
		logData.Response = logging.HTTPResponseData{
			Headers:    resp.Header,
			Body:       extractResponseBody(resp),
			StatusCode: resp.StatusCode,
		}
	}

	log.WithExtraData("request", logData).Debug("HTTPRemoteRequestFinished")

	return resp, nil
}

func extractResponseBody(res *http.Response) string {
	if res == nil {
		slog.With("event", "extractResponseBody").Debug("NullResponse")
		return ""
	}
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

// newLoggingClient creates a new *http.Client with the logging interceptor.
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
