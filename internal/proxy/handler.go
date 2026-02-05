package proxy

import (
	"log"
	"net/http"
	"net/url"
)

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
			http.Error(writer, "Could not resolve target", http.StatusNotFound)
			return
		}

		log.Printf("[handler] resolved target: %s", target.String())

		// Clone request and set target URL
		out := req.Clone(req.Context())
		out.URL = target
		out.Host = target.Host
		out.RequestURI = ""

		log.Printf("[handler] proxying to %s", out.URL.String())

		proxyHandler.ServeHTTP(writer, out)
	})
}
