package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBuildExitRuntime_Direct(t *testing.T) {
	exit := &Exit{Name: "direct", Type: ExitDirect}
	rt := BuildExitRuntime(exit)

	if rt.Exit != exit {
		t.Errorf("BuildExitRuntime() Exit mismatch")
	}
	if rt.Transport == nil {
		t.Errorf("BuildExitRuntime() Transport is nil")
	}
	if rt.Transport.Proxy != nil {
		t.Errorf("BuildExitRuntime() Proxy should be nil for ExitDirect")
	}
}

func TestBuildExitRuntime_HTTPProxy(t *testing.T) {
	proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	exit := &Exit{Name: "proxy", Type: ExitHTTPProxy, ProxyURL: proxyURL}
	rt := BuildExitRuntime(exit)

	if rt.Exit != exit {
		t.Errorf("BuildExitRuntime() Exit mismatch")
	}
	if rt.Transport == nil {
		t.Errorf("BuildExitRuntime() Transport is nil")
	}
	if rt.Transport.Proxy == nil {
		t.Errorf("BuildExitRuntime() Proxy should not be nil for ExitHTTPProxy")
	}

	// Test the proxy function returns the correct URL
	req := &http.Request{}
	proxyURLResult, _ := rt.Transport.Proxy(req)
	if proxyURLResult.String() != proxyURL.String() {
		t.Errorf("Proxy() = %v, want %v", proxyURLResult.String(), proxyURL.String())
	}
}
