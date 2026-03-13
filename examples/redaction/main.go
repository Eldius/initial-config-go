package main

import (
	"context"
	"github.com/eldius/initial-config-go/setup"
	"log/slog"
)

func main() {
	ctx := context.Background()

	// Initialize with redaction configured
	if err := setup.InitSetup(ctx, "redaction-app",
		setup.WithDefaultValues(map[string]any{
			"log.redacted_keys": []string{"password", "api_key", "secret"},
			"log.format":        "json", // Redaction works better to visualize in JSON
			"log.level":         "debug",
			"log.output_to_stdout": true,
		}),
	); err != nil {
		panic(err)
	}

	slog.Info("Redaction application started")

	// Sensitive data in log messages
	slog.Info("Attempting to log sensitive data", 
		slog.String("user", "john_doe"),
		slog.String("password", "p4ssw0rd!"), // Should be redacted
		slog.String("api_key", "12345-67890"), // Should be redacted
		slog.String("public_info", "this is fine"),
		slog.Group("credentials",
			slog.String("secret", "hidden-value"), // Should be redacted
		),
	)
}
