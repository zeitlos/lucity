package handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/deploy"
	"github.com/zeitlos/lucity/services/gateway/logto"
)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager            packager.PackagerServiceClient
	Builder             builder.BuilderServiceClient
	Deployer            deployer.DeployerServiceClient
	Cashier             cashier.CashierServiceClient // nil if billing disabled
	GitHubApp           *ghpkg.App                   // for minting installation tokens (repo access)
	Logto               *logto.Client
	DeployTracker       *deploy.Tracker
	RegistryPushURL     string // for builder push, e.g. "localhost:5000"
	RegistryImagePrefix string // for image refs in values.yaml, e.g. cluster-internal address
	WorkloadDomain      string // base domain for platform-generated domains (e.g., "lucity.local")
	DomainTarget        string // CNAME target for custom domains (e.g., "lb.lucity.app")
	GitHubAppSlug       string // GitHub App slug for installation URL generation

	// Cached Logto org role IDs (looked up by name on first use)
	orgRoleOnce sync.Once
	adminRoleID  string
	memberRoleID string

	// In-memory cache: workspace ID (org name) → Logto org ID
	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex
}

func New(packagerClient packager.PackagerServiceClient, builderClient builder.BuilderServiceClient, deployerClient deployer.DeployerServiceClient, cashierClient cashier.CashierServiceClient, githubApp *ghpkg.App, logtoClient *logto.Client, registryPushURL, registryImagePrefix, workloadDomain, domainTarget, githubAppSlug string) *Client {
	return &Client{
		Packager:            packagerClient,
		Builder:             builderClient,
		Deployer:            deployerClient,
		Cashier:             cashierClient,
		GitHubApp:           githubApp,
		Logto:               logtoClient,
		DeployTracker:       deploy.NewTracker(),
		RegistryPushURL:     registryPushURL,
		RegistryImagePrefix: registryImagePrefix,
		WorkloadDomain:      workloadDomain,
		DomainTarget:        domainTarget,
		GitHubAppSlug:       githubAppSlug,
		orgIDCache:          make(map[string]string),
	}
}

// orgRoleIDs returns the cached admin and member role IDs, looking them up on first call.
func (c *Client) orgRoleIDs(ctx context.Context) (adminID, memberID string, err error) {
	c.orgRoleOnce.Do(func() {
		if c.Logto == nil {
			err = fmt.Errorf("logto not configured")
			return
		}
		roles, rolesErr := c.Logto.OrganizationRoles(ctx)
		if rolesErr != nil {
			err = fmt.Errorf("failed to fetch organization roles: %w", rolesErr)
			return
		}
		for _, r := range roles {
			switch r.Name {
			case "admin":
				c.adminRoleID = r.ID
			case "member":
				c.memberRoleID = r.ID
			}
		}
		if c.adminRoleID == "" || c.memberRoleID == "" {
			err = fmt.Errorf("missing org roles: admin=%q member=%q", c.adminRoleID, c.memberRoleID)
			return
		}
		slog.Info("logto org roles cached", "admin", c.adminRoleID, "member", c.memberRoleID)
	})
	if err != nil {
		// Reset so next call retries
		c.orgRoleOnce = sync.Once{}
		return "", "", err
	}
	return c.adminRoleID, c.memberRoleID, nil
}
