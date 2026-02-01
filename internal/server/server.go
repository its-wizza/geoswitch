package server

import (
	"net/http"
	"time"
)

func New(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       30 * time.Second, // Timeout for reading request
		ReadHeaderTimeout: 10 * time.Second, // Timeout for reading headers
		IdleTimeout:       90 * time.Second, // Close idle connections
		WriteTimeout:      0,                // Disabled for unbounded writes (media streaming)
	}
}
