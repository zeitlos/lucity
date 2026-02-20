package handler

import (
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager            packager.PackagerServiceClient
	Builder             builder.BuilderServiceClient
	Deployer            deployer.DeployerServiceClient
	RegistryPushURL     string // for builder push, e.g. "localhost:5000"
	RegistryImagePrefix string // for image refs in values.yaml, e.g. cluster-internal address
}

func New(packagerClient packager.PackagerServiceClient, builderClient builder.BuilderServiceClient, deployerClient deployer.DeployerServiceClient, registryPushURL, registryImagePrefix string) *Client {
	return &Client{
		Packager:            packagerClient,
		Builder:             builderClient,
		Deployer:            deployerClient,
		RegistryPushURL:     registryPushURL,
		RegistryImagePrefix: registryImagePrefix,
	}
}
