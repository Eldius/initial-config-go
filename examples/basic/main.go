package main

import (
	"context"
	"log/slog"

	"github.com/eldius/initial-config-go/setup"
)

func main() {
	ctx := context.Background()

	// Initialize with default settings
	if err := setup.InitSetup(ctx, "basic-app"); err != nil {
		panic(err)
	}

	slog.Info("Basic application started")
	slog.Debug("This is a debug message")
	slog.Warn("This is a warning message")
	slog.Error("This is an error message")
}
