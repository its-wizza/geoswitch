package proxy

import (
	"log"
	"net/http"
)

// NewProxyHandler returns an http.Handler that resolves the target for
// each incoming request using the provided TargetResolver. It rewrites the
// incoming request to point to the resolved target and delegates to the
// provided proxyHandler for actual proxying.
func NewProxyHandler(proxies map[Exit]http.Handler, parsers ...IntentParser) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("[handler] %s %s from %s", req.Method, req.RequestURI, req.RemoteAddr)

		// Build context
		ctx, err := ParseRequestIntent(
			req,
			parsers...,
		)
		if err != nil {
			log.Printf("[handler] error parsing request intent: %v", err)
			http.Error(writer, "Error parsing request intent", http.StatusBadRequest)
			return
		}

		// Extract target from context
		target := ctx.ParsedTarget
		if target == nil {
			log.Printf("[handler] no target resolved")
			http.Error(writer, "No target resolved", http.StatusBadRequest)
			return
		}
		// Raise error if target is not absolute URL
		if target.Scheme != "http" && target.Scheme != "https" {
			log.Printf("[handler] unsupported URL scheme: %s", target.String())
			http.Error(writer, "Unsupported URL scheme", http.StatusBadRequest)
			return
		}

		log.Printf("[handler] resolved target: %s", target.String())

		// Extract exit from context
		exit := ctx.ParsedExit
		if exit == nil {
			log.Printf("[handler] no exit parsed, using default exit")
			def := DefaultExit
			exit = &def
		}

		proxy, ok := proxies[*exit]
		if !ok {
			log.Printf("[handler] no proxy found for exit: %s", *exit)
			http.Error(writer, "No proxy found for selected exit", http.StatusBadGateway)
			return
		}

		log.Printf("[handler] selected exit: %s", *exit)

		if len(ctx.RemainingPath) > 0 {
			log.Printf("[handler] warning: unconsumed path segments: %v", ctx.RemainingPath)
		}

		// Rewrite request in-place for ReverseProxy.
		// This is safe because the request is not reused after this point.
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawQuery = target.RawQuery
		req.URL.Fragment = ""
		req.Host = target.Host
		req.RequestURI = ""

		log.Printf("[handler] proxying to %s", req.URL.String())

		proxy.ServeHTTP(writer, req)
	})
}
