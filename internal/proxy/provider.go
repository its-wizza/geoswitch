package proxy

import (
	"fmt"
	"net/http"
)

type ExitHandlerProvider interface {
	GetHandler(ctx *ParsedRequest, exitName string, cfg ExitConfig) (http.Handler, error)
}

type StaticProvider struct {
	Handlers map[string]http.Handler
}

func (p *StaticProvider) GetHandler(
	_ *ParsedRequest,
	exitName string,
	_ ExitConfig,
) (http.Handler, error) {
	h, ok := p.Handlers[exitName]
	if !ok {
		return nil, fmt.Errorf("no handler for exit '%s'", exitName)
	}
	return h, nil
}

type GluetunProvider struct {
	// docker client
	// container cache
	// locks
}
