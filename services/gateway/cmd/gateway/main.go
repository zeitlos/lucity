package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/gateway/handler"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
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

	api := handler.New()
	graphqlServer := NewGraphQLServer(config.Port, api)

	graceful.Serve(ctx, graphqlServer)
}
