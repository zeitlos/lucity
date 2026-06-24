package conductor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/zeitlos/lucity/pkg/cashier"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/directory"
	"github.com/zeitlos/lucity/services/conductor/internal/environment"
	"github.com/zeitlos/lucity/services/conductor/internal/gateway"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
	"github.com/zeitlos/lucity/services/conductor/internal/planner"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/source"
)

// TokenRefresher refreshes the Logto access token using a refresh token.
// On success, it also writes updated cookies to the HTTP response.
// Returns the new access token for immediate use.
type TokenRefresher func(ctx context.Context, refreshToken string) (newAccessToken string, err error)

type Client struct {
	cashier        cashier.CashierServiceClient
	gitHubApp      *ghpkg.App
	logto          *logto.Client
	tokenRefresher TokenRefresher // refreshes expired Logto access tokens (nil if not configured)

	directory   directory.Interface
	platform    platform.Interface
	buildjob    buildjob.Interface
	planner     planner.Interface
	source      source.Interface
	hostname      *hostname.Client
	gateway       *gateway.Client
	deployer      deployer.Interface
	environment   environment.Interface
	objectStorage objectstorage.Interface

	config Config

	// Cached Logto org role IDs (looked up by name on first use)
	orgRoleOnce  sync.Once
	adminRoleID  string
	memberRoleID string

	orgIDCache   map[string]string
	orgIDCacheMu sync.RWMutex
}

type Config struct {
	Version              string
	RegistryPullSecret   authn.Keychain
	RegistryURL          string
	RegistryPushURL      string
	RegistryPullURL      string
	WorkloadDomain       string
	DatabaseDomain       string
	LoadBalancerHostname string
	LoadBalancerIP       string
	GitHubAppSlug        string
	DashboardURL         string
}

func New(cashier cashier.CashierServiceClient, githubApp *ghpkg.App, logto *logto.Client, tokenRefresher TokenRefresher, directory directory.Interface, platform platform.Interface, buildjob buildjob.Interface, planner planner.Interface, source source.Interface, hostname *hostname.Client, gateway *gateway.Client, deployer deployer.Interface, environment environment.Interface, objectStorage objectstorage.Interface, config Config) *Client {
	return &Client{
		cashier:        cashier,
		gitHubApp:      githubApp,
		logto:          logto,
		tokenRefresher: tokenRefresher,
		config:         config,
		orgIDCache:     make(map[string]string),
		directory:      directory,
		platform:       platform,
		buildjob:       buildjob,
		planner:        planner,
		source:         source,
		hostname:       hostname,
		gateway:        gateway,
		deployer:       deployer,
		environment:    environment,
		objectStorage:  objectStorage,
	}
}

// orgRoleIDs returns the cached admin and member role IDs, looking them up on first call.
func (c *Client) orgRoleIDs(ctx context.Context) (adminID, memberID string, err error) {
	c.orgRoleOnce.Do(func() {
		roles, rolesErr := c.logto.OrganizationRoles(ctx)

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
