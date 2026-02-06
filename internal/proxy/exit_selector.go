package proxy

import (
	"net/http"
	"net/url"
)

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
