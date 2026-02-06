package main

import (
	"log"
	"net/http"

	"geoswitch/internal/proxy"
)

func main() {
	log.Println("[main] initialising geoswitch")

	proxies := map[proxy.Exit]http.Handler{
		proxy.DefaultExit: proxy.NewReverseProxy(),
	}

	// Dummy exit for testing
	const testExit proxy.Exit = "test"
	proxies[testExit] = proxy.NewReverseProxy()

	handler := proxy.NewProxyHandler(
		proxy.RelativePathReferenceResolver,
		proxy.PathSegmentExitSelector,
		proxies,
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("[main] starting geoswitch on :8080")
	log.Fatal(server.ListenAndServe())
}
