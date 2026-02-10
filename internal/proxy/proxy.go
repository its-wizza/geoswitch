package proxy

import (
	"net/http"
	"net/http/httputil"
)

// ProxyOption is a functional option for configuring a reverse proxy.
type ProxyOption func(*proxyConfig)

type proxyConfig struct {
	transport http.RoundTripper
}

// WithTransport sets a custom HTTP transport for the proxy.
func WithTransport(transport http.RoundTripper) ProxyOption {
	return func(c *proxyConfig) {
		c.transport = transport
	}
}

// NewReverseProxy returns a reverse proxy that expects the request's
// URL and Host to be fully set before ServeHTTP is called.
// This proxy does NOT perform any routing decisions.
//
// Options can be provided to customize the proxy behavior:
//   - WithTransport: Use a custom http.RoundTripper (default: http.DefaultTransport)
func NewReverseProxy(opts ...ProxyOption) *httputil.ReverseProxy {
	config := &proxyConfig{
		transport: http.DefaultTransport,
	}

	for _, opt := range opts {
		opt(config)
	}

	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			if req.Out == nil || req.Out.URL == nil || req.Out.URL.Host == "" {
				panic("ReverseProxy requires URL.Host to be set")
			}
		},
		Transport: config.transport,
	}
}
