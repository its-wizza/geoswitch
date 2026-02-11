package provider

import (
	"context"
	"log"
	"net/http"
	"sync"

	"geoswitch/internal/config"
	"geoswitch/internal/proxy"
)

type GluetunProvider struct {
	mu       sync.Mutex
	runtimes map[string]*exitRuntime
	// TODO: Add docker client and network when implementing container management
}

func (p *GluetunProvider) GetHandler(
	ctx context.Context,
	exitName string,
	cfg config.ExitConfig,
) (http.Handler, error) {

	p.mu.Lock()
	defer p.mu.Unlock()

	if rt, ok := p.runtimes[exitName]; ok {
		return rt.handler, nil
	}

	// TODO: start container and get its proxy address
	// For now, create a placeholder proxy handler
	log.Printf("would start Gluetun container for exit %s", exitName)
	handler := proxy.NewReverseProxy()

	rt := &exitRuntime{
		handler: handler,
	}

	p.runtimes[exitName] = rt
	return handler, nil
}
