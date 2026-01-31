package main

import (
	"log"

	"geoswitch/internal/server"
)

func main() {
	addr := ":8080"
	log.Printf("GeoSwitch starting on %s\n", addr)

	if err := server.Start(addr); err != nil {
		log.Fatal(err)
	}
}
