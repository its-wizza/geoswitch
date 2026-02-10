package proxy

import (
	"net/http/httptest"
	"testing"
)

func TestSplitPath_BasicAndEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  []string
	}{
		{"empty", "", nil},
		{"root", "/", nil},
		{"single", "/a", []string{"a"}},
		{"trailingSlash", "/a/b/", []string{"a", "b"}},
		{"noLeadingSlash", "a/b", []string{"a", "b"}},
		{"doubleSlashInside", "/a//b", []string{"a", "", "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitPath(tc.in)

			if len(got) != len(tc.out) {
				t.Fatalf("expected %d segments, got %d", len(tc.out), len(got))
			}

			for i := range got {
				if got[i] != tc.out[i] {
					t.Errorf("segment %d: expected '%s', got '%s'", i, tc.out[i], got[i])
				}
			}
		})
	}
}

func TestPathIntentParser_ParsesTargetWithoutExit(t *testing.T) {
	req := httptest.NewRequest("GET", "/http://example.com/path?x=1", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: splitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedTarget == nil {
		t.Fatalf("expected ParsedTarget to be set")
	}

	if got := ctx.ParsedTarget.String(); got != "http://example.com/path?x=1" {
		t.Errorf("expected target 'http://example.com/path?x=1', got '%s'", got)
	}

	if ctx.ParsedExit != nil {
		t.Errorf("expected no exit, got %v", *ctx.ParsedExit)
	}

	if ctx.RemainingPath != nil {
		t.Errorf("expected RemainingPath to be nil, got %v", ctx.RemainingPath)
	}
}

func TestPathIntentParser_ParsesExitAndRemainingPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/test/extra/http://example.com/path", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: splitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || *ctx.ParsedExit != "test" {
		t.Fatalf("expected exit 'test', got %v", ctx.ParsedExit)
	}

	if len(ctx.RemainingPath) != 1 || ctx.RemainingPath[0] != "extra" {
		t.Errorf("expected RemainingPath ['extra'], got %v", ctx.RemainingPath)
	}
}

func TestPathIntentParser_DoesNotOverrideExistingExit(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo/bar/http://example.com", nil)

	existing := Exit("pre")
	ctx := &RequestContext{
		Original:      req,
		ParsedExit:    &existing,
		RemainingPath: splitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || *ctx.ParsedExit != "pre" {
		t.Fatalf("expected exit 'pre', got %v", ctx.ParsedExit)
	}

	if len(ctx.RemainingPath) != 1 || ctx.RemainingPath[0] != "bar" {
		t.Errorf("expected RemainingPath ['bar'], got %v", ctx.RemainingPath)
	}
}

func TestHeaderExitParser_SetsExitFromHeader(t *testing.T) {
	parser := HeaderExitParser("X-Exit")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Exit", "my-exit")

	ctx := &RequestContext{Original: req}

	if err := parser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || *ctx.ParsedExit != "my-exit" {
		t.Fatalf("expected exit 'my-exit', got %v", ctx.ParsedExit)
	}
}

func TestHeaderExitParser_DoesNotOverrideExistingExit(t *testing.T) {
	parser := HeaderExitParser("X-Exit")

	existing := Exit("existing")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Exit", "new")

	ctx := &RequestContext{Original: req, ParsedExit: &existing}

	if err := parser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || *ctx.ParsedExit != "existing" {
		t.Fatalf("expected exit 'existing', got %v", ctx.ParsedExit)
	}
}

func TestParseRequestIntent_HeaderThenPathParser(t *testing.T) {
	req := httptest.NewRequest("GET", "/test/http://example.com/foo", nil)
	req.Header.Set("X-GeoSwitch-Exit", "header-exit")

	ctx, err := ParseRequestIntent(
		req,
		HeaderExitParser("X-GeoSwitch-Exit"),
		PathIntentParser,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedTarget == nil || ctx.ParsedTarget.String() != "http://example.com/foo" {
		t.Fatalf("expected target 'http://example.com/foo', got %v", ctx.ParsedTarget)
	}

	if ctx.ParsedExit == nil || *ctx.ParsedExit != "header-exit" {
		t.Fatalf("expected exit 'header-exit', got %v", ctx.ParsedExit)
	}
}
