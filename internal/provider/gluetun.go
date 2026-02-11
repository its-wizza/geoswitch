package provider

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"

	"geoswitch/internal/config"
	"geoswitch/internal/proxy"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type GluetunProvider struct {
	mu       sync.Mutex
	runtimes map[string]*exitRuntime
	docker   *client.Client
	network  string
}

func (p *GluetunProvider) GetHandler(
	ctx context.Context,
	exitName string,
	cfg config.ExitConfig,
) (http.Handler, error) {

	p.mu.Lock()
	defer p.mu.Unlock()

	if rt, ok := p.runtimes[exitName]; ok {
		log.Printf("[gluetun] reusing cached handler for exit '%s'", exitName)
		return rt.handler, nil
	}

	log.Printf("[gluetun] creating new handler for exit '%s' (country=%s)", exitName, cfg.Country)

	if err := p.ensureNetwork(ctx); err != nil {
		log.Printf("[gluetun] failed to ensure network: %v", err)
		return nil, err
	}

	containerName := "gluetun-" + exitName

	// Check if container already exists
	_, err := p.docker.ContainerInspect(ctx, containerName)
	if err != nil {
		log.Printf("[gluetun] container '%s' does not exist, creating it", containerName)
		if err := p.createContainer(ctx, containerName, cfg); err != nil {
			log.Printf("[gluetun] failed to create container '%s': %v", containerName, err)
			return nil, err
		}
	} else {
		log.Printf("[gluetun] reusing existing container '%s'", containerName)
	}

	// Create reverse proxy
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   containerName + ":8888",
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	handler := proxy.NewReverseProxy(proxy.WithTransport(transport))

	rt := &exitRuntime{
		handler: handler,
	}

	p.runtimes[exitName] = rt
	log.Printf("[gluetun] handler created and cached for exit '%s'", exitName)
	return handler, nil
}

func (p *GluetunProvider) ensureNetwork(ctx context.Context) error {
	log.Printf("[gluetun] ensuring network '%s' exists", p.network)
	_, err := p.docker.NetworkInspect(ctx, p.network, network.InspectOptions{})
	if err == nil {
		log.Printf("[gluetun] network '%s' already exists", p.network)
		return nil
	}

	log.Printf("[gluetun] creating network '%s'", p.network)
	_, err = p.docker.NetworkCreate(ctx, p.network, network.CreateOptions{})
	if err != nil {
		log.Printf("[gluetun] failed to create network '%s': %v", p.network, err)
	}
	return err
}

func (p *GluetunProvider) createContainer(
	ctx context.Context,
	name string,
	cfg config.ExitConfig,
) error {

	log.Printf("[gluetun] creating container '%s' with gluetun:latest", name)

	env := []string{
		"HTTPPROXY=on",
		"SERVER_COUNTRIES=" + cfg.Country,
		// Provider-specific variables later
	}

	resp, err := p.docker.ContainerCreate(
		ctx,
		&container.Config{
			Image: "qmcgaw/gluetun:latest",
			Env:   env,
			ExposedPorts: nat.PortSet{
				"8888/tcp": struct{}{},
			},
		},
		&container.HostConfig{
			AutoRemove: false,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				p.network: {},
			},
		},
		nil,
		name,
	)
	if err != nil {
		log.Printf("[gluetun] failed to create container '%s': %v", name, err)
		return err
	}

	log.Printf("[gluetun] starting container '%s' (ID: %s)", name, resp.ID)
	err = p.docker.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("[gluetun] failed to start container '%s': %v", name, err)
	}
	return err
}

func NewGluetunProvider(network string) (*GluetunProvider, error) {
	log.Printf("[gluetun] initializing GluetunProvider with network '%s'", network)
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Printf("[gluetun] failed to create docker client: %v", err)
		return nil, err
	}

	log.Printf("[gluetun] docker client initialized successfully")
	return &GluetunProvider{
		runtimes: make(map[string]*exitRuntime),
		docker:   cli,
		network:  network,
	}, nil
}
