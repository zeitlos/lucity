package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/handler"
	"github.com/zeitlos/lucity/pkg/logto"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// OIDC (PKCE — no client secret needed)
	OIDCIssuerURL    string `envconfig:"OIDC_ISSUER_URL" required:"true"`
	OIDCDiscoveryURL string `envconfig:"OIDC_DISCOVERY_URL"` // internal URL for discovery/token exchange (avoids hairpin routing)
	OIDCClientID     string `envconfig:"OIDC_CLIENT_ID" required:"true"`
	OIDCCallbackURL  string `envconfig:"OIDC_CALLBACK_URL" default:"http://localhost:8080/auth/callback"`

	// Auth
	DashboardURL   string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`
	SessionSecret  string `envconfig:"SESSION_SECRET" required:"true"` // HS256 secret for signing session JWTs
	AuthTestSecret string `envconfig:"AUTH_TEST_SECRET"`               // HS256 test secret for dev/test tokens (never set in production)

	// Logto Management API (M2M)
	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`    // e.g. "https://id.lucity.cloud"
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`

	// Backend services
	BuilderAddr  string `envconfig:"BUILDER_ADDR" default:"localhost:9001"`
	PackagerAddr string `envconfig:"PACKAGER_ADDR" default:"localhost:9002"`
	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	// Registry
	RegistryURL         string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryImagePrefix string `envconfig:"REGISTRY_IMAGE_PREFIX"` // cluster-internal address for image refs; defaults to REGISTRY_URL

	// GitHub App (for installation tokens + OAuth)
	GitHubAppID            int64  `envconfig:"GITHUB_APP_ID"`
	GitHubPrivateKeyPath   string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`
	GitHubClientID         string `envconfig:"GITHUB_CLIENT_ID"`
	GitHubClientSecret     string `envconfig:"GITHUB_CLIENT_SECRET"`
	GitHubOAuthCallbackURL string `envconfig:"GITHUB_OAUTH_CALLBACK_URL" default:"http://localhost:8080/auth/github/callback"`

	// Domains
	WorkloadDomain string `envconfig:"WORKLOAD_DOMAIN" default:"lucity.local"`
	DomainTarget   string `envconfig:"DOMAIN_TARGET"`  // CNAME target for custom domains (e.g., lb.lucity.app)
	IPAddress      string `envconfig:"IP_ADDRESS"`     // LB IP for A record config (e.g., 46.225.47.40)

	// Billing (optional — disabled when not configured)
	CashierAddr string `envconfig:"CASHIER_ADDR"`

	// GitHub App (for workspace installation linking)
	GitHubAppSlug string `envconfig:"GITHUB_APP_SLUG"` // e.g. "lucity-dev"

	// Internal JWT (ES256 for gRPC service-to-service auth)
	InternalJWTPrivateKeyPath string `envconfig:"INTERNAL_JWT_PRIVATE_KEY_PATH"` // PEM file; optional — legacy metadata headers used when not set
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	ctx, cancel := graceful.Context()
	defer cancel()

	oidcProvider, err := NewOIDCProvider(ctx, config.OIDCIssuerURL, config.OIDCDiscoveryURL, config.OIDCClientID, config.OIDCCallbackURL)
	if err != nil {
		slog.Error("failed to initialize OIDC provider", "error", err)
		os.Exit(1)
	}

	// Create JWT verifier for session tokens.
	// Session JWTs are HMAC-signed by the gateway, not Logto JWKS.
	// The OIDC verifier acts as an optional primary (for future JWT access tokens),
	// with the HMAC validator as fallback for session cookies.
	verifier, err := auth.NewVerifier(ctx, config.OIDCIssuerURL, config.OIDCClientID)
	if err != nil {
		slog.Error("failed to create JWT verifier", "error", err)
		os.Exit(1)
	}

	sessionSecret := config.SessionSecret
	if config.AuthTestSecret != "" {
		sessionSecret = config.AuthTestSecret
		slog.Warn("test token authentication enabled — do not use in production")
	}
	verifier = verifier.WithFallback(hmacValidateFunc(sessionSecret))

	// Connect to builder
	builderConn, err := grpc.NewClient(config.BuilderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to builder", "error", err, "addr", config.BuilderAddr)
		os.Exit(1)
	}
	defer builderConn.Close()

	builderClient := builder.NewBuilderServiceClient(builderConn)

	// Connect to packager
	packagerConn, err := grpc.NewClient(config.PackagerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to packager", "error", err, "addr", config.PackagerAddr)
		os.Exit(1)
	}
	defer packagerConn.Close()

	packagerClient := packager.NewPackagerServiceClient(packagerConn)

	// Connect to deployer
	deployerConn, err := grpc.NewClient(config.DeployerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to deployer", "error", err, "addr", config.DeployerAddr)
		os.Exit(1)
	}
	defer deployerConn.Close()

	deployerClient := deployer.NewDeployerServiceClient(deployerConn)

	// Initialize GitHub App for installation tokens (optional — repo features disabled without it)
	var githubApp *ghpkg.App
	if config.GitHubAppID != 0 && config.GitHubPrivateKeyPath != "" {
		var err error
		githubApp, err = ghpkg.NewApp(config.GitHubAppID, config.GitHubClientID, config.GitHubClientSecret, "", config.GitHubOAuthCallbackURL, config.GitHubPrivateKeyPath)
		if err != nil {
			slog.Error("failed to create github app", "error", err)
			os.Exit(1)
		}
		slog.Info("github app initialized", "app_id", config.GitHubAppID)
	} else {
		slog.Info("github app not configured — repo listing and commit enrichment disabled")
	}

	registryImagePrefix := config.RegistryImagePrefix
	if registryImagePrefix == "" {
		registryImagePrefix = config.RegistryURL
	}

	domainTarget := config.DomainTarget
	if domainTarget == "" {
		domainTarget = "lb." + config.WorkloadDomain
	}

	// Connect to cashier (optional — billing disabled without it)
	var cashierClient cashier.CashierServiceClient
	var cashierConn *grpc.ClientConn
	if config.CashierAddr != "" {
		var err error
		cashierConn, err = grpc.NewClient(config.CashierAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			slog.Error("failed to connect to cashier", "error", err, "addr", config.CashierAddr)
			os.Exit(1)
		}
		defer cashierConn.Close()
		cashierClient = cashier.NewCashierServiceClient(cashierConn)
		slog.Info("billing enabled", "addr", config.CashierAddr)
	} else {
		slog.Info("cashier not configured — billing disabled")
	}

	// Initialize internal JWT issuer for gRPC service-to-service auth (optional)
	var internalIssuer *auth.Issuer
	if config.InternalJWTPrivateKeyPath != "" {
		var err error
		internalIssuer, err = auth.NewIssuerFromFile(config.InternalJWTPrivateKeyPath)
		if err != nil {
			slog.Error("failed to create internal JWT issuer", "error", err)
			os.Exit(1)
		}
		slog.Info("internal JWT issuer initialized (ES256)")
	} else {
		slog.Warn("internal JWT not configured — gRPC calls use legacy plain metadata headers")
	}

	// Initialize Logto client for workspace/member management
	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	secure := secureCookies(config.DashboardURL)
	tokenRefresher := newTokenRefresher(oidcProvider, secure)

	api := handler.New(packagerClient, builderClient, deployerClient, cashierClient, internalIssuer, githubApp, logtoClient, tokenRefresher, config.RegistryURL, registryImagePrefix, config.WorkloadDomain, domainTarget, config.IPAddress, config.GitHubAppSlug, config.DashboardURL)

	components := []grpcComponent{
		{name: "builder", conn: builderConn},
		{name: "packager", conn: packagerConn},
		{name: "deployer", conn: deployerConn},
	}
	if cashierConn != nil {
		components = append(components, grpcComponent{name: "cashier", conn: cashierConn})
	}

	graphqlServer := NewGraphQLServer(config.Port, api, oidcProvider, verifier, logtoClient, internalIssuer, sessionSecret, config.DashboardURL, config.GitHubAppSlug, components)

	graceful.Serve(ctx, graphqlServer)
}
