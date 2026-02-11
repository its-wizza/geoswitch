package provider

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("no handler for exit '%s'", exitName)
	}
	return h, nil
}
