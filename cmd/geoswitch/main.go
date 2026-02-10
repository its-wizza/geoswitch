package main

import (
	"log"
	"net/http"

	"geoswitch/internal/proxy"
)

func main() {
	log.Println("[main] initialising geoswitch")

	// Define exits (for now they all use the same reverse proxy)
	proxies := map[proxy.Exit]http.Handler{
		proxy.DefaultExit: proxy.NewReverseProxy(),
	}

	// Dummy exit for testing
	const testExit proxy.Exit = "test"
	proxies[testExit] = proxy.NewReverseProxy()

	// Build handler with intent parsers (ORDER MATTERS)
	handler := proxy.NewProxyHandler(
		proxies,

		// 1. Highest priority: explicit header-based exit
		proxy.HeaderExitParser("X-GeoSwitch-Exit"),

		// 2. Path-based target/exit (e.g. /test/http://example.com)
		proxy.PathIntentParser,
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("[main] starting GeoSwitch on :8080")
	log.Fatal(server.ListenAndServe())
}
