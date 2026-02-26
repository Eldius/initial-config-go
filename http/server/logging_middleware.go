package server

import (
	"bytes"
	"fmt"
	"github.com/eldius/initial-config-go/logs"
	"io"
	"net/http"
)

var (
	_ http.ResponseWriter = &loggingResponseWriter{}
)

type loggingResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *loggingResponseWriter) Response() responseData {
	return responseData{
		Headers:    w.Header(),
		Body:       string(w.body),
		StatusCode: w.statusCode,
	}
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqInfo := requestLogRecord{
			URL:    r.URL.String(),
			Method: r.Method,
			Request: requestData{
				Headers: r.Header,
			},
		}
		if b, err := extractRequestBody(r); err != nil {
			logs.NewLogger(r.Context(), logs.KeyValueData{
				"request": reqInfo,
			}).WithError(err).Warn("RequestError")
		} else {
			reqInfo.Request.Body = string(b)
		}
		wWrapper := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wWrapper, r)

		reqInfo.Response = wWrapper.Response()

		logs.NewLogger(r.Context(), logs.KeyValueData{
			"request": reqInfo,
		}).Info("httpRequest")
	})
}

func extractRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	reader := r.Body
	defer func() {
		_ = reader.Close()
	}()
	b, err := io.ReadAll(reader)
	if err != nil {
		err = fmt.Errorf("failed to read body content: %w", err)
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(b))
	return b, nil
}

type requestLogRecord struct {
	URL      string       `json:"url,omitempty"`
	Method   string       `json:"method,omitempty"`
	Request  requestData  `json:"request,omitempty"`
	Response responseData `json:"response,omitempty"`
}

type requestData struct {
	Body    string              `json:"body,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
}

type responseData struct {
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	StatusCode int                 `json:"status_code,omitempty"`
}
