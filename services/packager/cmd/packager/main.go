package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/packager/gitops"
	packagergrpc "github.com/zeitlos/lucity/services/packager/grpc"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"9002"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Soft-serve config
	SoftServeSSH     string `envconfig:"SOFTSERVE_SSH_ADDR" default:"localhost:23231"`
	SoftServeHTTP    string `envconfig:"SOFTSERVE_HTTP_ADDR" default:"http://localhost:23232"`
	SoftServeKeyPath string `envconfig:"SOFTSERVE_SSH_KEY_PATH" required:"true"`
	SoftServeToken   string `envconfig:"SOFTSERVE_TOKEN"`

	// Backend services
	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`
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

	provider, err := buildSoftServeProvider(config)
	if err != nil {
		slog.Error("failed to create softserve provider", "error", err)
		os.Exit(1)
	}

	// Connect to deployer for triggering ArgoCD syncs after commits.
	deployerConn, err := grpc.NewClient(config.DeployerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to deployer", "error", err, "addr", config.DeployerAddr)
		os.Exit(1)
	}
	defer deployerConn.Close()

	deployerClient := deployer.NewDeployerServiceClient(deployerConn)

	svc := packagergrpc.NewServer(provider, deployerClient)
	grpcServer := packagergrpc.NewGRPCServer(":"+config.Port, svc)
	graceful.Serve(ctx, grpcServer)
}

func buildSoftServeProvider(config Config) (*gitops.SoftServeProvider, error) {
	keyData, err := os.ReadFile(config.SoftServeKeyPath)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, err
	}

	slog.Info("using soft-serve git provider",
		"ssh", config.SoftServeSSH,
		"http", config.SoftServeHTTP)

	return gitops.NewSoftServeProvider(
		config.SoftServeSSH,
		config.SoftServeHTTP,
		signer,
		config.SoftServeToken,
	), nil
}
