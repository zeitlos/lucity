package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/builder/engine"
	buildergrpc "github.com/zeitlos/lucity/services/builder/grpc"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"9001"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	JWTSecret   string `envconfig:"JWT_SECRET" required:"true"`
	RegistryURL string `envconfig:"REGISTRY_URL" default:"ghcr.io"`
	WorkDir     string `envconfig:"WORK_DIR" default:"/tmp/lucity-builds"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	// Ensure work directory exists
	if err := os.MkdirAll(config.WorkDir, 0o755); err != nil {
		slog.Error("failed to create work dir", "error", err, "path", config.WorkDir)
		os.Exit(1)
	}

	ctx, cancel := graceful.Context()
	defer cancel()

	eng := engine.NewLocalEngine()
	svc := buildergrpc.NewServer(eng, config.RegistryURL, config.WorkDir)
	grpcServer := buildergrpc.NewGRPCServer(":"+config.Port, config.JWTSecret, svc)

	graceful.Serve(ctx, grpcServer)
}
