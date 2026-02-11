package proxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

// RequestContext holds information extracted from an HTTP request to be used by parsers.
type RequestContext struct {
	Original *http.Request // original HTTP request

	ParsedTarget *url.URL // nil if none explicitly requested
	ParsedExit   *Exit    // nil if none explicitly requested

	RemainingPath []string // unconsumed path segments
}

type IntentParser func(*RequestContext) error

func PathIntentParser(ctx *RequestContext) error {
	if ctx.ParsedTarget != nil || len(ctx.RemainingPath) == 0 {
		return nil
	}

	targetURL, controlSegments := findTargetURLInPath(ctx)
	if targetURL == nil {
		return nil
	}

	ctx.ParsedTarget = targetURL
	updateExitFromControl(ctx, controlSegments)

	if ctx.ParsedExit != nil {
		log.Printf("[parser] path intent parser: found target '%s' and exit '%s' from path", targetURL.String(), ctx.ParsedExit.Name)
		return nil
	} else {
		log.Printf("[parser] path intent parser: found target '%s' from path", targetURL.String())
	}

	return nil
}

func HeaderExitParser(headerName string) IntentParser {
	return func(ctx *RequestContext) error {
		if ctx.ParsedExit != nil {
			return nil
		}

		val := strings.TrimSpace(ctx.Original.Header.Get(headerName))
		if val == "" {
			return nil
		}

		ctx.ParsedExit = &Exit{Name: val}
		log.Printf("[parser] header exit parser: found exit '%s' from header '%s'", ctx.ParsedExit.Name, headerName)
		return nil
	}
}

func ParseRequestIntent(r *http.Request, parsers ...IntentParser) (*RequestContext, error) {
	ctx := &RequestContext{
		Original:      r,
		RemainingPath: SplitPath(r.URL.Path),
	}

	log.Printf("[parser] parsing request intent: %s %s, path segments: %v", r.Method, r.URL.Path, ctx.RemainingPath)

	// Chain parsers, operating on the request context
	for _, parse := range parsers {
		if err := parse(ctx); err != nil {
			return nil, err
		}
	}

	logParsedIntent(ctx)

	return ctx, nil
}

// SplitPath splits a URL path into its segments, ignoring a leading or trailing slash.
// Empty paths return an empty slice. This is useful for path-based routing and parsing.
//
// Examples:
//
//	SplitPath("/foo/bar")     -> ["foo", "bar"]
//	SplitPath("/foo/bar/")    -> ["foo", "bar"]
//	SplitPath("foo/bar")      -> ["foo", "bar"]
//	SplitPath("/")            -> []
//	SplitPath("")             -> []
func SplitPath(path string) []string {
	// Trim (single) leading and trailing slashes
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

// findTargetURLInPath searches for an absolute URL in the remaining path segments.
// Returns the URL and the control segments that preceded it.
func findTargetURLInPath(ctx *RequestContext) (*url.URL, []string) {
	for i := 0; i < len(ctx.RemainingPath); i++ {
		candidate := strings.Join(ctx.RemainingPath[i:], "/")

		if ctx.Original.URL.RawQuery != "" {
			candidate += "?" + ctx.Original.URL.RawQuery
		}

		if u := parseAbsoluteURL(candidate); u != nil {
			return u, ctx.RemainingPath[:i]
		}
	}
	return nil, nil
}

// parseAbsoluteURL attempts to parse a string as an absolute URL.
// Returns nil if parsing fails or the URL is not absolute.
func parseAbsoluteURL(candidate string) *url.URL {
	u, err := url.Parse(candidate)
	if err != nil || !u.IsAbs() {
		return nil
	}
	return u
}

// updateExitFromControl updates the exit and remaining path based on control segments.
func updateExitFromControl(ctx *RequestContext, control []string) {
	if ctx.ParsedExit == nil && len(control) > 0 {
		// This parser consumes the first control segment as exit
		exit := Exit{Name: control[0]}
		ctx.ParsedExit = &exit
		ctx.RemainingPath = control[1:]
	} else if len(control) > 0 {
		// Exit already set elsewhere, preserve all control segments
		ctx.RemainingPath = control
	} else {
		// Explicitly set to empty slice for consistency
		ctx.RemainingPath = []string{}
	}
}

// logParsedIntent logs the parsed exit and target information.
func logParsedIntent(ctx *RequestContext) {
	var exitStr string
	if ctx.ParsedExit != nil {
		exitStr = ctx.ParsedExit.Name
	} else {
		exitStr = "<nil>"
	}

	var targetStr string
	if ctx.ParsedTarget != nil {
		targetStr = ctx.ParsedTarget.String()
	} else {
		targetStr = "<nil>"
	}

	log.Printf(
		"[parser] request intent parsed: exit=%s, target=%s, remaining=%v",
		exitStr,
		targetStr,
		ctx.RemainingPath,
	)
}
