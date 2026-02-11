package main

import (
	"log"
	"net/http"

	"geoswitch/internal/proxy"
)

func main() {
	log.Println("[main] initialising geoswitch")

	cfg := &proxy.Config{
		DefaultExit: "us",
		Exits: map[string]proxy.ExitConfig{
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

	resolver := &proxy.ConfigExitResolver{
		Config: cfg,
	}

	proxies := make(map[string]http.Handler)
	for name := range cfg.Exits {
		proxies[name] = proxy.NewReverseProxy()
	}

	provider := &proxy.StaticProvider{
		Handlers: proxies,
	}

	handler := proxy.NewProxyHandler(
		resolver,
		provider,
		proxy.HeaderExitParser("X-GeoSwitch-Exit"),
		proxy.PathIntentParser,
	)

	log.Println("[main] starting GeoSwitch on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
