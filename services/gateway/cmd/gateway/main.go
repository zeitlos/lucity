package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	gh "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// GitHub App
	GitHubAppID        int64  `envconfig:"GITHUB_APP_ID" required:"true"`
	GitHubClientID     string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GitHubClientSecret string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
	GitHubWebhookSecret string `envconfig:"GITHUB_WEBHOOK_SECRET" default:"dev-secret"`

	// Auth
	JWTSecret    string `envconfig:"JWT_SECRET" required:"true"`
	DashboardURL string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`
	CallbackURL  string `envconfig:"CALLBACK_URL" default:"http://localhost:8080/auth/github/callback"`
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

	githubApp := gh.NewApp(
		config.GitHubAppID,
		config.GitHubClientID,
		config.GitHubClientSecret,
		config.GitHubWebhookSecret,
		config.CallbackURL,
	)

	api := handler.New()
	graphqlServer := NewGraphQLServer(config.Port, api, githubApp, config.JWTSecret, config.DashboardURL)

	graceful.Serve(ctx, graphqlServer)
}
