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
	"github.com/zeitlos/lucity/pkg/conductor"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	cashiergrpc "github.com/zeitlos/lucity/services/cashier/grpc"
	cashierhttp "github.com/zeitlos/lucity/services/cashier/http"
	"github.com/zeitlos/lucity/services/cashier/metering"
	stripelib "github.com/zeitlos/lucity/services/cashier/stripe"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"9005"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9006"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// Conductor is the unified control-plane binary that absorbed
	// the deployer, packager, builder, and webhook services.
	// Backward-compat: DEPLOYER_ADDR is honored as a fallback.
	ConductorAddr string `envconfig:"CONDUCTOR_ADDR"`
	DeployerAddr  string `envconfig:"DEPLOYER_ADDR" default:"localhost:9090"`

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

	EcoCPUMeterEvent   string `envconfig:"STRIPE_ECO_CPU_METER_EVENT"`
	EcoMemMeterEvent   string `envconfig:"STRIPE_ECO_MEM_METER_EVENT"`
	EcoDiskMeterEvent  string `envconfig:"STRIPE_ECO_DISK_METER_EVENT"`
	ProdCPUMeterEvent  string `envconfig:"STRIPE_PROD_CPU_METER_EVENT"`
	ProdMemMeterEvent  string `envconfig:"STRIPE_PROD_MEM_METER_EVENT"`
	ProdDiskMeterEvent string `envconfig:"STRIPE_PROD_DISK_METER_EVENT"`

	MeteringInterval   time.Duration `envconfig:"METERING_INTERVAL" default:"1h"`
	VictoriaMetricsURL string        `envconfig:"VICTORIA_METRICS_URL"`

	// Logto Management API (M2M)
	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`

	// Internal JWT (ES256 for gRPC service-to-service auth)
	InternalJWTPublicKeyPath  string `envconfig:"INTERNAL_JWT_PUBLIC_KEY_PATH" required:"true"`
	InternalJWTPrivateKeyPath string `envconfig:"INTERNAL_JWT_PRIVATE_KEY_PATH" required:"true"`
}

func main() {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	// Connect to conductor (formerly deployer) for SuspendWorkspace
	// and ListResourceAllocations. CONDUCTOR_ADDR wins over the legacy
	// DEPLOYER_ADDR when both are set.
	conductorAddr := config.ConductorAddr
	if conductorAddr == "" {
		conductorAddr = config.DeployerAddr
	}
	conductorConn, err := grpc.NewClient(conductorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to connect to conductor", "error", err, "addr", conductorAddr)
		os.Exit(1)
	}
	defer conductorConn.Close()

	conductorClient := conductor.NewConductorServiceClient(conductorConn)

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

	// Logto client for workspace metadata and suspension state
	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	// Internal JWT for service-to-service auth
	verifier, err := auth.NewInternalVerifierFromFile(config.InternalJWTPublicKeyPath)
	if err != nil {
		slog.Error("failed to create internal JWT verifier", "error", err)
		os.Exit(1)
	}

	issuer, err := auth.NewIssuerFromFile(config.InternalJWTPrivateKeyPath)
	if err != nil {
		slog.Error("failed to create internal JWT issuer", "error", err)
		os.Exit(1)
	}

	// gRPC server
	svc := cashiergrpc.NewServer(stripeClient, conductorClient, logtoClient, issuer)

	grpcServer := cashiergrpc.NewGRPCServer(":"+config.Port, svc, verifier)

	// Stripe webhook HTTP server
	webhookHandler := stripelib.NewWebhookHandler(config.StripeWebhookSecret, stripeClient, svc)
	httpServer := cashierhttp.NewServer(config.WebhookPort, webhookHandler)

	servers := []graceful.Server{grpcServer, httpServer}

	// Metering worker (optional — requires VictoriaMetrics)
	if config.VictoriaMetricsURL != "" {
		vmClient, err := metering.NewVMClient(config.VictoriaMetricsURL)
		if err != nil {
			slog.Error("failed to connect to VictoriaMetrics", "error", err)
			os.Exit(1)
		}

		// K8s client for metering checkpoint persistence. Optional — worker runs
		// without checkpoint/backfill if unavailable (e.g. local dev without cluster).
		k8sClient := buildK8sClient()

		worker := metering.NewWorker(stripeClient, conductorClient, logtoClient, vmClient, k8sClient, issuer, config.MeteringInterval)
		servers = append(servers, worker)
		slog.Info("metering enabled", "interval", config.MeteringInterval)
	} else {
		slog.Info("metering disabled: VICTORIA_METRICS_URL not set")
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
