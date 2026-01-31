package relay

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HTTPRelay handles prefix-style URL relaying:
//
//	/https://example.com/path
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outReq, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		target.String(),
		r.Body,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Clone headers explicitly
	outReq.Header = r.Header.Clone()

	resp, err := h.Client.Do(outReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	// Stream response body
	io.Copy(w, resp.Body)
}

func extractTargetURL(r *http.Request) (*url.URL, error) {
	raw := strings.TrimPrefix(r.URL.Path, "/")

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
	for k, v := range src {
		dst[k] = v
	}
}
