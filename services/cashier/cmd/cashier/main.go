package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/metrics"
	"github.com/zeitlos/lucity/services/cashier/metering"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Config struct {
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	MetricsPort string `envconfig:"METRICS_PORT" default:"9009"`

	StripeSecretKey string `envconfig:"STRIPE_SECRET_KEY" required:"true"`
	HobbyPriceID    string `envconfig:"STRIPE_HOBBY_PRICE_ID" required:"true"`
	ProPriceID      string `envconfig:"STRIPE_PRO_PRICE_ID" required:"true"`

	EcoCPUMeterEvent  string `envconfig:"STRIPE_ECO_CPU_METER_EVENT"`
	EcoMemMeterEvent  string `envconfig:"STRIPE_ECO_MEM_METER_EVENT"`
	EcoDiskMeterEvent string `envconfig:"STRIPE_ECO_DISK_METER_EVENT"`
	StorageMeterEvent string `envconfig:"STRIPE_STORAGE_METER_EVENT"`
	EgressMeterEvent  string `envconfig:"STRIPE_EGRESS_METER_EVENT"`

	MeteringInterval   time.Duration `envconfig:"METERING_INTERVAL" default:"1h"`
	VictoriaMetricsURL string        `envconfig:"VICTORIA_METRICS_URL" required:"true"`
	SystemNamespace    string        `envconfig:"SYSTEM_NAMESPACE" default:"lucity-system"`

	// Logto Management API (M2M)
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

	stripeClient := stripelib.NewClient(config.StripeSecretKey, stripelib.PriceConfig{
		HobbyPriceID: config.HobbyPriceID,
		ProPriceID:   config.ProPriceID,
	}, stripelib.MeterConfig{
		EcoCPUEventName:  config.EcoCPUMeterEvent,
		EcoMemEventName:  config.EcoMemMeterEvent,
		EcoDiskEventName: config.EcoDiskMeterEvent,
		StorageEventName: config.StorageMeterEvent,
		EgressEventName:  config.EgressMeterEvent,
	})
	billingClient := billing.New(config.StripeSecretKey, config.HobbyPriceID, config.ProPriceID)

	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	vmClient, err := metering.NewVMClient(config.VictoriaMetricsURL)
	if err != nil {
		slog.Error("failed to connect to VictoriaMetrics", "error", err)
		os.Exit(1)
	}

	k8sClient, err := buildK8sClient()
	if err != nil {
		slog.Error("failed to create the Kubernetes client, metering needs it to read namespaces", "error", err)
		os.Exit(1)
	}

	registry := prometheus.NewRegistry()

	worker := metering.NewWorker(stripeClient, billingClient, logtoClient, vmClient, k8sClient, metrics.NewLoops(registry), config.SystemNamespace, config.MeteringInterval)
	slog.Info("metering enabled", "interval", config.MeteringInterval)

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, worker, metrics.NewServer(config.MetricsPort, registry))
}

// buildK8sClient creates a Kubernetes client, trying in-cluster config first
// then falling back to KUBECONFIG.
func buildK8sClient() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			return nil, errors.New("not running in a cluster and KUBECONFIG is not set")
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	return kubernetes.NewForConfig(cfg)
}
