package handler

import (
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/deploy"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager            packager.PackagerServiceClient
	Builder             builder.BuilderServiceClient
	Deployer            deployer.DeployerServiceClient
	GitHubApp           *ghpkg.App // for minting installation tokens (repo access)
	DeployTracker       *deploy.Tracker
	RegistryPushURL     string // for builder push, e.g. "localhost:5000"
	RegistryImagePrefix string // for image refs in values.yaml, e.g. cluster-internal address
	WorkloadDomain      string // base domain for platform-generated domains (e.g., "lucity.local")
	DomainTarget        string // CNAME target for custom domains (e.g., "lb.lucity.app")
}

func New(packagerClient packager.PackagerServiceClient, builderClient builder.BuilderServiceClient, deployerClient deployer.DeployerServiceClient, githubApp *ghpkg.App, registryPushURL, registryImagePrefix, workloadDomain, domainTarget string) *Client {
	return &Client{
		Packager:            packagerClient,
		Builder:             builderClient,
		Deployer:            deployerClient,
		GitHubApp:           githubApp,
		DeployTracker:       deploy.NewTracker(),
		RegistryPushURL:     registryPushURL,
		RegistryImagePrefix: registryImagePrefix,
		WorkloadDomain:      workloadDomain,
		DomainTarget:        domainTarget,
	}
}
