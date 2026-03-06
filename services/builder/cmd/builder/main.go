package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/builder/build"
	"github.com/zeitlos/lucity/services/builder/engine"
	buildergrpc "github.com/zeitlos/lucity/services/builder/grpc"
)

type Config struct {
	Port             string `envconfig:"PORT" default:"9001"`
	LogLevel         string `envconfig:"LOG_LEVEL" default:"info"`
	JWTSecret        string `envconfig:"JWT_SECRET" required:"true"`
	RegistryURL      string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryToken    string `envconfig:"REGISTRY_TOKEN"`
	RegistryInsecure bool   `envconfig:"REGISTRY_INSECURE" default:"true"`
	WorkDir          string `envconfig:"WORK_DIR" default:"/tmp/lucity-builds"`
	BuildEngine      string `envconfig:"BUILD_ENGINE" default:"local"`
	BuildImage     string `envconfig:"BUILD_IMAGE"`
	BuildNamespace string `envconfig:"BUILD_NAMESPACE" default:"lucity-system"`
}

func main() {
	// Check for run-build subcommand (used inside K8s Job pods)
	if len(os.Args) > 1 && os.Args[1] == "run-build" {
		logger.Setup("info")
		runBuild()
		return
	}

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

	eng, tracker, err := setupEngine(config)
	if err != nil {
		slog.Error("failed to setup engine", "error", err)
		os.Exit(1)
	}

	svc := buildergrpc.NewServer(eng, tracker, config.RegistryURL, config.RegistryToken, config.RegistryInsecure, config.WorkDir)
	grpcServer := buildergrpc.NewGRPCServer(":"+config.Port, config.JWTSecret, svc)

	graceful.Serve(ctx, grpcServer)
}

func setupEngine(config Config) (engine.Engine, build.Tracker, error) {
	switch config.BuildEngine {
	case "local":
		slog.Info("using local build engine")
		return engine.NewLocalEngine(), build.NewInMemoryTracker(), nil

	case "kubernetes":
		slog.Info("using kubernetes build engine")

		if config.BuildImage == "" {
			return nil, nil, fmt.Errorf("BUILD_IMAGE is required for kubernetes engine")
		}

		k8sClient, err := kubernetesClient()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create k8s client: %w", err)
		}

		eng := engine.NewKubernetesEngine(engine.KubernetesEngineOpts{
			Client:      k8sClient,
			Namespace:   config.BuildNamespace,
			BuildImage:  config.BuildImage,
			RegistryURL: config.RegistryURL,
			Insecure:    config.RegistryInsecure,
		})

		tracker := build.NewK8sTracker(k8sClient, config.BuildNamespace)

		return eng, tracker, nil

	default:
		return nil, nil, fmt.Errorf("unknown build engine: %s", config.BuildEngine)
	}
}

func kubernetesClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(config)
}
