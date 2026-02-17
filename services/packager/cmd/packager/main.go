package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	gh "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	packagergrpc "github.com/zeitlos/lucity/services/packager/grpc"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"9002"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// GitHub App (for installation tokens to manage GitOps repos)
	GitHubAppID          int64  `envconfig:"GITHUB_APP_ID" required:"true"`
	GitHubPrivateKeyPath string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`

	// Auth
	JWTSecret string `envconfig:"JWT_SECRET" required:"true"`
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

	// The packager only needs appID + privateKey for installation tokens.
	// OAuth fields are not used here.
	githubApp, err := gh.NewApp(
		config.GitHubAppID,
		"", "", "", "",
		config.GitHubPrivateKeyPath,
	)
	if err != nil {
		slog.Error("failed to initialize github app", "error", err)
		os.Exit(1)
	}

	svc := packagergrpc.NewServer(githubApp)
	grpcServer := packagergrpc.NewGRPCServer(":"+config.Port, config.JWTSecret, svc)

	graceful.Serve(ctx, grpcServer)
}
