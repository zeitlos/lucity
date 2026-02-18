package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/deployer/argocd"
	deployergrpc "github.com/zeitlos/lucity/services/deployer/grpc"
)

type Config struct {
	Port          string `envconfig:"PORT" default:"9003"`
	LogLevel      string `envconfig:"LOG_LEVEL" default:"info"`
	ArgocdAddr    string `envconfig:"ARGOCD_ADDR" required:"true"`
	ArgocdToken   string `envconfig:"ARGOCD_TOKEN" required:"true"`
	ArgocdInsecure bool  `envconfig:"ARGOCD_INSECURE" default:"false"`
	SoftServeHTTP string `envconfig:"SOFTSERVE_HTTP_ADDR" default:"http://lucity-infra-soft-serve.lucity-system.svc.cluster.local:23232"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	argoClient := argocd.NewClient(config.ArgocdAddr, config.ArgocdToken, config.ArgocdInsecure)

	svc := deployergrpc.NewServer(argoClient, config.SoftServeHTTP)
	grpcServer := deployergrpc.NewGRPCServer(":"+config.Port, svc)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, grpcServer)
}
