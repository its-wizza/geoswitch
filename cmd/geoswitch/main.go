package main

import (
	"log"
	"net/http"

	"geoswitch/internal/proxy"
)

func main() {
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

	handler := proxy.NewProxyHandler(
		resolver,
		proxies,
		proxy.HeaderExitParser("X-GeoSwitch-Exit"),
		proxy.PathIntentParser,
	)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
