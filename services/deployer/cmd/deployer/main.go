package main

import (
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	Kubeconfig           string `envconfig:"KUBECONFIG"`
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

	// Build K8s config: in-cluster when running in K8s, kubeconfig for local dev.
	var k8sConfig *rest.Config
	var k8sErr error
	if config.Kubeconfig != "" {
		k8sConfig, k8sErr = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	} else {
		k8sConfig, k8sErr = rest.InClusterConfig()
		if k8sErr != nil {
			k8sConfig, k8sErr = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		}
	}
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
