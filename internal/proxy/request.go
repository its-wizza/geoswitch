package proxy

import (
	"net/http"
	"strings"
)

// RequestContext holds information extracted from an HTTP request
// to be used by resolvers and exit selectors.
type RequestContext struct {
	OriginalRequest *http.Request

	PathSegments []string

	Path     string
	RawQuery string
	Headers  http.Header
}

func NewRequestContext(r *http.Request) *RequestContext {
	return &RequestContext{
		OriginalRequest: r,

		PathSegments: splitPath(r.URL.Path),

		Path: r.URL.Path,

		RawQuery: r.URL.RawQuery,
		Headers:  r.Header,
	}
}

// splitPath splits a URL path into its segments, ignoring leading slashes.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
