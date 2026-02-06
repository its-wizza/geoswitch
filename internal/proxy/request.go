package proxy

import (
	"net/http"
	"net/url"
)

// RequestContext holds information extracted from an HTTP request
// to be used by resolvers and exit selectors.
type RequestContext struct {
	OriginalRequest *http.Request

	PathSegments []string

	RawPath string
	Query   url.Values
	Headers http.Header
}

func NewRequestContext(r *http.Request) *RequestContext {
	return &RequestContext{
		OriginalRequest: r,
		PathSegments:    splitPath(r.URL.Path),
		RawPath:         r.URL.Path,
		Query:           r.URL.Query(),
		Headers:         r.Header,
	}
}
