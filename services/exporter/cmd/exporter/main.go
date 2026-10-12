package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	exporterhttp "github.com/zeitlos/lucity/services/exporter/http"
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

	collector := storage.New(client, config.OVHRegion, config.Interval)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, collector, exporterhttp.NewServer(config.Port, registry))
}
