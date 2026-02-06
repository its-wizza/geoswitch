package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"geoswitch/internal/proxy"
)

// relativePathReferenceResolver extracts a URL from the request path.
// For example, a request to /http://example.com/foo will be proxied to http://example.com/foo.
func relativePathReferenceResolver(req *http.Request) (*url.URL, error) {
	log.Printf("[resolver] extracting relative path reference from %s", req.URL.String())

	// Get the full URL path (including query and fragment), trimming the leading slash
	target := strings.TrimPrefix(req.URL.EscapedPath(), "/")
	if req.URL.RawQuery != "" {
		target += "?" + req.URL.RawQuery
	}

	parsedURL, err := url.Parse(target)
	if err != nil {
		log.Printf("[resolver] parse error: %v", err)
		return nil, err
	}

	log.Printf("[resolver] parsed URL: %s", parsedURL.String())
	return parsedURL, nil
}

func main() {
	log.Println("[main] initialising geoswitch")

	proxies := map[proxy.Exit]http.Handler{
		proxy.DefaultExit: proxy.NewReverseProxy(),
	}

	// Dummy exit for testing
	const testExit proxy.Exit = "test"
	proxies[testExit] = proxy.NewReverseProxy()

	handler := proxy.NewProxyHandler(
		relativePathReferenceResolver,
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
