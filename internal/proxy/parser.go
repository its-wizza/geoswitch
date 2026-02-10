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

type Exit string

const DefaultExit Exit = "default"

type IntentParser func(*RequestContext) error

func PathIntentParser(ctx *RequestContext) error {
	if ctx.ParsedTarget != nil || len(ctx.RemainingPath) == 0 {
		return nil
	}

	for i := 0; i < len(ctx.RemainingPath); i++ {
		candidate := strings.Join(ctx.RemainingPath[i:], "/")

		if ctx.Original.URL.RawQuery != "" {
			candidate += "?" + ctx.Original.URL.RawQuery
		}

		u, err := url.Parse(candidate)
		if err != nil || !u.IsAbs() {
			continue
		}

		// Found absolute URL
		ctx.ParsedTarget = u

		control := ctx.RemainingPath[:i]

		if ctx.ParsedExit == nil && len(control) > 0 {
			// This parser consumes the first control segment as exit
			exit := Exit(control[0])
			ctx.ParsedExit = &exit
			ctx.RemainingPath = control[1:]
		} else {
			// Exit already set elsewhere, preserve all control segments
			ctx.RemainingPath = control
		}

		var exitStr string
		if ctx.ParsedExit != nil {
			exitStr = string(*ctx.ParsedExit)
		} else {
			exitStr = "<nil>"
		}

		log.Printf(
			"[parser] path intent parsed: exit=%s, target=%s, remaining=%v",
			exitStr,
			u.String(),
			ctx.RemainingPath,
		)

		return nil
	}

	return nil
}

func HeaderExitParser(headerName string) IntentParser {
	return func(ctx *RequestContext) error {
		if ctx.ParsedExit != nil {
			return nil
		}

		val := ctx.Original.Header.Get(headerName)
		if val == "" {
			return nil
		}

		exit := Exit(val)
		ctx.ParsedExit = &exit
		log.Printf("[parser] header exit parser: found exit '%s' from header '%s'", exit, headerName)
		return nil
	}
}

func ParseRequestIntent(r *http.Request, parsers ...IntentParser) (*RequestContext, error) {
	ctx := &RequestContext{
		Original:      r,
		RemainingPath: splitPath(r.URL.Path),
	}

	log.Printf("[parser] parsing request intent: %s %s, path segments: %v", r.Method, r.URL.Path, ctx.RemainingPath)

	// Chain parsers, operating on the request context
	for _, parse := range parsers {
		if err := parse(ctx); err != nil {
			return nil, err
		}
	}

	var exitStr string
	if ctx.ParsedExit != nil {
		exitStr = string(*ctx.ParsedExit)
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

	return ctx, nil
}

// splitPath splits a URL path into its segments, ignoring bordering slashes.
func splitPath(path string) []string {
	// Trim leading and trailing slashes (supports empty segments for double slash)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}
