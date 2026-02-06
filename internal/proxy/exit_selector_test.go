package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPathSegmentExitSelector_SelectsFirstPathSegment(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(
		http.MethodGet,
		"/test/https://example.com",
		nil,
	)

	target, _ := url.Parse("https://example.com")

	// Act
	exit, err := PathSegmentExitSelector(req, target)
	if err != nil {
		t.Fatalf("unexpected error from exit selector: %v", err)
	}

	// Assert
	if exit != Exit("test") {
		t.Fatalf("expected exit %q, got %q", "test", exit)
	}
}
