package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/cashier"
	"github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/handler"
	"github.com/zeitlos/lucity/services/gateway/rauthy"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// OIDC (PKCE — no client secret needed)
	OIDCIssuerURL   string `envconfig:"OIDC_ISSUER_URL" required:"true"`
	OIDCClientID    string `envconfig:"OIDC_CLIENT_ID" required:"true"`
	OIDCCallbackURL string `envconfig:"OIDC_CALLBACK_URL" default:"http://localhost:8080/auth/callback"`

	// Auth
	JWTSecret    string `envconfig:"JWT_SECRET" required:"true"`
	DashboardURL string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`

	// Backend services
	BuilderAddr  string `envconfig:"BUILDER_ADDR" default:"localhost:9001"`
	PackagerAddr string `envconfig:"PACKAGER_ADDR" default:"localhost:9002"`
	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	// Registry
	RegistryURL         string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryImagePrefix string `envconfig:"REGISTRY_IMAGE_PREFIX"` // cluster-internal address for image refs; defaults to REGISTRY_URL

	// GitHub App (for installation tokens — repo access)
	GitHubAppID          int64  `envconfig:"GITHUB_APP_ID"`
	GitHubPrivateKeyPath string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`

	// Domains
	WorkloadDomain string `envconfig:"WORKLOAD_DOMAIN" default:"lucity.local"`
	DomainTarget   string `envconfig:"DOMAIN_TARGET"` // CNAME target for custom domains (e.g., lb.lucity.app)

	// Billing (optional — disabled when not configured)
	CashierAddr string `envconfig:"CASHIER_ADDR"`

	// Rauthy admin API (for workspace/member management)
	RauthyAPIURL string `envconfig:"RAUTHY_API_URL"` // e.g. "https://id.lucity.cloud/auth/v1"
	RauthyAPIKey string `envconfig:"RAUTHY_API_KEY"`

	// GitHub App (for workspace installation linking)
	GitHubAppSlug string `envconfig:"GITHUB_APP_SLUG"` // e.g. "lucity-dev"
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

	oidcProvider, err := NewOIDCProvider(ctx, config.OIDCIssuerURL, config.OIDCClientID, config.OIDCCallbackURL)
	if err != nil {
		slog.Error("failed to initialize OIDC provider", "error", err)
		os.Exit(1)
	}

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
		githubApp, err = ghpkg.NewApp(config.GitHubAppID, "", "", "", "", config.GitHubPrivateKeyPath)
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
	if config.CashierAddr != "" {
		cashierConn, err := grpc.NewClient(config.CashierAddr,
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

	// Initialize Rauthy client for workspace/member management (optional)
	var rauthyClient *rauthy.Client
	if config.RauthyAPIURL != "" && config.RauthyAPIKey != "" {
		rauthyClient = rauthy.New(config.RauthyAPIURL, config.RauthyAPIKey)
		slog.Info("rauthy admin API configured", "url", config.RauthyAPIURL)
	} else {
		slog.Info("rauthy admin API not configured — workspace management disabled")
	}

	api := handler.New(packagerClient, builderClient, deployerClient, cashierClient, githubApp, rauthyClient, config.RegistryURL, registryImagePrefix, config.WorkloadDomain, domainTarget, config.GitHubAppSlug)
	graphqlServer := NewGraphQLServer(config.Port, api, oidcProvider, config.JWTSecret, config.DashboardURL, config.GitHubAppSlug)

	graceful.Serve(ctx, graphqlServer)
}
