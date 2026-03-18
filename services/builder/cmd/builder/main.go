package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/builder/build"
	"github.com/zeitlos/lucity/services/builder/engine"
	buildergrpc "github.com/zeitlos/lucity/services/builder/grpc"
)

type Config struct {
	Port               string `envconfig:"PORT" default:"9001"`
	LogLevel           string `envconfig:"LOG_LEVEL" default:"info"`
	RegistryURL        string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryUsername   string `envconfig:"REGISTRY_USERNAME"`
	RegistryPassword   string `envconfig:"REGISTRY_PASSWORD"`
	RegistryAuthSecret string `envconfig:"REGISTRY_AUTH_SECRET"`
	RegistryInsecure   bool   `envconfig:"REGISTRY_INSECURE" default:"true"`
	WorkDir            string `envconfig:"WORK_DIR" default:"/tmp/lucity-builds"`
	BuildEngine        string `envconfig:"BUILD_ENGINE" default:"local"`
	BuildImage         string `envconfig:"BUILD_IMAGE"`
	BuildkitAddr       string `envconfig:"BUILDKIT_ADDR"`
	BuildNamespace     string `envconfig:"BUILD_NAMESPACE" default:"lucity-system"`
	KubeContext        string `envconfig:"KUBE_CONTEXT"`
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

	svc := buildergrpc.NewServer(eng, tracker, config.RegistryURL, config.RegistryUsername, config.RegistryPassword, config.RegistryInsecure, config.WorkDir)
	grpcServer := buildergrpc.NewGRPCServer(":"+config.Port, svc)

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

		k8sClient, err := kubernetesClient(config.KubeContext)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create k8s client: %w", err)
		}

		eng := engine.NewKubernetesEngine(engine.KubernetesEngineOpts{
			Client:             k8sClient,
			Namespace:          config.BuildNamespace,
			BuildImage:         config.BuildImage,
			BuildkitAddr:       config.BuildkitAddr,
			RegistryURL:        config.RegistryURL,
			RegistryAuthSecret: config.RegistryAuthSecret,
			Insecure:           config.RegistryInsecure,
		})

		tracker := build.NewK8sTracker(k8sClient, config.BuildNamespace)

		return eng, tracker, nil

	default:
		return nil, nil, fmt.Errorf("unknown build engine: %s", config.BuildEngine)
	}
}

func kubernetesClient(kubeContext string) (kubernetes.Interface, error) {
	// Try in-cluster config first (running inside K8s)
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig (local dev)
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		overrides := &clientcmd.ConfigOverrides{}
		if kubeContext != "" {
			overrides.CurrentContext = kubeContext
		}
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}
	return kubernetes.NewForConfig(cfg)
}
