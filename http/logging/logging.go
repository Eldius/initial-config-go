package logging

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func ExtractRequestBody(req *http.Request) string {
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

type HTTPRequestLogRecord struct {
	URL      string           `json:"url,omitempty"`
	Method   string           `json:"method,omitempty"`
	Request  HTTPRequestData  `json:"request,omitempty"`
	Response HTTPResponseData `json:"response,omitempty"`
	Duration time.Duration    `json:"duration,omitempty"`
}

type HTTPRequestData struct {
	Body    string              `json:"body,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
}

type HTTPResponseData struct {
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	StatusCode int                 `json:"status_code,omitempty"`
}
