package conductor

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
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
	"github.com/zeitlos/lucity/services/conductor/internal/inproc/builder"
	"github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
)

// TokenRefresher refreshes the Logto access token using a refresh token.
// On success, it also writes updated cookies to the HTTP response.
// Returns the new access token for immediate use.
type TokenRefresher func(ctx context.Context, refreshToken string) (newAccessToken string, err error)

type Client struct {
	Packager       *packager.Client
	Builder        *builder.Client
	Deployer       *deployer.Client
	Cashier        cashier.CashierServiceClient // nil if billing disabled
	Issuer         *auth.Issuer                 // ES256 JWT issuer for gRPC auth (nil = no auth)
	GitHubApp      *ghpkg.App                   // for minting installation tokens (repo access)
	Logto          *logto.Client
	DeployTracker  *deploy.Tracker
	TokenRefresher TokenRefresher // refreshes expired Logto access tokens (nil if not configured)

	// Refactored clients
	directory directory.Interface

	Config Config

	// Cached Logto org role IDs (looked up by name on first use)
	orgRoleOnce  sync.Once
	adminRoleID  string
	memberRoleID string

	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex
}

type Config struct {
	RegistryPushURL     string // for builder push, e.g. "localhost:5000"
	RegistryImagePrefix string // for image refs in values.yaml, e.g. cluster-internal address
	WorkloadDomain      string // base domain for platform-generated domains (e.g., "lucity.app")
	DomainTarget        string // CNAME target for custom domains (e.g., "lb.lucity.app")
	IPAddress           string // load balancer IP for A record config
	GitHubAppSlug       string // GitHub App slug for installation URL generation
	DashboardURL        string // base URL for the dashboard (e.g., "http://localhost:5173")
}

func New(packager *packager.Client, builder *builder.Client, deployer *deployer.Client, cashier cashier.CashierServiceClient, issuer *auth.Issuer, githubApp *ghpkg.App, logto *logto.Client, tokenRefresher TokenRefresher, config Config) *Client {
	return &Client{
		Packager:       packager,
		Builder:        builder,
		Deployer:       deployer,
		Cashier:        cashier,
		Issuer:         issuer,
		GitHubApp:      githubApp,
		Logto:          logto,
		DeployTracker:  deploy.NewTracker(),
		TokenRefresher: tokenRefresher,
		Config:         config,
		orgIDCache:     make(map[string]string),
	}
}

// orgRoleIDs returns the cached admin and member role IDs, looking them up on first call.
func (c *Client) orgRoleIDs(ctx context.Context) (adminID, memberID string, err error) {
	c.orgRoleOnce.Do(func() {
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
