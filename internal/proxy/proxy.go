package proxy

import (
	"net/http/httputil"
)

// NewReverseProxy returns a reverse proxy that expects the request's
// URL and Host to be fully set before ServeHTTP is called.
// This proxy does NOT perform any routing decisions.
func NewReverseProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			// Target URL is already set in the request
		},
	}
}
