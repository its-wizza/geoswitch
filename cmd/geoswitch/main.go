package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"geoswitch/internal/config"
	"geoswitch/internal/handler"
	"geoswitch/internal/provider"
)

func main() {
	log.Println("[main] initialising GeoSwitch")

	cfg := &config.Config{
		DefaultExit: "kr",
		Exits: map[string]config.ExitConfig{
			"kr": {
				Provider: "gluetun",
				Country:  "Korea",
			},
			"uk": {
				Provider: "gluetun",
				Country:  "United Kingdom",
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("[main] invalid config: %v", err)
	}

	log.Println("[main] configuration validated successfully")

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	log.Printf("[main] initialising Gluetun provider")
	prov, err := provider.NewGluetunProvider(
		provider.WithNetwork("geoswitch-net"),
		provider.WithImageVersion("qmcgaw/gluetun:v3.41.0"),
	)
	if err != nil {
		log.Fatalf("[main] failed to create Gluetun provider: %v", err)
	}

	// Ensure cleanup happens on exit
	defer func() {
		log.Println("[main] cleaning up resources")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := prov.Close(ctx); err != nil {
			log.Printf("[main] error during cleanup: %v", err)
		}
	}()

	handler := handler.NewProxyHandler(
		resolver,
		prov,
		handler.HeaderExitParser("X-GeoSwitch-Exit"),
		handler.PathIntentParser,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Println("[main] starting GeoSwitch on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	log.Printf("[main] received signal: %v, initiating graceful shutdown", sig)

	// Attempt graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[main] error during server shutdown: %v", err)
	}

	log.Println("[main] shutdown complete")
}
