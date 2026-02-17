package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/crypto/ssh"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/packager/gitops"
	packagergrpc "github.com/zeitlos/lucity/services/packager/grpc"
)

type Config struct {
	Port     string `envconfig:"PORT" default:"9002"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// Auth
	JWTSecret string `envconfig:"JWT_SECRET" required:"true"`

	// Git provider: "softserve" or "github"
	GitProvider string `envconfig:"GIT_PROVIDER" default:"softserve"`

	// Soft-serve config (when GIT_PROVIDER=softserve)
	SoftServeSSH     string `envconfig:"SOFTSERVE_SSH_ADDR" default:"localhost:23231"`
	SoftServeHTTP    string `envconfig:"SOFTSERVE_HTTP_ADDR" default:"http://localhost:23232"`
	SoftServeKeyPath string `envconfig:"SOFTSERVE_SSH_KEY_PATH"`
	SoftServeToken   string `envconfig:"SOFTSERVE_TOKEN"`
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

	var svc *packagergrpc.Server

	switch config.GitProvider {
	case "softserve":
		provider, err := buildSoftServeProvider(config)
		if err != nil {
			slog.Error("failed to create softserve provider", "error", err)
			os.Exit(1)
		}
		svc = packagergrpc.NewServerWithProvider(provider)
	case "github":
		svc = packagergrpc.NewServer()
	default:
		slog.Error("unknown git provider", "provider", config.GitProvider)
		os.Exit(1)
	}

	grpcServer := packagergrpc.NewGRPCServer(":"+config.Port, config.JWTSecret, svc)
	graceful.Serve(ctx, grpcServer)
}

func buildSoftServeProvider(config Config) (*gitops.SoftServeProvider, error) {
	if config.SoftServeKeyPath == "" {
		slog.Error("SOFTSERVE_SSH_KEY_PATH is required when GIT_PROVIDER=softserve")
		os.Exit(1)
	}

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
