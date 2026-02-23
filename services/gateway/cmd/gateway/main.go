package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	gh "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// GitHub App
	GitHubAppID          int64  `envconfig:"GITHUB_APP_ID" required:"true"`
	GitHubClientID       string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GitHubClientSecret   string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
	GitHubWebhookSecret  string `envconfig:"GITHUB_WEBHOOK_SECRET" default:"dev-secret"`
	GitHubPrivateKeyPath string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`

	// Auth
	JWTSecret    string `envconfig:"JWT_SECRET" required:"true"`
	DashboardURL string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`
	CallbackURL  string `envconfig:"CALLBACK_URL" default:"http://localhost:8080/auth/github/callback"`

	// Backend services
	BuilderAddr  string `envconfig:"BUILDER_ADDR" default:"localhost:9001"`
	PackagerAddr string `envconfig:"PACKAGER_ADDR" default:"localhost:9002"`
	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	// Registry
	RegistryURL         string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryImagePrefix string `envconfig:"REGISTRY_IMAGE_PREFIX"` // cluster-internal address for image refs; defaults to REGISTRY_URL

	// Domains
	WorkloadDomain string `envconfig:"WORKLOAD_DOMAIN" default:"lucity.local"`
	DomainTarget   string `envconfig:"DOMAIN_TARGET"` // CNAME target for custom domains (e.g., lb.lucity.app)
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

	githubApp, err := gh.NewApp(
		config.GitHubAppID,
		config.GitHubClientID,
		config.GitHubClientSecret,
		config.GitHubWebhookSecret,
		config.CallbackURL,
		config.GitHubPrivateKeyPath,
	)
	if err != nil {
		slog.Error("failed to initialize github app", "error", err)
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

	registryImagePrefix := config.RegistryImagePrefix
	if registryImagePrefix == "" {
		registryImagePrefix = config.RegistryURL
	}

	domainTarget := config.DomainTarget
	if domainTarget == "" {
		domainTarget = "lb." + config.WorkloadDomain
	}

	api := handler.New(packagerClient, builderClient, deployerClient, config.RegistryURL, registryImagePrefix, config.WorkloadDomain, domainTarget)
	graphqlServer := NewGraphQLServer(config.Port, api, githubApp, config.JWTSecret, config.DashboardURL)

	graceful.Serve(ctx, graphqlServer)
}
