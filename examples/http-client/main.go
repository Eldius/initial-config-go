package main

import (
	"context"
	"github.com/eldius/initial-config-go/http/client"
	"github.com/eldius/initial-config-go/logs"
	"github.com/eldius/initial-config-go/setup"
	"io"
	"log/slog"
)

func main() {
	ctx := context.Background()

	// Initialize application
	if err := setup.InitSetup(ctx, "http-client-app"); err != nil {
		panic(err)
	}

	log := logs.NewLogger(ctx)
	log.Info("HTTP Client application started")

	// Create instrumented HTTP client
	c := client.NewClient()

	// Perform a request
	resp, err := c.Get("https://httpbin.org/get")
	if err != nil {
		log.WithError(err).Error("Failed to perform HTTP request")
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, _ := io.ReadAll(resp.Body)
	slog.Info("Response status: " + resp.Status)
	slog.Debug("Response body: " + string(body))
}
