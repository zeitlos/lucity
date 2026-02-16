package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"9002"`
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

	slog.Info("packager starting", "port", config.Port)

	// TODO: initialize gRPC server with PackagerService implementation
	// TODO: initialize Soft-serve client for GitOps repo management

	_ = ctx
}
