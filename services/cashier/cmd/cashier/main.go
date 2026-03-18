package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	cashiergrpc "github.com/zeitlos/lucity/services/cashier/grpc"
	cashierhttp "github.com/zeitlos/lucity/services/cashier/http"
	"github.com/zeitlos/lucity/services/cashier/metering"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"9005"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9006"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	DeployerAddr string `envconfig:"DEPLOYER_ADDR" default:"localhost:9003"`

	StripeSecretKey     string `envconfig:"STRIPE_SECRET_KEY" required:"true"`
	StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET" required:"true"`

	HobbyPriceID    string `envconfig:"STRIPE_HOBBY_PRICE_ID" required:"true"`
	ProPriceID      string `envconfig:"STRIPE_PRO_PRICE_ID" required:"true"`
	EcoCPUPriceID   string `envconfig:"STRIPE_ECO_CPU_PRICE_ID" required:"true"`
	EcoMemPriceID   string `envconfig:"STRIPE_ECO_MEM_PRICE_ID" required:"true"`
	EcoDiskPriceID  string `envconfig:"STRIPE_ECO_DISK_PRICE_ID" required:"true"`
	ProdCPUPriceID  string `envconfig:"STRIPE_PROD_CPU_PRICE_ID" required:"true"`
	ProdMemPriceID  string `envconfig:"STRIPE_PROD_MEM_PRICE_ID" required:"true"`
	ProdDiskPriceID string `envconfig:"STRIPE_PROD_DISK_PRICE_ID" required:"true"`

	EcoCPUMeterEvent  string `envconfig:"STRIPE_ECO_CPU_METER_EVENT"`
	EcoMemMeterEvent  string `envconfig:"STRIPE_ECO_MEM_METER_EVENT"`
	EcoDiskMeterEvent string `envconfig:"STRIPE_ECO_DISK_METER_EVENT"`
	ProdCPUMeterEvent  string `envconfig:"STRIPE_PROD_CPU_METER_EVENT"`
	ProdMemMeterEvent  string `envconfig:"STRIPE_PROD_MEM_METER_EVENT"`
	ProdDiskMeterEvent string `envconfig:"STRIPE_PROD_DISK_METER_EVENT"`

	MeteringInterval    time.Duration `envconfig:"METERING_INTERVAL" default:"1h"`
	SignozClickhouseDSN string        `envconfig:"SIGNOZ_CLICKHOUSE_DSN"`

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

	// Connect to deployer for workspace metadata
	deployerConn, err := grpc.NewClient(config.DeployerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to deployer", "error", err, "addr", config.DeployerAddr)
		os.Exit(1)
	}
	defer deployerConn.Close()

	deployerClient := deployer.NewDeployerServiceClient(deployerConn)

	// Stripe client
	prices := stripelib.PriceConfig{
		HobbyPriceID:    config.HobbyPriceID,
		ProPriceID:      config.ProPriceID,
		EcoCPUPriceID:   config.EcoCPUPriceID,
		EcoMemPriceID:   config.EcoMemPriceID,
		EcoDiskPriceID:  config.EcoDiskPriceID,
		ProdCPUPriceID:  config.ProdCPUPriceID,
		ProdMemPriceID:  config.ProdMemPriceID,
		ProdDiskPriceID: config.ProdDiskPriceID,
	}
	meters := stripelib.MeterConfig{
		EcoCPUEventName:   config.EcoCPUMeterEvent,
		EcoMemEventName:   config.EcoMemMeterEvent,
		EcoDiskEventName:  config.EcoDiskMeterEvent,
		ProdCPUEventName:  config.ProdCPUMeterEvent,
		ProdMemEventName:  config.ProdMemMeterEvent,
		ProdDiskEventName: config.ProdDiskMeterEvent,
	}
	stripeClient := stripelib.NewClient(config.StripeSecretKey, prices, meters)

	// gRPC server
	svc := cashiergrpc.NewServer(stripeClient, deployerClient)

	verifier, err := auth.NewInternalVerifierFromFile(config.InternalJWTPublicKeyPath)
	if err != nil {
		slog.Error("failed to create internal JWT verifier", "error", err)
		os.Exit(1)
	}

	grpcServer := cashiergrpc.NewGRPCServer(":"+config.Port, svc, verifier)

	// Stripe webhook HTTP server
	webhookHandler := stripelib.NewWebhookHandler(config.StripeWebhookSecret, stripeClient, svc)
	httpServer := cashierhttp.NewServer(config.WebhookPort, webhookHandler)

	servers := []graceful.Server{grpcServer, httpServer}

	// Metering worker (optional — requires SigNoz ClickHouse)
	if config.SignozClickhouseDSN != "" {
		signozClient, err := metering.NewSigNozClient(config.SignozClickhouseDSN)
		if err != nil {
			slog.Error("failed to connect to SigNoz ClickHouse", "error", err)
			os.Exit(1)
		}
		defer signozClient.Close()

		// K8s client for metering checkpoint persistence. Optional — worker runs
		// without checkpoint/backfill if unavailable (e.g. local dev without cluster).
		k8sClient := buildK8sClient()

		worker := metering.NewWorker(stripeClient, deployerClient, signozClient, k8sClient, config.MeteringInterval)
		servers = append(servers, worker)
		slog.Info("metering enabled", "interval", config.MeteringInterval)
	} else {
		slog.Info("metering disabled — SIGNOZ_CLICKHOUSE_DSN not set")
	}

	ctx, cancel := graceful.Context()
	defer cancel()

	graceful.Serve(ctx, servers...)
}

// buildK8sClient creates a Kubernetes client, trying in-cluster config first
// then falling back to KUBECONFIG. Returns nil if neither is available.
func buildK8sClient() kubernetes.Interface {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig (local dev).
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			slog.Warn("metering: no K8s config available, checkpoint/backfill disabled")
			return nil
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			slog.Warn("metering: failed to load kubeconfig, checkpoint/backfill disabled", "error", err)
			return nil
		}
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		slog.Warn("metering: failed to create K8s client, checkpoint/backfill disabled", "error", err)
		return nil
	}
	return client
}
