package main

import (
	"log"
	"net/http"

	"geoswitch/internal/config"
	"geoswitch/internal/handler"
	"geoswitch/internal/provider"
)

func main() {
	log.Println("[main] initialising GeoSwitch")

	cfg := &config.Config{
		DefaultExit: "us",
		Exits: map[string]config.ExitConfig{
			"us": {
				Provider: "gluetun",
				Country:  "US",
			},
			"de": {
				Provider: "gluetun",
				Country:  "DE",
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
	prov, err := provider.NewGluetunProvider("geoswitch-net")
	if err != nil {
		log.Fatalf("[main] failed to create Gluetun provider: %v", err)
	}

	handler := handler.NewProxyHandler(
		resolver,
		prov,
		handler.HeaderExitParser("X-GeoSwitch-Exit"),
		handler.PathIntentParser,
	)

	log.Println("[main] starting GeoSwitch on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
