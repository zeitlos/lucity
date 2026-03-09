package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/packager"
	webhook "github.com/zeitlos/lucity/services/webhook"
	webhookhttp "github.com/zeitlos/lucity/services/webhook/http"
)

type Config struct {
	Port          string `envconfig:"PORT" default:"9004"`
	LogLevel      string `envconfig:"LOG_LEVEL" default:"info"`
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" default:"dev-secret"`

	// GitHub App configuration for installation tokens.
	GitHubAppID          int64  `envconfig:"GITHUB_APP_ID"`
	GitHubPrivateKeyPath string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`

	// gRPC service addresses.
	BuilderAddr  string `envconfig:"BUILDER_ADDR" default:"localhost:9001"`
	PackagerAddr string `envconfig:"PACKAGER_ADDR" default:"localhost:9002"`
	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	// Registry for image paths.
	RegistryPushURL string `envconfig:"REGISTRY_PUSH_URL" default:"localhost:5000"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	// Build the push event handler if GitHub App + gRPC are configured.
	var handler *webhookhttp.Handler
	if config.GitHubAppID != 0 && config.GitHubPrivateKeyPath != "" {
		app, err := ghpkg.NewApp(config.GitHubAppID, "", "", config.WebhookSecret, "", config.GitHubPrivateKeyPath)
		if err != nil {
			slog.Error("failed to create github app", "error", err)
			os.Exit(1)
		}

		builderConn := dialGRPC(config.BuilderAddr)
		packagerConn := dialGRPC(config.PackagerAddr)
		deployerConn := dialGRPC(config.DeployerAddr)

		handler = &webhookhttp.Handler{
			GitHubApp: app,
			Pipeline: &webhook.Pipeline{
				Builder:         builder.NewBuilderServiceClient(builderConn),
				Packager:        packager.NewPackagerServiceClient(packagerConn),
				Deployer:        deployer.NewDeployerServiceClient(deployerConn),
				RegistryPushURL: config.RegistryPushURL,
			},
		}

		slog.Info("webhook CI/CD pipeline enabled",
			"builder", config.BuilderAddr,
			"packager", config.PackagerAddr,
			"deployer", config.DeployerAddr,
		)
	} else {
		slog.Info("webhook CI/CD pipeline disabled (missing GITHUB_APP_ID or GITHUB_PRIVATE_KEY_PATH)")
	}

	ctx, cancel := graceful.Context()
	defer cancel()

	httpServer := webhookhttp.NewServer(config.Port, config.WebhookSecret, handler)

	graceful.Serve(ctx, httpServer)
}

func dialGRPC(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to gRPC", "addr", addr, "error", err)
		os.Exit(1)
	}
	return conn
}
