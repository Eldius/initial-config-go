package server

import (
	"github.com/eldius/initial-config-go/http/logging"
	"github.com/eldius/initial-config-go/logs"
	"net/http"
	"time"
)

// LoggingMiddleware is a middleware that logs HTTP requests and responses.
// It extracts the request body and logs it. It also logs the request and response details.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logs.NewLogger(r.Context(), logs.KeyValueData{
			"pkg": "http_server_logging",
		})
		start := time.Now()
		reqInfo := logging.HTTPRequestLogRecord{
			Pattern: r.Pattern,
			URL:     r.URL.String(),
			Method:  r.Method,
			Request: logging.HTTPRequestData{
				Headers: r.Header,
			},
		}
		reqInfo.Request.Body = logging.ExtractRequestBody(r)

		log.WithExtraData("request", reqInfo).Info("IncomingHTTPRequestReceived")

		wWrapper := getResponseWriter(w)

		next.ServeHTTP(wWrapper, r)

		reqInfo.Response = wWrapper.Response()
		reqInfo.Duration = time.Since(start)

		log.WithExtraData("request", reqInfo).Info("IncomingHTTPRequestResponse")
	})
}
