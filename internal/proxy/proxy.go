package proxy

import (
	"net/http"
	"net/http/httputil"
)

func NewReverseProxy(es *ExitSelector) *httputil.ReverseProxy {
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			// No-op: we want to preserve the original request as-is
		},
		Transport: es,
	}
	return rp
}

type ProxyServer struct {
	Cfg      *Config
	Selector *ExitSelector
	Proxy    *httputil.ReverseProxy
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}
