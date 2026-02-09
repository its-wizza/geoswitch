package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestHostEqualsMatcher(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		reqHost  string
		expected bool
	}{
		{"exact match", "example.com", "example.com", true},
		{"case insensitive", "example.com", "EXAMPLE.COM", true},
		{"no match", "example.com", "other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := HostEqualsMatcher{Host: tt.host}
			req := &http.Request{Host: tt.reqHost}
			if got := m.Match(req); got != tt.expected {
				t.Errorf("HostEqualsMatcher.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHeaderEqualsMatcher(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		value    string
		reqHdr   string
		reqVal   string
		expected bool
	}{
		{"match", "X-Custom", "test", "X-Custom", "test", true},
		{"no match value", "X-Custom", "test", "X-Custom", "other", false},
		{"missing header", "X-Custom", "test", "X-Other", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := HeaderEqualsMatcher{Header: tt.header, Value: tt.value}
			req := &http.Request{Header: http.Header{}}
			if tt.reqHdr != "" {
				req.Header.Set(tt.reqHdr, tt.reqVal)
			}
			if got := m.Match(req); got != tt.expected {
				t.Errorf("HeaderEqualsMatcher.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTLDSuffixMatcher(t *testing.T) {
	tests := []struct {
		name     string
		tld      string
		host     string
		expected bool
	}{
		{"match .com", "com", "example.com", true},
		{"match .org", "org", "example.org", true},
		{"no match", "com", "example.org", false},
		{"subdomain", "com", "sub.example.com", true},
		{"single label no match", "com", "localhost", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := TLDMatcher{TLD: tt.tld}
			req := &http.Request{URL: &url.URL{Host: tt.host}}
			if got := m.Match(req); got != tt.expected {
				t.Errorf("TLDSuffixMatcher.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPathPrefixMatcher(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		path     string
		expected bool
	}{
		{"exact match", "/api", "/api", true},
		{"prefix match", "/api", "/api/users", true},
		{"no match", "/api", "/other", false},
		{"root", "/", "/api", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PathPrefixMatcher{Prefix: tt.prefix}
			req := &http.Request{URL: &url.URL{Path: tt.path}}
			if got := m.Match(req); got != tt.expected {
				t.Errorf("PathPrefixMatcher.Match() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllMatcher(t *testing.T) {
	t.Run("all match", func(t *testing.T) {
		m := AllMatcher{
			Matchers: []Matcher{
				HostEqualsMatcher{Host: "example.com"},
				PathPrefixMatcher{Prefix: "/api"},
			},
		}
		req := &http.Request{
			Host: "example.com",
			URL:  &url.URL{Path: "/api/users"},
		}
		if got := m.Match(req); !got {
			t.Errorf("AllMatcher.Match() = %v, want true", got)
		}
	})

	t.Run("one fails", func(t *testing.T) {
		m := AllMatcher{
			Matchers: []Matcher{
				HostEqualsMatcher{Host: "example.com"},
				PathPrefixMatcher{Prefix: "/api"},
			},
		}
		req := &http.Request{
			Host: "other.com",
			URL:  &url.URL{Path: "/api/users"},
		}
		if got := m.Match(req); got {
			t.Errorf("AllMatcher.Match() = %v, want false", got)
		}
	})
}

func TestAnyMatcher(t *testing.T) {
	t.Run("one matches", func(t *testing.T) {
		m := AnyMatcher{
			Matchers: []Matcher{
				HostEqualsMatcher{Host: "example.com"},
				HostEqualsMatcher{Host: "other.com"},
			},
		}
		req := &http.Request{Host: "other.com"}
		if got := m.Match(req); !got {
			t.Errorf("AnyMatcher.Match() = %v, want true", got)
		}
	})

	t.Run("none match", func(t *testing.T) {
		m := AnyMatcher{
			Matchers: []Matcher{
				HostEqualsMatcher{Host: "example.com"},
				HostEqualsMatcher{Host: "other.com"},
			},
		}
		req := &http.Request{Host: "third.com"}
		if got := m.Match(req); got {
			t.Errorf("AnyMatcher.Match() = %v, want false", got)
		}
	})
}
