package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/metrics"
	"github.com/zeitlos/lucity/services/exporter/edge"
	"github.com/zeitlos/lucity/services/exporter/ovh"
	"github.com/zeitlos/lucity/services/exporter/storage"
)

type Config struct {
	Port     string        `envconfig:"PORT" default:"9007"`
	LogLevel string        `envconfig:"LOG_LEVEL" default:"info"`
	Interval time.Duration `envconfig:"INTERVAL" default:"5m"`

	OVHEndpoint          string `envconfig:"OVH_ENDPOINT" default:"ovh-eu"`
	OVHApplicationKey    string `envconfig:"OVH_APPLICATION_KEY" required:"true"`
	OVHApplicationSecret string `envconfig:"OVH_APPLICATION_SECRET" required:"true"`
	OVHConsumerKey       string `envconfig:"OVH_CONSUMER_KEY" required:"true"`
	OVHProjectID         string `envconfig:"OVH_PROJECT_ID" required:"true"`
	OVHRegion            string `envconfig:"OVH_REGION" default:"GRA"`

	// The edge zone count is optional. Without a key it is not collected.
	BunnyAPIKey string `envconfig:"BUNNY_API_KEY"`
}

func main() {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	client, err := ovh.New(config.OVHEndpoint, config.OVHApplicationKey, config.OVHApplicationSecret, config.OVHConsumerKey, config.OVHProjectID)

	if err != nil {
		slog.Error("failed to create ovh client", "error", err)
		os.Exit(1)
	}

	registry := prometheus.NewRegistry()
	loops := metrics.NewLoops(registry)

	collector := storage.New(client, loops, config.OVHRegion, config.Interval)
	registry.MustRegister(collector)

	servers := []graceful.Server{collector, metrics.NewServer(config.Port, registry)}

	if config.BunnyAPIKey != "" {
		zones := edge.New(bunny.New(config.BunnyAPIKey), loops, config.Interval)
		registry.MustRegister(zones)
		servers = append(servers, zones)
	} else {
		slog.Info("edge not configured, edge zones are not collected")
	}

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, servers...)
}
