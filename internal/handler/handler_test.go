package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"geoswitch/internal/config"
	"geoswitch/internal/provider"
)

func TestNewProxyHandler_HappyPath_UsesDefaultExit(t *testing.T) {
	var gotReq *http.Request

	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotReq = r
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/http://example.com/api?x=1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if gotReq == nil {
		t.Fatalf("expected proxy handler to be invoked")
	}

	if gotReq.URL.Scheme != "http" {
		t.Errorf("expected scheme 'http', got '%s'", gotReq.URL.Scheme)
	}

	if gotReq.URL.Host != "example.com" {
		t.Errorf("expected host 'example.com', got '%s'", gotReq.URL.Host)
	}

	if gotReq.URL.Path != "/api" {
		t.Errorf("expected path '/api', got '%s'", gotReq.URL.Path)
	}

	if gotReq.URL.RawQuery != "x=1" {
		t.Errorf("expected query 'x=1', got '%s'", gotReq.URL.RawQuery)
	}

	if gotReq.Host != "example.com" {
		t.Errorf("expected request Host 'example.com', got '%s'", gotReq.Host)
	}

	if gotReq.RequestURI != "" {
		t.Errorf("expected empty RequestURI, got '%s'", gotReq.RequestURI)
	}
}

func TestNewProxyHandler_UsesHeaderExitWhenPresent(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
			"header-exit": {
				Provider: "test-2",
				Country:  "DE",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("default"))
		}),
		"header-exit": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("header-exit"))
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		HeaderExitParser("X-GeoSwitch-Exit"),
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/http://example.com", nil)
	req.Header.Set("X-GeoSwitch-Exit", "header-exit")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if body := w.Body.String(); body != "header-exit" {
		t.Errorf("expected body 'header-exit', got '%s'", body)
	}
}

func TestNewProxyHandler_ParseErrorReturnsBadRequest(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	failingParser := func(ctx *RequestContext) error {
		return errors.New("parse error")
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		failingParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if body := w.Body.String(); body == "" {
		t.Errorf("expected non-empty error body")
	}
}

func TestNewProxyHandler_NoTargetReturnsBadRequest(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	// No parsers provided, so no target will be resolved
	handler := NewProxyHandler(resolver, &provider.StaticProvider{Handlers: proxies})

	req := httptest.NewRequest(http.MethodGet, "/no/target", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewProxyHandler_UnsupportedSchemeReturnsBadRequest(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/ftp://example.com/resource", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewProxyHandler_UnknownExitReturnsBadRequest(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/missing/http://example.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewProxyHandler_MissingProxyForExit(t *testing.T) {
	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
			"missing-proxy": {
				Provider: "test",
				Country:  "DE",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	// Only provide proxy for "default", not "missing-proxy"
	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/missing-proxy/http://example.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, w.Code)
	}
}

func TestNewProxyHandler_HTTPSScheme(t *testing.T) {
	var gotReq *http.Request

	cfg := &config.Config{
		DefaultExit: "default",
		Exits: map[string]config.ExitConfig{
			"default": {
				Provider: "test",
				Country:  "US",
			},
		},
	}

	resolver := &config.ConfigExitResolver{
		Config: cfg,
	}

	proxies := map[string]http.Handler{
		"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotReq = r
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(
		resolver,
		&provider.StaticProvider{Handlers: proxies},
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/https://example.com/api", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if gotReq == nil {
		t.Fatal("expected proxy handler to be invoked")
	}

	if gotReq.URL.Scheme != "https" {
		t.Errorf("expected scheme 'https', got '%s'", gotReq.URL.Scheme)
	}
}

func TestNewProxyHandler_PreservesHTTPMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var gotMethod string

			cfg := &config.Config{
				DefaultExit: "default",
				Exits: map[string]config.ExitConfig{
					"default": {
						Provider: "test",
						Country:  "US",
					},
				},
			}

			resolver := &config.ConfigExitResolver{
				Config: cfg,
			}

			proxies := map[string]http.Handler{
				"default": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotMethod = r.Method
					w.WriteHeader(http.StatusOK)
				}),
			}

			handler := NewProxyHandler(
				resolver,
				&provider.StaticProvider{Handlers: proxies},
				PathIntentParser,
			)

			req := httptest.NewRequest(method, "/http://example.com", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if gotMethod != method {
				t.Errorf("expected method '%s', got '%s'", method, gotMethod)
			}
		})
	}
}
