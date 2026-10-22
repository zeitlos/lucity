package main

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blang/semver/v4"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/charts"
	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/bunny"
	helmDeployer "github.com/zeitlos/lucity/pkg/deployer/helm"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/domains"
	"github.com/zeitlos/lucity/pkg/edge"
	environmentK8s "github.com/zeitlos/lucity/pkg/environment/kubernetes"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/hostname"
	"github.com/zeitlos/lucity/pkg/jobs"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/pipeline"
	platformK8s "github.com/zeitlos/lucity/pkg/platform/kubernetes"
	"github.com/zeitlos/lucity/services/stagehand/admission"
	"github.com/zeitlos/lucity/services/stagehand/backups"
	"github.com/zeitlos/lucity/services/stagehand/customdomains"
	"github.com/zeitlos/lucity/services/stagehand/releases"
	"github.com/zeitlos/lucity/services/stagehand/workspaces"
)

// Version is set at build time via ldflags. It must equal the conductor's
// version, because it becomes the chart version of every release stagehand
// applies.
var Version = "0.0.0-dev"

const (
	workspaceMaxAge       = time.Minute
	pendingDomainInterval = 15 * time.Second
	domainSyncInterval    = 2 * time.Minute
	backupInterval        = 10 * time.Minute
)

type Config struct {
	LogLevel string        `envconfig:"LOG_LEVEL" default:"info"`
	Interval time.Duration `envconfig:"INTERVAL" default:"2m"`
	MaxMoves int           `envconfig:"MAX_MOVES" default:"3"`
	DryRun   bool          `envconfig:"DRY_RUN" default:"false"`

	SystemNamespace string `envconfig:"SYSTEM_NAMESPACE" default:"lucity-system"`
	BuildNamespace  string `envconfig:"BUILD_NAMESPACE" default:"lucity-builds"`

	MaxConcurrentReleases int `envconfig:"MAX_CONCURRENT_RELEASES" default:"5"`
	MaxConcurrentBuilds   int `envconfig:"MAX_CONCURRENT_BUILDS" default:"2"`

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

	// Namespace scaffolding, the same settings as the conductor's.
	RegistryPullSecret            string `envconfig:"REGISTRY_PULL_SECRET" default:"lucity-registry-pull"`
	PodCIDR                       string `envconfig:"POD_CIDR" default:"10.244.0.0/16"`
	ServiceCIDR                   string `envconfig:"SERVICE_CIDR" default:"10.96.0.0/12"`
	DatabaseBackupRegion          string `envconfig:"DATABASE_BACKUP_S3_REGION"`
	DatabaseBackupAccessKeyID     string `envconfig:"DATABASE_BACKUP_S3_ACCESS_KEY_ID"`
	DatabaseBackupSecretAccessKey string `envconfig:"DATABASE_BACKUP_S3_SECRET_ACCESS_KEY"`

	// Custom domains and the edge CDN, the same settings as the conductor's.
	WorkloadDomain        string `envconfig:"WORKLOAD_DOMAIN" required:"true"`
	IPAddress             string `envconfig:"IP_ADDRESS"`
	BunnyAPIKey           string `envconfig:"BUNNY_API_KEY"`
	EdgeOriginURL         string `envconfig:"EDGE_ORIGIN_URL"`
	EdgeHeaderSecret      string `envconfig:"EDGE_HEADER_SECRET"`
	EdgeWildcardTLSSecret string `envconfig:"EDGE_WILDCARD_TLS_SECRET" default:"lucity-app-wildcard-tls"`

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

	k8sClient, dynClient, metadataClient, err := buildKubeClients()

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

	platformClient := platformK8s.New(k8sClient, dynClient)
	platformClient.Start(ctx)

	environmentClient := environmentK8s.New(k8sClient, dynClient, config.SystemNamespace, config.RegistryPullSecret, config.PodCIDR, config.ServiceCIDR, environmentK8s.BackupCredentials{
		AccessKeyID:     config.DatabaseBackupAccessKeyID,
		SecretAccessKey: config.DatabaseBackupSecretAccessKey,
		Region:          config.DatabaseBackupRegion,
	})

	var edgeClient edge.Interface = edge.Disabled{}

	if config.BunnyAPIKey != "" && config.EdgeOriginURL != "" {
		edgeClient = edge.New(bunny.New(config.BunnyAPIKey), edge.Config{
			OriginURL:    config.EdgeOriginURL,
			HeaderSecret: config.EdgeHeaderSecret,
			ZonePrefix:   strings.ReplaceAll(config.WorkloadDomain, ".", "-"),
		})
	} else {
		slog.Info("edge not configured, custom domains are not synced to the edge")
	}

	hostnameClient := hostname.New(config.WorkloadDomain, "lb."+config.WorkloadDomain, config.IPAddress, edgeClient.EdgeAddresses, k8sClient, dynClient)
	domainsClient := domains.New(hostnameClient, edgeClient, platformClient, deployerClient, k8sClient, config.WorkloadDomain, config.GatewayNamespace, config.EdgeWildcardTLSSecret)

	jobsCache, err := jobs.NewCache(k8sClient, config.BuildNamespace, config.SystemNamespace)

	if err != nil {
		slog.Error("failed to create job cache", "error", err)
		os.Exit(1)
	}

	admissionWorker, err := admission.New(jobsCache, pipeline.New(k8sClient, jobsCache, config.MaxConcurrentReleases, config.MaxConcurrentBuilds))

	if err != nil {
		slog.Error("failed to create admission worker", "error", err)
		os.Exit(1)
	}

	jobsCache.Start(ctx)

	directory := workspaces.New(logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret), platformClient, workspaceMaxAge)

	releaseWorker := releases.New(k8sClient, environmentClient, deployerClient.Environments(), directory, billingClient, config.WorkloadDefaultPlacement, config.WorkloadPaidPlacement, config.Interval, config.MaxMoves, config.DryRun)
	domainWorker := customdomains.New(domainsClient, directory, pendingDomainInterval, domainSyncInterval)

	slog.Info("stagehand started", "version", Version, "interval", config.Interval, "max_moves", config.MaxMoves, "dry_run", config.DryRun, "max_concurrent_releases", config.MaxConcurrentReleases, "max_concurrent_builds", config.MaxConcurrentBuilds)

	var workers sync.WaitGroup

	workers.Go(func() { admissionWorker.Run(ctx) })
	workers.Go(func() { releaseWorker.Run(ctx) })
	workers.Go(func() { domainWorker.Run(ctx) })

	if config.DatabaseBackupEnabled {
		backupWorker := backups.New(platformClient, directory, backupInterval)
		workers.Go(func() { backupWorker.Run(ctx) })
	}

	workers.Wait()
}

func buildKubeClients() (kubernetes.Interface, dynamic.Interface, metadata.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()

	if err != nil {
		return nil, nil, nil, err
	}

	cfg.QPS = 50
	cfg.Burst = 100

	k8s, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		return nil, nil, nil, err
	}

	dyn, err := dynamic.NewForConfig(cfg)

	if err != nil {
		return nil, nil, nil, err
	}

	meta, err := metadata.NewForConfig(cfg)

	if err != nil {
		return nil, nil, nil, err
	}

	return k8s, dyn, meta, nil
}
