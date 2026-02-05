package proxy

import (
	"log"
	"net/http"
	"net/url"
)

// TargetResolver takes a request and returns the target URL to which the request should be proxied.
type TargetResolver func(*http.Request) (*url.URL, error)

// NewDynamicTargetHandler returns an http.Handler that resolves the target
// for each incoming request using the provided TargetResolver. The request is modified to point to the resolved target and delegates handling
// to the provided proxyHandler.
func NewDynamicTargetHandler(resolver TargetResolver, proxyHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("[handler] %s %s from %s", req.Method, req.RequestURI, req.RemoteAddr)

		target, err := resolver(req)
		if err != nil {
			log.Printf("[handler] resolver error: %v", err)
			http.Error(writer, "Could not resolve target", http.StatusBadRequest)
			return
		}
		if !target.IsAbs() || target.Host == "" {
			log.Printf("[handler] invalid target URL: %s", target.String())
			http.Error(writer, "Target must be absolute URL", http.StatusBadRequest)
			return
		}

		log.Printf("[handler] resolved target: %s", target.String())

		// Set target URL
		req.URL = target
		req.Host = target.Host
		req.RequestURI = ""

		log.Printf("[handler] proxying to %s", req.URL.String())

		proxyHandler.ServeHTTP(writer, req)
	})
}
