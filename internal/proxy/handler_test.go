package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestExitSelectorIsCalled(t *testing.T) {
	var resolverCalls int
	var exitSelectorCalls int

	// Fake resolver
	resolver := func(r *http.Request) (*url.URL, error) {
		resolverCalls++
		return url.Parse("https://example.com")
	}

	// Fake exit selector
	exitSelector := func(r *http.Request, u *url.URL) (Exit, error) {
		exitSelectorCalls++
		return DefaultExit, nil
	}

	// Fake proxy (does nothing)
	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(resolver, exitSelector, proxies)

	req := httptest.NewRequest("GET", "/https://example.com", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if resolverCalls != 1 {
		t.Fatalf("expected resolver to be called once, got %d", resolverCalls)
	}

	if exitSelectorCalls != 1 {
		t.Fatalf("expected exit selector to be called once, got %d", exitSelectorCalls)
	}
}

func TestUnknownExitReturnsBadGateway(t *testing.T) {
	resolver := func(r *http.Request) (*url.URL, error) {
		return url.Parse("https://example.com")
	}

	exitSelector := func(r *http.Request, u *url.URL) (Exit, error) {
		return Exit("nonexistent"), nil
	}

	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(resolver, exitSelector, proxies)

	req := httptest.NewRequest("GET", "/https://example.com", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rr.Code)
	}
}

func TestCorrectProxyIsUsed(t *testing.T) {
	var proxyCalled bool

	resolver := func(r *http.Request) (*url.URL, error) {
		return url.Parse("https://example.com")
	}

	exitSelector := func(r *http.Request, u *url.URL) (Exit, error) {
		return DefaultExit, nil
	}

	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			proxyCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(resolver, exitSelector, proxies)

	req := httptest.NewRequest("GET", "/https://example.com", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !proxyCalled {
		t.Fatal("expected proxy to be called, but it was not")
	}
}
