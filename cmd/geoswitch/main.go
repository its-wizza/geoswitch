package main

import (
	"log"

	"geoswitch/internal/relay"
	"geoswitch/internal/server"
)

func main() {
	handler := relay.NewHTTPRelay(nil)

	srv := server.New(":8080", handler)

	log.Println("GeoSwitch listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
