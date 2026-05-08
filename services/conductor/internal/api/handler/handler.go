package handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/api/deploy"
	inprocbuilder "github.com/zeitlos/lucity/services/conductor/internal/inproc/builder"
	inprocdeployer "github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	inprocpackager "github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
)

// TokenRefresher refreshes the Logto access token using a refresh token.
// On success, it also writes updated cookies to the HTTP response.
// Returns the new access token for immediate use.
type TokenRefresher func(ctx context.Context, refreshToken string) (newAccessToken string, err error)

// Client holds all dependencies for the gateway's business logic.
type Client struct {
	Packager            *inprocpackager.Server
	Builder             *inprocbuilder.Server
	Deployer            *inprocdeployer.Server
	Cashier             cashier.CashierServiceClient // nil if billing disabled
	Issuer              *auth.Issuer                 // ES256 JWT issuer for gRPC auth (nil = no auth)
	GitHubApp           *ghpkg.App                   // for minting installation tokens (repo access)
	Logto               *logto.Client
	DeployTracker       *deploy.Tracker
	TokenRefresher      TokenRefresher // refreshes expired Logto access tokens (nil if not configured)
	RegistryPushURL     string         // for builder push, e.g. "localhost:5000"
	RegistryImagePrefix string         // for image refs in values.yaml, e.g. cluster-internal address
	WorkloadDomain      string         // base domain for platform-generated domains (e.g., "lucity.app")
	DomainTarget        string         // CNAME target for custom domains (e.g., "lb.lucity.app")
	IPAddress           string         // load balancer IP for A record config
	GitHubAppSlug       string         // GitHub App slug for installation URL generation
	DashboardURL        string         // base URL for the dashboard (e.g., "http://localhost:5173")

	// Cached Logto org role IDs (looked up by name on first use)
	orgRoleOnce  sync.Once
	adminRoleID  string
	memberRoleID string

	// In-memory cache: workspace ID (org name) → Logto org ID
	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex
}

func New(packagerSvc *inprocpackager.Server, builderSvc *inprocbuilder.Server, deployerSvc *inprocdeployer.Server, cashierClient cashier.CashierServiceClient, issuer *auth.Issuer, githubApp *ghpkg.App, logtoClient *logto.Client, tokenRefresher TokenRefresher, registryPushURL, registryImagePrefix, workloadDomain, domainTarget, ipAddress, githubAppSlug, dashboardURL string) *Client {
	return &Client{
		Packager:            packagerSvc,
		Builder:             builderSvc,
		Deployer:            deployerSvc,
		Cashier:             cashierClient,
		Issuer:              issuer,
		GitHubApp:           githubApp,
		Logto:               logtoClient,
		DeployTracker:       deploy.NewTracker(),
		TokenRefresher:      tokenRefresher,
		RegistryPushURL:     registryPushURL,
		RegistryImagePrefix: registryImagePrefix,
		WorkloadDomain:      workloadDomain,
		DomainTarget:        domainTarget,
		IPAddress:           ipAddress,
		GitHubAppSlug:       githubAppSlug,
		DashboardURL:        dashboardURL,
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
