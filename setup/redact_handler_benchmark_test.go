package setup

import (
	"io"
	"log/slog"
	"math/rand/v2"
	"testing"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func rndStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.IntN(len(letterBytes))]
	}
	return string(b)
}

func BenchmarkRedactHandler(b *testing.B) {
	strAttr0 := rndStringBytes(10)
	strAttr1 := rndStringBytes(15)
	strAttr2 := rndStringBytes(20)

	b.Run("simple string attr values", func(b *testing.B) {
		b.Run("using simple string attributes and no keys to redact", func(b *testing.B) {
			l := slog.New(newRedactHandler(slog.NewJSONHandler(io.Discard, nil), nil))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l.With(
					slog.String("authentication", strAttr0),
					slog.String("no_redacted_attr", strAttr1),
					slog.String("secret_key", strAttr2),
				).Info("benchmarkingHandlerTest")
			}
		})

		b.Run("using simple string attributes and 1 key to redact", func(b *testing.B) {
			l := slog.New(newRedactHandler(slog.NewJSONHandler(io.Discard, nil), []string{"authentication"}))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l.With(
					slog.String("authentication", strAttr0),
					slog.String("no_redacted_attr", strAttr1),
					slog.String("secret_key", strAttr2),
				).Info("benchmarkingHandlerTest")
			}
		})
		b.Run("using simple string attributes and 2 keys to redact", func(b *testing.B) {
			l := slog.New(newRedactHandler(slog.NewJSONHandler(io.Discard, nil), []string{"authentication", "permission"}))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l.With(
					slog.String("authentication", strAttr0),
					slog.String("no_redacted_attr", strAttr1),
					slog.String("secret_key", strAttr2),
				).Info("benchmarkingHandlerTest")
			}
		})
		b.Run("using simple string attributes and 3 keys to redact", func(b *testing.B) {
			l := slog.New(newRedactHandler(slog.NewJSONHandler(io.Discard, nil), []string{"authentication", "permission", "secret_key"}))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l.With(
					slog.String("authentication", strAttr0),
					slog.String("no_redacted_attr", strAttr1),
					slog.String("secret_key", strAttr2),
				).Info("benchmarkingHandlerTest")
			}
		})
	})
}
