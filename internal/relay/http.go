package relay

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type TargetResolver func(*http.Request) (*url.URL, error)

func NewDynamicHostReverseProxy(resolve TargetResolver) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			target, err := resolve(r.In)
			if err != nil {
				return
			}

			r.SetURL(target)
		},
	}
}

func extractRelativePathReference(r *http.Request) (*url.URL, error) {
	rel := r.URL.Path

	if r.URL.RawQuery != "" {
		rel += "?" + r.URL.RawQuery
	}
	if r.URL.Fragment != "" {
		rel += "#" + r.URL.Fragment
	}

	return &url.URL{Path: rel}, nil
}
