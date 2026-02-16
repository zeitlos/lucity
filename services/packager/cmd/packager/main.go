package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	packagergrpc "github.com/zeitlos/lucity/services/packager/grpc"
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

	svc := packagergrpc.NewServer()
	grpcServer := packagergrpc.NewGRPCServer(":"+config.Port, svc)

	graceful.Serve(ctx, grpcServer)
}
