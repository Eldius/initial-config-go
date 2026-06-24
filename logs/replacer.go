package logs

import (
	"log/slog"
	"slices"
	"strings"
)

var logKeys = []string{
	"host",
	"service.name",
	"level",
	"message",
	"time",
	"error",
	"source",
	"function",
	"file",
	"line",
}

func LogAttrsReplacerFunc() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if slices.Contains(logKeys, a.Key) {
			return a
		}
		if strings.HasPrefix(a.Key, "request") ||
			strings.HasPrefix(a.Key, "response") ||
			strings.HasPrefix(a.Key, "service") {
			return a
		}

		if a.Key == "msg" {
			a.Key = "message"
			return a
		}
		return a
	}
}
