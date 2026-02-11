package proxy

import (
	"log"
	"net/http"
)

// NewProxyHandler returns an http.Handler that resolves the target for
// each incoming request using the provided TargetResolver. It rewrites the
// incoming request to point to the resolved target and delegates to the
// provided proxyHandler for actual proxying.
func NewProxyHandler(
	resolver *ConfigExitResolver,
	proxies map[string]http.Handler,
	parsers ...IntentParser,
) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		log.Printf("[handler] %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)

		// Build context
		ctx, err := ParseRequestIntent(
			r,
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
		exitName, exitCfg, err := resolver.Resolve(ctx.ParsedExit)
		if err != nil {
			log.Printf("[handler] exit '%s' resolution failed: %v", exitName, err)
			http.Error(writer, "Unknown or unavailable exit", http.StatusBadRequest)
			return
		}

		log.Printf(
			"[handler] resolved exit '%s' (provider=%s, country=%s)",
			exitName,
			exitCfg.Provider,
			exitCfg.Country,
		)

		proxy, ok := proxies[exitName]
		if !ok {
			log.Printf("[handler] no proxy found for exit: %s", exitName)
			http.Error(writer, "No proxy found for selected exit", http.StatusBadGateway)
			return
		}

		log.Printf("[handler] selected exit: %s", exitName)

		if len(ctx.RemainingPath) > 0 {
			log.Printf("[handler] warning: unconsumed path segments: %v", ctx.RemainingPath)
		}

		req := r.Clone(r.Context())
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
