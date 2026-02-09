package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestChooseExit(t *testing.T) {
	directExit := &Exit{Name: "direct", Type: ExitDirect}
	proxyURL, _ := url.Parse("http://proxy.example.com:8080")
	proxyExit := &Exit{Name: "proxy", Type: ExitHTTPProxy, ProxyURL: proxyURL}
	defaultExit := &Exit{Name: "default", Type: ExitDirect}

	cfg := &Config{
		Exits: map[string]*Exit{
			"direct":  directExit,
			"proxy":   proxyExit,
			"default": defaultExit,
		},
		Rules: []Rule{
			{
				Name:     "rule1",
				Matcher:  HostEqualsMatcher{Host: "example.com"},
				ExitName: "proxy",
			},
			{
				Name:     "rule2",
				Matcher:  PathPrefixMatcher{Prefix: "/api"},
				ExitName: "direct",
			},
		},
	}

	tests := []struct {
		name     string
		host     string
		path     string
		expected string
	}{
		{"rule1 matches", "example.com", "/", "proxy"},
		{"rule2 matches", "other.com", "/api/users", "direct"},
		{"no rule matches", "unknown.com", "/", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				Host: tt.host,
				URL:  &url.URL{Path: tt.path},
			}
			exit := ChooseExit(cfg, req)
			if exit.Name != tt.expected {
				t.Errorf("ChooseExit() = %v, want %v", exit.Name, tt.expected)
			}
		})
	}
}
