package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"geoswitch/internal/config"
)

type StaticProvider struct {
	Handlers map[string]http.Handler
}

func (p *StaticProvider) GetHandler(
	_ context.Context,
	exitName string,
	_ config.ExitConfig,
) (http.Handler, error) {
	h, ok := p.Handlers[exitName]
	if !ok {
		log.Printf("[static] no handler found for exit '%s'", exitName)
		return nil, fmt.Errorf("no handler for exit '%s'", exitName)
	}
	log.Printf("[static] returning handler for exit '%s'", exitName)
	return h, nil
}

func NewStaticProvider(handlers map[string]http.Handler) *StaticProvider {
	log.Printf("[static] initializing StaticProvider with %d handlers", len(handlers))
	return &StaticProvider{
		Handlers: handlers,
	}
}
