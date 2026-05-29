package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	newdeployer "github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
	"github.com/zeitlos/lucity/services/conductor/internal/gateway"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/source"
)

// TokenRefresher refreshes the Logto access token using a refresh token.
// On success, it also writes updated cookies to the HTTP response.
// Returns the new access token for immediate use.
type TokenRefresher func(ctx context.Context, refreshToken string) (newAccessToken string, err error)

type Client struct {
	Packager       *packager.Client
	Deployer       *deployer.Client
	Cashier        cashier.CashierServiceClient // nil if billing disabled
	Issuer         *auth.Issuer                 // ES256 JWT issuer for gRPC auth (nil = no auth)
	GitHubApp      *ghpkg.App                   // for minting installation tokens (repo access)
	Logto          *logto.Client
	TokenRefresher TokenRefresher // refreshes expired Logto access tokens (nil if not configured)

	// Refactored clients
	directory directory.Interface
	platform  platform.Interface
	buildjob  buildjob.Interface
	planner   planner.Interface
	source    source.Interface
	hostname  *hostname.Client
	gateway   *gateway.Client
	deployer  newdeployer.Interface

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
	RegistryPullSecret  authn.Keychain
	RegistryImagePrefix string // for image refs in values.yaml, e.g. cluster-internal address
	WorkloadDomain      string // base domain for platform-generated domains (e.g., "lucity.app")
	DomainTarget        string // CNAME target for custom domains (e.g., "lb.lucity.app")
	IPAddress           string // load balancer IP for A record config
	GitHubAppSlug       string // GitHub App slug for installation URL generation
	DashboardURL        string // base URL for the dashboard (e.g., "http://localhost:5173")
}

func New(packager *packager.Client, legacyDeployer *deployer.Client, cashier cashier.CashierServiceClient, issuer *auth.Issuer, githubApp *ghpkg.App, logto *logto.Client, tokenRefresher TokenRefresher, directory directory.Interface, platform platform.Interface, buildjob buildjob.Interface, planner planner.Interface, source source.Interface, hostname *hostname.Client, gateway *gateway.Client, deployer newdeployer.Interface, config Config) *Client {
	return &Client{
		Packager:       packager,
		Deployer:       legacyDeployer,
		Cashier:        cashier,
		Issuer:         issuer,
		GitHubApp:      githubApp,
		Logto:          logto,
		TokenRefresher: tokenRefresher,
		Config:         config,
		orgIDCache:     make(map[string]string),
		directory:      directory,
		platform:       platform,
		buildjob:       buildjob,
		planner:        planner,
		source:         source,
		hostname:       hostname,
		gateway:        gateway,
		deployer:       deployer,
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
