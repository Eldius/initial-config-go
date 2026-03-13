package main

import (
	"context"
	"fmt"
	"github.com/eldius/initial-config-go/setup"
	"github.com/spf13/viper"
	"log/slog"
)

func main() {
	ctx := context.Background()

	// Initialize with custom properties and defaults
	if err := setup.InitSetup(ctx, "custom-config-app",
		setup.WithDefaultValues(map[string]any{
			"app.feature.enabled": true,
			"app.timeout":         30,
			"app.api.url":         "https://api.example.com",
		}),
		setup.WithProps(
			setup.Prop{Key: "custom.key", Value: "initial-value"},
		),
		setup.WithEnvPrefix("CUSTOMAPP"),
	); err != nil {
		panic(err)
	}

	slog.Info("Custom configuration application started")

	// Read custom configuration using Viper
	featureEnabled := viper.GetBool("app.feature.enabled")
	timeout := viper.GetInt("app.timeout")
	apiUrl := viper.GetString("app.api.url")
	customKey := viper.GetString("custom.key")

	slog.Info("Configuration values:")
	slog.Info(fmt.Sprintf("  app.feature.enabled: %v", featureEnabled))
	slog.Info(fmt.Sprintf("  app.timeout: %d", timeout))
	slog.Info(fmt.Sprintf("  app.api.url: %s", apiUrl))
	slog.Info(fmt.Sprintf("  custom.key: %s", customKey))

	if featureEnabled {
		slog.Info("Special feature is active!")
	}
}
