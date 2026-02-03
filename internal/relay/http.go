package relay

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var hopByHopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"TE":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

type HTTPRelay struct {
	Client *http.Client
}

func NewHTTPRelay(client *http.Client) *HTTPRelay {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPRelay{Client: client}
}

func (h *HTTPRelay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := extractTargetURL(r)
	if err != nil {
		http.Error(w, "Invalid target URL: "+err.Error(), http.StatusBadRequest)
		return
	}

	outReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		target.String(),
		r.Body,
	)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers selectively
	copyHeaders(outReq.Header, r.Header)
	outReq.Header.Set("Host", target.Host) // Override with target host
	outReq.RequestURI = ""                 // Must be cleared for client requests

	resp, err := h.Client.Do(outReq)
	if err != nil {
		http.Error(w, "Backend error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Check for copy errors
	if _, err := io.Copy(w, resp.Body); err != nil {
		// Too late to send error, but log it if you have logging
		_ = err
	}
}

func extractTargetURL(r *http.Request) (*url.URL, error) {
	raw := ""
	if r.URL.IsAbs() {
		raw = strings.TrimPrefix(r.URL.String(), "/")
	} else {
		raw = strings.TrimPrefix(r.URL.EscapedPath(), "/")
		if r.URL.RawQuery != "" {
			raw += "?" + r.URL.RawQuery
		}
	}

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return nil, errors.New("target URL must start with http:// or https://")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		// Skip hop-by-hop headers
		if hopByHopHeaders[k] {
			continue
		}

		// Skip sensitive headers
		if k == "Authorization" || k == "Cookie" {
			continue
		}

		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
