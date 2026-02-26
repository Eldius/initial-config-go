package server

import (
	"github.com/eldius/initial-config-go/http/logging"
	"github.com/eldius/initial-config-go/logs"
	"net/http"
	"time"
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

func (w *loggingResponseWriter) Response() logging.HTTPResponseData {
	return logging.HTTPResponseData{
		Headers:    w.Header(),
		Body:       string(w.body),
		StatusCode: w.statusCode,
	}
}

// LoggingMiddleware is a middleware that logs HTTP requests and responses.
// It extracts the request body and logs it. It also logs the request and response details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logs.NewLogger(r.Context(), logs.KeyValueData{
			"pkg": "http_server_logging",
		})
		start := time.Now()
		reqInfo := logging.HTTPRequestLogRecord{
			URL:    r.URL.String(),
			Method: r.Method,
			Request: logging.HTTPRequestData{
				Headers: r.Header,
			},
		}
		reqInfo.Request.Body = logging.ExtractRequestBody(r)

		log.WithExtraData("request", reqInfo).Info("IncomingHTTPRequestReceived")

		wWrapper := &loggingResponseWriter{ResponseWriter: w}

		next.ServeHTTP(wWrapper, r)

		reqInfo.Response = wWrapper.Response()
		reqInfo.Duration = time.Since(start)

		log.WithExtraData("request", reqInfo).Info("IncomingHTTPRequestResponse")
	})
}
