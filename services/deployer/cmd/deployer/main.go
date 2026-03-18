package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/pkg/auth"
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
	GatewayName          string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace     string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	ClusterIssuer        string `envconfig:"CLUSTER_ISSUER" default:"letsencrypt-http01"`
	RegistryPullSecret   string `envconfig:"REGISTRY_PULL_SECRET" default:"lucity-registry-pull"`

	// Internal JWT (ES256 for gRPC service-to-service auth)
	InternalJWTPublicKeyPath string `envconfig:"INTERNAL_JWT_PUBLIC_KEY_PATH" required:"true"`
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

	dynClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		slog.Error("failed to create dynamic k8s client", "error", err)
		os.Exit(1)
	}

	svc := deployergrpc.NewServer(argoClient, clusterHTTP, config.SoftServeToken, k8sClient, dynClient, config.GatewayName, config.GatewayNamespace, config.ClusterIssuer, config.RegistryPullSecret)

	verifier, err := auth.NewInternalVerifierFromFile(config.InternalJWTPublicKeyPath)
	if err != nil {
		slog.Error("failed to create internal JWT verifier", "error", err)
		os.Exit(1)
	}

	grpcServer := deployergrpc.NewGRPCServer(":"+config.Port, svc, verifier)

	ctx, cancel := graceful.Context()
	defer cancel()

	go func() {
		if err := svc.ReconcileCustomDomains(ctx); err != nil {
			slog.Warn("initial custom domain reconciliation failed", "error", err)
		}
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := svc.ReconcileCustomDomains(ctx); err != nil {
					slog.Warn("custom domain reconciliation failed", "error", err)
				}
			}
		}
	}()

	graceful.Serve(ctx, grpcServer)
}
