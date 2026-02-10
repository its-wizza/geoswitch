package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewReverseProxy_ReturnsValidProxy(t *testing.T) {
	// Act
	proxy := NewReverseProxy()

	// Assert
	if proxy == nil {
		t.Fatal("expected non-nil ReverseProxy")
	}
}

func TestNewReverseProxy_ProxiesRequest(t *testing.T) {
	// Arrange - Create a target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied response"))
	}))
	defer targetServer.Close()

	// Create a reverse proxy
	proxy := NewReverseProxy()

	// Create a request with the target server URL already set
	targetURL, err := url.Parse(targetServer.URL)
	if err != nil {
		t.Fatalf("failed to parse target URL: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.URL = targetURL
	req.URL.Path = "/api/test"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "proxied response" {
		t.Errorf("expected body 'proxied response', got '%s'", w.Body.String())
	}
}

func TestNewReverseProxy_ProxiesWithPath(t *testing.T) {
	// Arrange - Create a target server that echoes the path
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	targetURL, _ := url.Parse(targetServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.URL = targetURL
	req.URL.Path = "/api/v1/users"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "/api/v1/users" {
		t.Errorf("expected path '/api/v1/users' in response, got '%s'", w.Body.String())
	}
}

func TestNewReverseProxy_ProxiesWithQueryString(t *testing.T) {
	// Arrange - Create a target server that echoes the query string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.RawQuery))
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	targetURL, _ := url.Parse(targetServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test&limit=10", nil)
	req.URL = targetURL
	req.URL.Path = "/search"
	req.URL.RawQuery = "q=test&limit=10"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "q=test&limit=10" {
		t.Errorf("expected query 'q=test&limit=10', got '%s'", w.Body.String())
	}
}

func TestNewReverseProxy_ProxiesHTTPMethods(t *testing.T) {
	// Arrange - Create a target server that echoes the method
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Method))
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	targetURL, _ := url.Parse(targetServer.URL)

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
		req.URL = targetURL
		req.URL.Path = "/test"
		req.Host = targetURL.Host
		req.RequestURI = ""

		w := httptest.NewRecorder()

		// Act
		proxy.ServeHTTP(w, req)

		// Assert
		if w.Code != http.StatusOK {
			t.Errorf("method %s: expected status %d, got %d", method, http.StatusOK, w.Code)
		}

		if w.Body.String() != method {
			t.Errorf("method %s: expected body '%s', got '%s'", method, method, w.Body.String())
		}
	}
}

func TestNewReverseProxy_ProxiesHeaders(t *testing.T) {
	// Arrange - Create a target server that echoes a header
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Header.Get("X-Custom-Header")))
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	targetURL, _ := url.Parse(targetServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Header", "test-value")
	req.URL = targetURL
	req.URL.Path = "/test"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "test-value" {
		t.Errorf("expected header value 'test-value', got '%s'", w.Body.String())
	}
}

func TestNewReverseProxy_ProxiesResponseHeaders(t *testing.T) {
	// Arrange - Create a target server that sets custom headers
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Header", "response-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("body"))
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	targetURL, _ := url.Parse(targetServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.URL = targetURL
	req.URL.Path = "/test"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("X-Response-Header") != "response-value" {
		t.Errorf("expected header value 'response-value', got '%s'", w.Header().Get("X-Response-Header"))
	}
}

func TestNewReverseProxy_ProxiesStatusCodes(t *testing.T) {
	// Arrange - Create a target server that returns different status codes
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusOK
		if r.URL.Query().Get("code") != "" {
			switch r.URL.Query().Get("code") {
			case "400":
				code = http.StatusBadRequest
			case "404":
				code = http.StatusNotFound
			case "500":
				code = http.StatusInternalServerError
			}
		}
		w.WriteHeader(code)
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()

	testCases := []struct {
		name       string
		code       string
		statusCode int
	}{
		{"OK", "200", http.StatusOK},
		{"BadRequest", "400", http.StatusBadRequest},
		{"NotFound", "404", http.StatusNotFound},
		{"InternalServerError", "500", http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetURL, _ := url.Parse(targetServer.URL)

			req := httptest.NewRequest(http.MethodGet, "/test?code="+tc.code, nil)
			req.URL = targetURL
			req.URL.Path = "/test"
			req.URL.RawQuery = "code=" + tc.code
			req.Host = targetURL.Host
			req.RequestURI = ""

			w := httptest.NewRecorder()

			// Act
			proxy.ServeHTTP(w, req)

			// Assert
			if w.Code != tc.statusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, w.Code)
			}
		})
	}
}

func TestNewReverseProxy_ProxiesRequestBody(t *testing.T) {
	// Arrange - Create a target server that echoes the request body
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		w.WriteHeader(http.StatusOK)
		w.Write(buf[:n])
	}))
	defer targetServer.Close()

	proxy := NewReverseProxy()
	targetURL, _ := url.Parse(targetServer.URL)

	testBody := "test request body"
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(testBody))
	req.URL = targetURL
	req.URL.Path = "/test"
	req.Host = targetURL.Host
	req.RequestURI = ""
	req.ContentLength = int64(len(testBody))

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != testBody {
		t.Errorf("expected body '%s', got '%s'", testBody, w.Body.String())
	}
}

func TestNewReverseProxy_MultipleInstances(t *testing.T) {
	// Arrange - Verify that multiple instances can be created
	proxy1 := NewReverseProxy()
	proxy2 := NewReverseProxy()

	// Assert
	if proxy1 == nil {
		t.Fatal("proxy1 should not be nil")
	}

	if proxy2 == nil {
		t.Fatal("proxy2 should not be nil")
	}

	if proxy1 == proxy2 {
		t.Fatal("proxy1 and proxy2 should be different instances")
	}
}

func TestNewReverseProxy_HandlesHTTPSTarget(t *testing.T) {
	// Create a TLS test server
	targetServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("secure response"))
	}))
	defer targetServer.Close()

	// Create proxy with the test server's transport (which trusts the test cert)
	proxy := NewReverseProxy(WithTransport(targetServer.Client().Transport))

	targetURL, _ := url.Parse(targetServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.URL = targetURL
	req.URL.Path = "/test"
	req.Host = targetURL.Host
	req.RequestURI = ""

	w := httptest.NewRecorder()

	// Act
	proxy.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Body.String(); got != "secure response" {
		t.Errorf("expected body 'secure response', got '%s'", got)
	}
}
