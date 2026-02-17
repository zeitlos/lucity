package handler

import (
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/packager"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager    packager.PackagerServiceClient
	Builder     builder.BuilderServiceClient
	RegistryURL string // e.g., "ghcr.io"
}

func New(packagerClient packager.PackagerServiceClient, builderClient builder.BuilderServiceClient, registryURL string) *Client {
	return &Client{
		Packager:    packagerClient,
		Builder:     builderClient,
		RegistryURL: registryURL,
	}
}
