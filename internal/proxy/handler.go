package proxy

import (
	"log"
	"net/http"
	"net/url"
)

// TargetResolver is a function type that takes an incoming HTTP request
// and returns a target URL to which the request should be proxied.
type TargetResolver func(*http.Request) (*url.URL, error)

type Exit string

// ExitSelector is a function type that takes an incoming HTTP request and
// a target URL, and returns an Exit value indicating the selected exit point.
type ExitSelector func(*http.Request, *url.URL) (Exit, error)

// NewDynamicTargetHandler returns an http.Handler that resolves the target
// for each incoming request using the provided TargetResolver. The incoming
// request is rewritten to point to the resolved target (URL, Host, and
// RequestURI) and then delegated to the provided proxyHandler for actual
// proxying.
func NewDynamicTargetHandler(resolver TargetResolver, proxyHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("[handler] %s %s from %s", req.Method, req.RequestURI, req.RemoteAddr)

		target, err := resolver(req)
		if err != nil {
			log.Printf("[handler] resolver error: %v", err)
			http.Error(writer, "Could not resolve target", http.StatusBadRequest)
			return
		}

		// Raise error if target is not absolute URL
		if !target.IsAbs() || target.Host == "" {
			log.Printf("[handler] invalid target URL: %s", target.String())
			http.Error(writer, "Target must be absolute URL", http.StatusBadRequest)
			return
		}

		log.Printf("[handler] resolved target: %s", target.String())

		// Modify the request to point to the resolved target
		req.URL = target
		req.Host = target.Host
		req.RequestURI = "" // RequestURI must be empty when making client requests

		log.Printf("[handler] proxying to %s", req.URL.String())

		proxyHandler.ServeHTTP(writer, req)
	})
}
