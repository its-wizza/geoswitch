package proxy

import (
	"net/http/httptest"
	"strings"
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
			got := SplitPath(tc.in)

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
		RemainingPath: SplitPath(req.URL.Path),
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

	if len(ctx.RemainingPath) != 0 {
		t.Errorf("expected RemainingPath to be empty, got %v", ctx.RemainingPath)
	}
}

func TestPathIntentParser_ParsesExitAndRemainingPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/test/extra/http://example.com/path", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "test" {
		t.Fatalf("expected exit 'test', got %v", ctx.ParsedExit)
	}

	if len(ctx.RemainingPath) != 1 || ctx.RemainingPath[0] != "extra" {
		t.Errorf("expected RemainingPath ['extra'], got %v", ctx.RemainingPath)
	}
}

func TestPathIntentParser_DoesNotOverrideExistingExit(t *testing.T) {
	req := httptest.NewRequest("GET", "/foo/bar/http://example.com", nil)

	existing := Exit{
		Name: "pre",
	}
	ctx := &RequestContext{
		Original:      req,
		ParsedExit:    &existing,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "pre" {
		t.Fatalf("expected exit 'pre', got %v", ctx.ParsedExit)
	}

	if len(ctx.RemainingPath) != 2 || ctx.RemainingPath[0] != "foo" || ctx.RemainingPath[1] != "bar" {
		t.Errorf("expected RemainingPath ['foo', 'bar'], got %v", ctx.RemainingPath)
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

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "my-exit" {
		t.Fatalf("expected exit 'my-exit', got %v", ctx.ParsedExit)
	}
}

func TestHeaderExitParser_DoesNotOverrideExistingExit(t *testing.T) {
	parser := HeaderExitParser("X-Exit")

	existing := Exit{
		Name: "existing",
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Exit", "new")

	ctx := &RequestContext{Original: req, ParsedExit: &existing}

	if err := parser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "existing" {
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

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "header-exit" {
		t.Fatalf("expected exit 'header-exit', got %v", ctx.ParsedExit)
	}
}

func TestPathIntentParser_MultipleURLsInPath(t *testing.T) {
	// What happens with /http://example.com/http://another.com?
	req := httptest.NewRequest("GET", "/http://example.com/http://another.com", nil)
	ctx := &RequestContext{
		Original:      req,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should parse first URL
	if ctx.ParsedTarget == nil || ctx.ParsedTarget.String() != "http://example.com/http://another.com" {
		t.Errorf("expected first URL to be parsed fully")
	}
}

func TestPathIntentParser_URLWithFragment(t *testing.T) {
	// Fragments should be preserved in the target URL
	req := httptest.NewRequest("GET", "/http://example.com/path#section", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedTarget == nil {
		t.Fatalf("expected ParsedTarget to be set")
	}

	expected := "http://example.com/path#section"
	if got := ctx.ParsedTarget.String(); got != expected {
		t.Errorf("expected target '%s', got '%s'", expected, got)
	}
}

func TestPathIntentParser_URLWithEncodedCharacters(t *testing.T) {
	req := httptest.NewRequest("GET", "/http://example.com/path%20with%20spaces", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.ParsedTarget == nil {
		t.Fatalf("expected ParsedTarget to be set")
	}

	if got := ctx.ParsedTarget.String(); got != "http://example.com/path%20with%20spaces" {
		t.Errorf("expected encoded characters preserved, got '%s'", got)
	}
}

func TestPathIntentParser_UnconsumedSegments(t *testing.T) {
	req := httptest.NewRequest("GET", "/exit/http://example.com/extra/segments", nil)

	ctx := &RequestContext{
		Original:      req,
		RemainingPath: SplitPath(req.URL.Path),
	}

	if err := PathIntentParser(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The URL parsing should consume from "http://example.com/extra/segments"
	// as a complete URL
	if ctx.ParsedTarget == nil {
		t.Fatalf("expected ParsedTarget to be set")
	}

	expected := "http://example.com/extra/segments"
	if got := ctx.ParsedTarget.String(); got != expected {
		t.Errorf("expected target '%s', got '%s'", expected, got)
	}

	if ctx.ParsedExit == nil || ctx.ParsedExit.Name != "exit" {
		t.Errorf("expected exit 'exit', got %v", ctx.ParsedExit)
	}

	if len(ctx.RemainingPath) != 0 {
		t.Errorf("expected empty remaining path, got %v", ctx.RemainingPath)
	}
}

func TestHeaderExitParser_InvalidExitNames(t *testing.T) {
	tests := []struct {
		name      string
		headerVal string
		shouldSet bool
	}{
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"valid name", "us-west", true},
		{"with special chars", "exit@123", true},
		{"very long string", strings.Repeat("a", 1000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := HeaderExitParser("X-Exit")
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Exit", tt.headerVal)

			ctx := &RequestContext{
				Original:      req,
				RemainingPath: []string{},
			}

			if err := parser(ctx); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.shouldSet && ctx.ParsedExit == nil {
				t.Errorf("expected exit to be set for '%s'", tt.headerVal)
			}

			if !tt.shouldSet && ctx.ParsedExit != nil {
				t.Errorf("expected exit not to be set for '%s', got %v", tt.headerVal, *ctx.ParsedExit)
			}
		})
	}
}

func TestSplitPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{"double slashes", "//foo//bar//", []string{"", "foo", "", "bar", ""}},
		{"many slashes", "/////", []string{"", "", "", ""}},
		{"dot segments", "/./foo/../bar", []string{".", "foo", "..", "bar"}},
		{"unicode", "/café/مرحبا", []string{"café", "مرحبا"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitPath(tt.path)
			if len(got) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(got))
				return
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("at index %d: expected '%s', got '%s'", i, tt.expected[i], got[i])
				}
			}
		})
	}
}
