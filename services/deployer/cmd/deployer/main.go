package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

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
	SoftServeHTTP        string `envconfig:"SOFTSERVE_HTTP_ADDR" default:"http://lucity-infra-soft-serve.lucity-system.svc.cluster.local:23232"`
	SoftServeClusterHTTP string `envconfig:"SOFTSERVE_CLUSTER_HTTP_ADDR"`
	SoftServeToken       string `envconfig:"SOFTSERVE_TOKEN"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	argoClient := argocd.NewClient(config.ArgocdAddr, config.ArgocdToken, config.ArgocdInsecure)

	clusterHTTP := config.SoftServeClusterHTTP
	if clusterHTTP == "" {
		clusterHTTP = config.SoftServeHTTP
	}

	// Build K8s config using standard loading rules.
	// Handles KUBECONFIG with multiple paths, in-cluster config, and ~/.kube/config fallback.
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	k8sConfig, k8sErr := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{},
	).ClientConfig()
	if k8sErr != nil {
		slog.Error("failed to create k8s config", "error", k8sErr)
		os.Exit(1)
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		slog.Error("failed to create k8s client", "error", err)
		os.Exit(1)
	}

	svc := deployergrpc.NewServer(argoClient, clusterHTTP, config.SoftServeToken, k8sClient)
	grpcServer := deployergrpc.NewGRPCServer(":"+config.Port, svc)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, grpcServer)
}
