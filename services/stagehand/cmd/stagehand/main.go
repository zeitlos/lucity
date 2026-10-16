package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/blang/semver/v4"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/charts"
	"github.com/zeitlos/lucity/pkg/billing"
	helmDeployer "github.com/zeitlos/lucity/pkg/deployer/helm"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/services/stagehand/placement"
)

// Version is set at build time via ldflags. It must equal the conductor's
// version, because it becomes the chart version of every release stagehand
// applies.
var Version = "0.0.0-dev"

type Config struct {
	LogLevel string        `envconfig:"LOG_LEVEL" default:"info"`
	Interval time.Duration `envconfig:"INTERVAL" default:"2m"`
	MaxMoves int           `envconfig:"MAX_MOVES" default:"3"`
	DryRun   bool          `envconfig:"DRY_RUN" default:"false"`

	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`

	// Billing is optional. Without a key every environment gets the default
	// placement.
	StripeSecretKey    string `envconfig:"STRIPE_SECRET_KEY"`
	StripeHobbyPriceID string `envconfig:"STRIPE_HOBBY_PRICE_ID"`
	StripeProPriceID   string `envconfig:"STRIPE_PRO_PRICE_ID"`

	WorkloadDefaultPlacement values.Placement `envconfig:"WORKLOAD_DEFAULT_PLACEMENT"`
	WorkloadPaidPlacement    values.Placement `envconfig:"WORKLOAD_PAID_PLACEMENT"`

	// Every apply re-stamps these values, so they must match the conductor's.
	RegistryPullURL        string `envconfig:"REGISTRY_PULL_URL" required:"true"`
	GatewayName            string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace       string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	GatewayHTTPSListener   string `envconfig:"GATEWAY_HTTPS_LISTENER"`
	ClusterIssuer          string `envconfig:"CUSTOM_DOMAIN_CLUSTER_ISSUER" default:"letsencrypt-http01"`
	DatabaseBackupEnabled  bool   `envconfig:"DATABASE_BACKUP_ENABLED" default:"false"`
	DatabaseBackupEndpoint string `envconfig:"DATABASE_BACKUP_S3_ENDPOINT"`
	DatabaseBackupBucket   string `envconfig:"DATABASE_BACKUP_S3_BUCKET"`
	EdgeHeaderEnforced     bool   `envconfig:"EDGE_HEADER_ENFORCED" default:"false"`
}

func main() {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	version, err := semver.Parse(Version)

	if err != nil {
		slog.Error("failed to parse version", "error", err, "version", Version)
		os.Exit(1)
	}

	k8sClient, metadataClient, err := buildKubeClients()

	if err != nil {
		slog.Error("failed to build kube clients", "error", err)
		os.Exit(1)
	}

	chartRef, err := helmDeployer.LoadChartFromFS(charts.LucityApp, "lucity-app")

	if err != nil {
		slog.Error("failed to load lucity-app chart", "error", err)
		os.Exit(1)
	}

	chartRef.Metadata.Version = version.String()

	ctx, cancel := graceful.Context()
	defer cancel()

	releaseCache := helmDeployer.NewReleaseCache(metadataClient)
	releaseCache.Start(ctx)

	deployerClient, err := helmDeployer.New(chartRef, config.GatewayName, config.GatewayNamespace, config.ClusterIssuer, helmDeployer.BackupConfig{
		Enabled:  config.DatabaseBackupEnabled,
		Endpoint: config.DatabaseBackupEndpoint,
		Bucket:   config.DatabaseBackupBucket,
	}, config.RegistryPullURL, config.WorkloadDefaultPlacement, helmDeployer.WithEdgeHeader(config.EdgeHeaderEnforced), helmDeployer.WithHTTPSListener(config.GatewayHTTPSListener), helmDeployer.WithReleaseCache(releaseCache))

	if err != nil {
		slog.Error("failed to create deployer client", "error", err)
		os.Exit(1)
	}

	var billingClient *billing.Client

	if config.StripeSecretKey != "" {
		if config.StripeHobbyPriceID == "" || config.StripeProPriceID == "" {
			slog.Error("STRIPE_SECRET_KEY requires STRIPE_HOBBY_PRICE_ID and STRIPE_PRO_PRICE_ID")
			os.Exit(1)
		}

		billingClient = billing.New(config.StripeSecretKey, config.StripeHobbyPriceID, config.StripeProPriceID)
	} else {
		slog.Info("billing not configured, every environment gets the default placement")
	}

	directory := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)

	worker := placement.New(k8sClient, deployerClient.Environments(), directory, billingClient, config.WorkloadDefaultPlacement, config.WorkloadPaidPlacement, config.Interval, config.MaxMoves, config.DryRun)

	slog.Info("stagehand started", "version", Version, "interval", config.Interval, "max_moves", config.MaxMoves, "dry_run", config.DryRun)

	worker.Run(ctx)
}

func buildKubeClients() (kubernetes.Interface, metadata.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()

	if err != nil {
		return nil, nil, err
	}

	k8s, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, nil, err
	}

	meta, err := metadata.NewForConfig(cfg)

	if err != nil {
		return nil, nil, err
	}

	return k8s, meta, nil
}
