package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/eldius/initial-config-go/http/server"
	"github.com/eldius/initial-config-go/setup"
)

type myUser struct {
	id   string
	name string
}

func (u myUser) UserID() string           { return u.id }
func (u myUser) UserData() map[string]any { return map[string]any{"name": u.name} }

func main() {
	if err := setup.InitSetup(context.Background(), "http-server-app",
		setup.WithDefaultValues(map[string]any{
			"log.level":           "debug",
			"log.format":           "text",
			"log.output_to_stdout": true,
		}),
	); err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Protected endpoint with single API key auth
	auth := server.AuthenticationMiddleware(
		server.SingleUserApiKeyAuthenticationFunc("my-secret-key", "", myUser{id: "1", name: "test-user"}),
	)
	mux.Handle("GET /api/protected", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := server.AuthenticatedUserFromContext(r.Context())
		slog.Info("authenticated request", "user_id", user.UserID())
		_, _ = w.Write([]byte(`{"secret":"data"}`))
	})))

	slog.Info("server listening on :8080")
	if err := http.ListenAndServe(":8080", server.TelemetryMiddleware(mux)); err != nil {
		slog.Error("server failed", "error", err)
	}
}
