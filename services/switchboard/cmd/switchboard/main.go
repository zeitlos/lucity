package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/zeitlos/lucity/pkg/agentrunner"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/switchboard/internal/bindings"
	"github.com/zeitlos/lucity/services/switchboard/internal/bridge"
	"github.com/zeitlos/lucity/services/switchboard/internal/people"
)

type Config struct {
	LogLevel          string `envconfig:"LOG_LEVEL" default:"info"`
	Namespace         string `envconfig:"NAMESPACE" required:"true"`
	DashboardURL      string `envconfig:"DASHBOARD_URL" required:"true"`
	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`
}

func main() {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	restConfig, err := rest.InClusterConfig()

	if err != nil {
		slog.Error("failed to load in-cluster config", "error", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)

	if err != nil {
		slog.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	runner, err := agentrunner.New(restConfig, clientset)

	if err != nil {
		slog.Error("failed to create agent runner client", "error", err)
		os.Exit(1)
	}

	directory := people.New(logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret))
	watcher := bindings.New(clientset, config.Namespace)
	server := bridge.New(runner, directory, watcher, strings.TrimSuffix(config.DashboardURL, "/"))

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, server)
}
