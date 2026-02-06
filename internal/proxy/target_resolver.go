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

// RelativePathReferenceResolver extracts a URL from the request path.
// For example, a request to /http://example.com/foo will be proxied to http://example.com/foo.
func RelativePathReferenceResolver(req *http.Request) (*url.URL, error) {
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
