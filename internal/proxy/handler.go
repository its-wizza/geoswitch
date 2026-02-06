package proxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

// TargetResolver is a function type that takes an incoming HTTP request
// and returns a target URL to which the request should be proxied.
type TargetResolver func(*http.Request) (*url.URL, error)

// ExitSelector is a function type that takes an incoming HTTP request and
// a target URL, and returns an Exit value representing an exit point (proxy server).
type ExitSelector func(*http.Request, *url.URL) (Exit, error)

type Exit string

const DefaultExit Exit = "default"

func DefaultExitSelector(_ *http.Request, _ *url.URL) (Exit, error) {
	return DefaultExit, nil
}

// PathSegmentExitSelector selects the exit based on the first path segment.
// For example, a request to /exit1/http://example.com will select exit "exit1".
func PathSegmentExitSelector(req *http.Request, _ *url.URL) (Exit, error) {
	segments := splitPath(req.URL.Path)
	if len(segments) < 1 || segments[0] == "" {
		return DefaultExit, nil
	}
	return Exit(segments[0]), nil
}

// NewProxyHandler returns an http.Handler that resolves the target for
// each incoming request using the provided TargetResolver. It rewrites the
// incoming request to point to the resolved target and delegates to the
// provided proxyHandler for actual proxying.
func NewProxyHandler(resolver TargetResolver, exitSelector ExitSelector, proxies map[Exit]http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		log.Printf("[handler] %s %s from %s", req.Method, req.RequestURI, req.RemoteAddr)

		// 1. Get target
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

		// 2. Select exit
		exit, err := exitSelector(req, target)
		if err != nil {
			log.Printf("[handler] exit selector error: %v", err)
			http.Error(writer, "Could not select exit", http.StatusInternalServerError)
			return
		}

		proxy, ok := proxies[exit]
		if !ok {
			log.Printf("[handler] no proxy found for exit: %s", exit)
			http.Error(writer, "No proxy found for selected exit", http.StatusBadGateway)
			return
		}

		log.Printf("[handler] selected exit: %s", exit)

		// Modify the request to point to the resolved target
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawQuery = target.RawQuery
		req.RequestURI = "" // RequestURI must be empty when making client requests

		log.Printf("[handler] proxying to %s", req.URL.String())

		proxy.ServeHTTP(writer, req)
	})
}

// splitPath splits a URL path into its segments, ignoring leading slashes.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
