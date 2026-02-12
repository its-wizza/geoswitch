package provider

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"geoswitch/internal/config"
	"geoswitch/internal/proxy"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type exitRuntime struct {
	handler       http.Handler
	containerID   string
	containerName string
}

type gluetunConfig struct {
	network      string
	imageVersion string
}

// GluetunOption is a functional option for configuring a GluetunProvider.
type GluetunOption func(*gluetunConfig)

// WithNetwork sets a custom Docker network name.
// If not provided, defaults to "gluetun".
func WithNetwork(network string) GluetunOption {
	return func(c *gluetunConfig) {
		c.network = network
	}
}

// WithImageVersion sets a specific Gluetun image version.
// If not provided, defaults to "qmcgaw/gluetun:latest".
func WithImageVersion(version string) GluetunOption {
	return func(c *gluetunConfig) {
		c.imageVersion = version
	}
}

type GluetunProvider struct {
	mu       sync.Mutex
	runtimes map[string]*exitRuntime
	docker   *client.Client
	network  string
	image    string
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

		inspect, err := p.docker.ContainerInspect(ctx, rt.containerID)
		if err != nil {
			return nil, err
		}

		if inspect.State.Health != nil &&
			inspect.State.Health.Status != "healthy" {
			return nil, fmt.Errorf("exit '%s' not healthy", exitName)
		}

		return rt.handler, nil
	}

	log.Printf("[gluetun] creating new handler for exit '%s' (country=%s)", exitName, cfg.Country)

	if err := p.ensureNetwork(ctx); err != nil {
		log.Printf("[gluetun] failed to ensure network: %v", err)
		return nil, err
	}

	containerName := "gluetun-" + exitName
	var containerID string

	// Check if container already exists
	resp, err := p.docker.ContainerInspect(ctx, containerName)
	if err != nil {
		log.Printf("[gluetun] container '%s' does not exist, creating it", containerName)
		// Pull image if it doesn't exist
		if err := p.ensureImage(ctx); err != nil {
			return nil, err
		}
		// Create and start container
		containerID, err = p.createContainer(ctx, containerName, cfg)
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("[gluetun] reusing existing container '%s'", containerName)
		containerID = resp.ID
	}

	// Write logs to main logger
	p.streamLogs(ctx, containerID)

	// Wait for container to become healthy
	if err := p.waitForHealthy(ctx, containerName, 60*time.Second); err != nil {
		log.Printf("[gluetun] container '%s' failed health check: %v", containerName, err)
		return nil, err
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
		handler:       handler,
		containerID:   containerID,
		containerName: containerName,
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

func (p *GluetunProvider) ensureImage(ctx context.Context) error {
	_, err := p.docker.ImageInspect(ctx, p.image)
	if err == nil {
		return nil // image already exists
	}

	log.Printf("[gluetun] pulling image %s", p.image)

	reader, err := p.docker.ImagePull(ctx, p.image, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Drain output (required or pull won't complete properly)
	_, err = io.Copy(io.Discard, reader)
	return err
}

func (p *GluetunProvider) createContainer(
	ctx context.Context,
	name string,
	cfg config.ExitConfig,
) (string, error) {

	log.Printf("[gluetun] creating container '%s' with %s", name, p.image)

	env := []string{
		"HTTPPROXY=on",
		"SERVER_COUNTRIES=" + cfg.Country,
		// Provider-specific variables later
	}

	resp, err := p.docker.ContainerCreate(
		ctx,
		&container.Config{
			Image: p.image,
			Env:   env,
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
		return "", err
	}

	log.Printf("[gluetun] starting container '%s' (ID: %s)", name, resp.ID)
	err = p.docker.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("[gluetun] failed to start container '%s': %v", name, err)
		return "", err
	}

	return resp.ID, nil
}

func (p *GluetunProvider) streamLogs(ctx context.Context, containerID string) {
	go func() {
		// Use background context for log streaming to prevent early cancellation
		logCtx := context.Background()
		reader, err := p.docker.ContainerLogs(logCtx, containerID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			log.Printf("[gluetun] log stream error: %v", err)
			return
		}
		defer reader.Close()

		io.Copy(log.Writer(), reader)
	}()
}

func (p *GluetunProvider) waitForHealthy(
	ctx context.Context,
	containerName string,
	timeout time.Duration,
) error {

	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("container %s did not become healthy in time", containerName)
		}

		inspect, err := p.docker.ContainerInspect(ctx, containerName)
		if err != nil {
			return err
		}

		if inspect.State == nil || inspect.State.Health == nil {
			return fmt.Errorf("container %s has no healthcheck configured", containerName)
		}

		status := inspect.State.Health.Status

		switch status {
		case "healthy":
			log.Printf("[gluetun] container '%s' is healthy", containerName)
			return nil

		case "unhealthy":
			return fmt.Errorf("container %s is unhealthy", containerName)

		case "starting":
			time.Sleep(1 * time.Second)
		}
	}
}

// Close cleans up all resources including stopping containers and removing the network.
func (p *GluetunProvider) Close(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("[gluetun] shutting down provider, cleaning up %d runtimes", len(p.runtimes))

	// Stop all running containers
	for exitName, rt := range p.runtimes {
		log.Printf("[gluetun] stopping container '%s' for exit '%s'", rt.containerName, exitName)
		stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := p.docker.ContainerStop(stopCtx, rt.containerID, container.StopOptions{})
		cancel()
		if err != nil {
			log.Printf("[gluetun] error stopping container '%s': %v", rt.containerName, err)
		}
	}

	p.runtimes = make(map[string]*exitRuntime)

	// Remove network
	log.Printf("[gluetun] removing network '%s'", p.network)
	if err := p.docker.NetworkRemove(ctx, p.network); err != nil {
		log.Printf("[gluetun] error removing network '%s': %v", p.network, err)
	}

	// Close Docker client
	log.Printf("[gluetun] closing docker client")
	return p.docker.Close()
}

func NewGluetunProvider(opts ...GluetunOption) (*GluetunProvider, error) {
	config := &gluetunConfig{
		network:      "gluetun",
		imageVersion: "qmcgaw/gluetun:latest",
	}

	for _, opt := range opts {
		opt(config)
	}

	log.Printf("[gluetun] initialising GluetunProvider with network '%s'", config.network)
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Printf("[gluetun] failed to create docker client: %v", err)
		return nil, err
	}

	log.Printf("[gluetun] docker client initialised successfully")
	return &GluetunProvider{
		runtimes: make(map[string]*exitRuntime),
		docker:   cli,
		network:  config.network,
		image:    config.imageVersion,
	}, nil
}
