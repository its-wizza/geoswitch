package server

import (
	"fmt"
	"net/http"
	"strings"
)

func Start(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Expect /{country}/anything
		path := strings.TrimPrefix(r.URL.Path, "/")
		parts := strings.SplitN(path, "/", 2)

		country := "unknown"
		if len(parts) > 0 && parts[0] != "" {
			country = parts[0]
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "GeoSwitch alive\nCountry: %s\n", country)
	})

	return http.ListenAndServe(addr, mux)
}
