package proxy

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type ExitHandlerProvider interface {
	GetHandler(ctx *RequestContext, exitName string, cfg ExitConfig) (http.Handler, error)
}

type StaticProvider struct {
	Handlers map[string]http.Handler
}

func (p *StaticProvider) GetHandler(
	_ *RequestContext,
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
	mu       sync.Mutex
	runtimes map[string]*exitRuntime
}

func (p *GluetunProvider) GetHandler(
	ctx context.Context,
	exitName string,
	cfg ExitConfig,
) (http.Handler, error) {

	p.mu.Lock()
	defer p.mu.Unlock()

	if rt, ok := p.runtimes[exitName]; ok {
		return rt.handler, nil
	}

	// TODO: start container (stub for now)
	handler := NewReverseProxy() // placeholder

	rt := &exitRuntime{
		handler: handler,
	}

	p.runtimes[exitName] = rt
	return handler, nil
}

type exitRuntime struct {
	handler http.Handler
}
