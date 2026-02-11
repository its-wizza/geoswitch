package provider

import (
	"context"
	"net/http"

	"geoswitch/internal/config"
)

type ExitHandlerProvider interface {
	GetHandler(ctx context.Context, exitName string, cfg config.ExitConfig) (http.Handler, error)
}

type exitRuntime struct {
	handler http.Handler
}
