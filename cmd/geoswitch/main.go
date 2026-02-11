package main

import (
	"log"
	"net/http"

	"geoswitch/internal/config"
	"geoswitch/internal/handler"
	"geoswitch/internal/provider"
	"geoswitch/internal/proxy"
)

func main() {
	log.Println("[main] initialising geoswitch")

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
		log.Fatalf("invalid config: %v", err)
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := make(map[string]http.Handler)
	for name := range cfg.Exits {
		proxies[name] = proxy.NewReverseProxy()
	}

	prov := &provider.StaticProvider{
		Handlers: proxies,
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
