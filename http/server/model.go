package server

import (
	"github.com/eldius/initial-config-go/http/logging"
	"net/http"
)

var (
	_ http.ResponseWriter  = &loggingResponseWriter{}
	_ http.Flusher         = &flusherLoggingResponseWriter{}
	_ customResponseWriter = &flusherLoggingResponseWriter{}
	_ customResponseWriter = &loggingResponseWriter{}
)

type customResponseWriter interface {
	http.ResponseWriter
	Response() logging.HTTPResponseData
}

type loggingResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
}

type flusherLoggingResponseWriter struct {
	loggingResponseWriter
}

func (f *flusherLoggingResponseWriter) Flush() {
	f.loggingResponseWriter.ResponseWriter.(http.Flusher).Flush()
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

func getResponseWriter(w http.ResponseWriter) customResponseWriter {
	if _, ok := w.(http.Flusher); ok {
		return &flusherLoggingResponseWriter{
			loggingResponseWriter: loggingResponseWriter{
				ResponseWriter: w,
			},
		}
	}
	return &loggingResponseWriter{ResponseWriter: w}
}
