package proxy

import (
	"net/http"
	"net/url"
)

type ExitRuntime struct {
	Exit      *Exit
	Transport *http.Transport
}

func BuildExitRuntime(exit *Exit) *ExitRuntime {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not *http.Transport")
	}
	tr := base.Clone()

	switch exit.Type {
	case ExitHTTPProxy:
		proxyURL := exit.ProxyURL
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxyURL, nil
		}
	case ExitDirect:
		tr.Proxy = nil
	}

	return &ExitRuntime{
		Exit:      exit,
		Transport: tr,
	}
}
