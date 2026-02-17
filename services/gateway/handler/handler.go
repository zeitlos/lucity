package handler

import (
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager    packager.PackagerServiceClient
	Builder     builder.BuilderServiceClient
	Deployer    deployer.DeployerServiceClient
	RegistryURL string // e.g., "localhost:5000"
}

func New(packagerClient packager.PackagerServiceClient, builderClient builder.BuilderServiceClient, deployerClient deployer.DeployerServiceClient, registryURL string) *Client {
	return &Client{
		Packager:    packagerClient,
		Builder:     builderClient,
		Deployer:    deployerClient,
		RegistryURL: registryURL,
	}
}
