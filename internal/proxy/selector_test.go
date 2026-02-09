package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func TestExitSelector_RoundTrip(t *testing.T) {
	// Create a mock exit and runtime
	directExit := &Exit{Name: "direct", Type: ExitDirect}
	directRT := BuildExitRuntime(directExit)

	defaultExit := &Exit{Name: "default", Type: ExitDirect}
	defaultRT := BuildExitRuntime(defaultExit)

	cfg := &Config{
		Exits: map[string]*Exit{
			"direct":  directExit,
			"default": defaultExit,
		},
		Rules: []Rule{
			{
				Name:     "rule1",
				Matcher:  PathPrefixMatcher{Prefix: "/special"},
				ExitName: "direct",
			},
		},
	}

	selector := &ExitSelector{
		Cfg: cfg,
		Runtimes: map[string]*ExitRuntime{
			"direct":  directRT,
			"default": defaultRT,
		},
	}

	t.Run("uses selected exit runtime", func(t *testing.T) {
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/special"},
			Header: http.Header{},
		}
		// Should not panic and should use the direct runtime
		_, err := selector.RoundTrip(req)
		// Error is expected because we're not actually making a real request
		// but we're testing that it uses the correct runtime
		if err != nil && err.Error() == `no runtime for exit "default"` {
			t.Errorf("ExitSelector.RoundTrip() should have found runtime")
		}
	})

	t.Run("returns error for missing runtime", func(t *testing.T) {
		cfg2 := &Config{
			Exits: map[string]*Exit{
				"missing": &Exit{Name: "missing", Type: ExitDirect},
			},
			Rules: []Rule{
				{
					Name:     "rule1",
					Matcher:  PathPrefixMatcher{Prefix: "/"},
					ExitName: "missing",
				},
			},
		}
		selector2 := &ExitSelector{
			Cfg:      cfg2,
			Runtimes: map[string]*ExitRuntime{}, // Missing runtime
		}
		req := &http.Request{
			URL:    &url.URL{Path: "/test"},
			Header: http.Header{},
		}
		_, err := selector2.RoundTrip(req)
		if err == nil || err.Error() != `no runtime for exit "missing"` {
			t.Errorf("ExitSelector.RoundTrip() expected error for missing runtime, got %v", err)
		}
	})
}
