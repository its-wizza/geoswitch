package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProxyHandler_HappyPath_UsesDefaultExit(t *testing.T) {
	var gotReq *http.Request

	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotReq = r
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}),
	}

	handler := NewProxyHandler(proxies, PathIntentParser)

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
	var headerExit = Exit{
		Name: "header-exit",
	}

	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("default"))
		}),
		headerExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("header-exit"))
		}),
	}

	handler := NewProxyHandler(
		proxies,
		HeaderExitParser("X-GeoSwitch-Exit"),
		PathIntentParser,
	)

	req := httptest.NewRequest(http.MethodGet, "/http://example.com", nil)
	req.Header.Set("X-GeoSwitch-Exit", headerExit.Name)
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
	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	failingParser := func(ctx *RequestContext) error {
		return errors.New("parse error")
	}

	handler := NewProxyHandler(proxies, failingParser)

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
	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	// No parsers provided, so no target will be resolved
	handler := NewProxyHandler(proxies)

	req := httptest.NewRequest(http.MethodGet, "/no/target", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewProxyHandler_UnsupportedSchemeReturnsBadRequest(t *testing.T) {
	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(proxies, PathIntentParser)

	req := httptest.NewRequest(http.MethodGet, "/ftp://example.com/resource", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewProxyHandler_UnknownExitReturnsBadGateway(t *testing.T) {
	proxies := map[Exit]http.Handler{
		DefaultExit: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}

	handler := NewProxyHandler(proxies, PathIntentParser)

	req := httptest.NewRequest(http.MethodGet, "/missing/http://example.com", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, w.Code)
	}
}
