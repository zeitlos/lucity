package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeitlos/lucity/charts"
	helmDeployer "github.com/zeitlos/lucity/pkg/deployer/helm"
	"github.com/zeitlos/lucity/pkg/deployer/values"
	"github.com/zeitlos/lucity/pkg/domains"
	"github.com/zeitlos/lucity/pkg/edge"
	environmentK8s "github.com/zeitlos/lucity/pkg/environment/kubernetes"
	"github.com/zeitlos/lucity/pkg/hostname"
	"github.com/zeitlos/lucity/pkg/jobs"
	metricsHTTP "github.com/zeitlos/lucity/pkg/metrics"
	"github.com/zeitlos/lucity/pkg/pipeline"
	platformK8s "github.com/zeitlos/lucity/pkg/platform/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/agentaudit"
	"github.com/zeitlos/lucity/services/conductor/internal/agentrunner"
	webhookhttp "github.com/zeitlos/lucity/services/conductor/internal/api/webhook/http"
	"github.com/zeitlos/lucity/services/conductor/internal/backuparchive"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
	buildjobK8s "github.com/zeitlos/lucity/services/conductor/internal/buildjob/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	deployjobK8s "github.com/zeitlos/lucity/services/conductor/internal/deployjob/kubernetes"
	directoryLogto "github.com/zeitlos/lucity/services/conductor/internal/directory/logto"
	"github.com/zeitlos/lucity/services/conductor/internal/metrics"
	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
	objectstorageBunny "github.com/zeitlos/lucity/services/conductor/internal/objectstorage/bunny"
	objectstorageOVH "github.com/zeitlos/lucity/services/conductor/internal/objectstorage/ovh"
	"github.com/zeitlos/lucity/services/conductor/internal/planner/railpack"
	scanjobK8s "github.com/zeitlos/lucity/services/conductor/internal/scanjob/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/scanreport"
	sourceGH "github.com/zeitlos/lucity/services/conductor/internal/source/github"
	"github.com/zeitlos/lucity/services/conductor/internal/upstream"
	"github.com/zeitlos/lucity/services/conductor/internal/vulnerabilities"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/bunny"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/kvstore"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/oidc"
	"github.com/zeitlos/lucity/pkg/session"

	kauth "github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/kelseyhightower/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9004"`
	MetricsPort string `envconfig:"METRICS_PORT" default:"9008"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// OIDC
	OIDCIssuerURL    string `envconfig:"OIDC_ISSUER_URL" required:"true"`
	OIDCDiscoveryURL string `envconfig:"OIDC_DISCOVERY_URL"`
	OIDCClientID     string `envconfig:"OIDC_CLIENT_ID" required:"true"`
	OIDCCallbackURL  string `envconfig:"OIDC_CALLBACK_URL" default:"http://localhost:8080/auth/callback"`
	OIDCAudience     string `envconfig:"OIDC_AUDIENCE"`
	OIDCCLIClientID  string `envconfig:"OIDC_CLI_CLIENT_ID"`

	// Auth
	DashboardURL string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`
	// SessionSecret keys the sealed browser session cookie (pkg/session).
	SessionSecret string `envconfig:"SESSION_SECRET" required:"true"`

	// Logto Management API (M2M)
	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`

	// Stripe billing: optional, disabled without a key
	StripeSecretKey       string `envconfig:"STRIPE_SECRET_KEY"`
	StripeHobbyPriceID    string `envconfig:"STRIPE_HOBBY_PRICE_ID"`
	StripeProPriceID      string `envconfig:"STRIPE_PRO_PRICE_ID"`
	StripeEcoCPUPriceID   string `envconfig:"STRIPE_ECO_CPU_PRICE_ID"`
	StripeEcoMemPriceID   string `envconfig:"STRIPE_ECO_MEM_PRICE_ID"`
	StripeEcoDiskPriceID  string `envconfig:"STRIPE_ECO_DISK_PRICE_ID"`
	StripeProdCPUPriceID  string `envconfig:"STRIPE_PROD_CPU_PRICE_ID"`
	StripeProdMemPriceID  string `envconfig:"STRIPE_PROD_MEM_PRICE_ID"`
	StripeProdDiskPriceID string `envconfig:"STRIPE_PROD_DISK_PRICE_ID"`
	StripeStoragePriceID  string `envconfig:"STRIPE_STORAGE_PRICE_ID"`
	StripeEgressPriceID   string `envconfig:"STRIPE_EGRESS_PRICE_ID"`

	// Webhook (GitHub events)
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" required:"true"`

	// Cluster
	GatewayName          string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace     string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	GatewayHTTPSListener string `envconfig:"GATEWAY_HTTPS_LISTENER"`
	RegistryPullSecret   string `envconfig:"REGISTRY_PULL_SECRET" default:"lucity-registry-pull"`

	// Per-env NetworkPolicy needs the cluster's pod and service CIDRs to
	// carve out "internet but not the cluster" egress. These are
	// cluster-wide constants; defaults match kubeadm + K8s defaults.
	PodCIDR     string `envconfig:"POD_CIDR" default:"10.244.0.0/16"`
	ServiceCIDR string `envconfig:"SERVICE_CIDR" default:"10.96.0.0/12"`

	// Builder
	RegistryURL        string         `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryPullURL    string         `envconfig:"REGISTRY_PULL_URL" required:"true"`
	RegistryPushURL    string         `envconfig:"REGISTRY_PUSH_URL" required:"true"`
	RegistryAuthSecret string         `envconfig:"REGISTRY_AUTH_SECRET" required:"true"`
	BuildImage         string         `envconfig:"BUILD_IMAGE"`
	BuildkitAddr       string         `envconfig:"BUILDKIT_ADDR"`
	BuildNamespace     string         `envconfig:"BUILD_NAMESPACE" default:"lucity-builds"`
	BuildJobPlacement  jobs.Placement `envconfig:"BUILD_JOB_PLACEMENT"`
	BuildkitTLSSecret  string         `envconfig:"BUILDKIT_TLS_SECRET"`
	BuildkitServerName string         `envconfig:"BUILDKIT_SERVER_NAME"`

	// Deploy jobs
	DeployImage          string         `envconfig:"DEPLOY_IMAGE" required:"true"`
	DeployServiceAccount string         `envconfig:"DEPLOY_SERVICE_ACCOUNT"`
	DeployJobPlacement   jobs.Placement `envconfig:"DEPLOY_JOB_PLACEMENT"`

	// Tenant workload placement for releases installed by the conductor or its
	// deploy Jobs. Stagehand moves releases to their plan's placement afterwards.
	WorkloadDefaultPlacement string `envconfig:"WORKLOAD_DEFAULT_PLACEMENT"`
	WorkloadPaidPlacement    string `envconfig:"WORKLOAD_PAID_PLACEMENT"`

	// Secret scan tuning
	ScanTimeout               time.Duration `envconfig:"SCAN_TIMEOUT" default:"60m"`
	ScanGitleaksWorkers       int           `envconfig:"SCAN_GITLEAKS_WORKERS" default:"3"`
	ScanTrufflehogConcurrency int           `envconfig:"SCAN_TRUFFLEHOG_CONCURRENCY" default:"2"`

	MaxConcurrentReleases int `envconfig:"MAX_CONCURRENT_RELEASES" default:"5"`
	MaxConcurrentBuilds   int `envconfig:"MAX_CONCURRENT_BUILDS" default:"2"`
	MaxQueuedReleases     int `envconfig:"MAX_QUEUED_RELEASES" default:"10"`

	SystemNamespace string `envconfig:"SYSTEM_NAMESPACE" default:"lucity-system"`

	// Observability
	VictoriaMetricsURL string `envconfig:"VICTORIA_METRICS_URL" required:"true"`

	// Keyless CI deploys (GitHub Actions OIDC)
	GitHubActionsAudience string        `envconfig:"GITHUB_ACTIONS_AUDIENCE"`
	CISessionTTL          time.Duration `envconfig:"CI_SESSION_TTL" default:"1h"`

	// GitHub App (for installation tokens + OAuth)
	GitHubAppID          int64  `envconfig:"GITHUB_APP_ID" required:"true"`
	GitHubPrivateKeyPath string `envconfig:"GITHUB_PRIVATE_KEY_PATH" required:"true"`
	GitHubClientID       string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GitHubClientSecret   string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
	GitHubAppSlug        string `envconfig:"GITHUB_APP_SLUG" required:"true"`

	// Domains
	WorkloadDomain            string `envconfig:"WORKLOAD_DOMAIN" required:"true"`
	DatabaseDomain            string `envconfig:"DATABASE_DOMAIN" required:"true"`
	IPAddress                 string `envconfig:"IP_ADDRESS"`
	CustomDomainClusterIssuer string `envconfig:"CUSTOM_DOMAIN_CLUSTER_ISSUER" default:"letsencrypt-http01"`

	// OVH
	OVHEndpoint    string `envconfig:"OVH_ENDPOINT" default:"ovh-eu"`
	OVHAppKey      string `envconfig:"OVH_APPLICATION_KEY" required:"true"`
	OVHAppSecret   string `envconfig:"OVH_APPLICATION_SECRET" required:"true"`
	OVHConsumerKey string `envconfig:"OVH_CONSUMER_KEY" required:"true"`
	OVHProjectID   string `envconfig:"OVH_PROJECT_ID" required:"true"`
	OVHRegion      string `envconfig:"OVH_REGION" default:"GRA"`

	// Public buckets (CDN)
	BunnyAPIKey        string `envconfig:"BUNNY_API_KEY"`
	PublicBucketDomain string `envconfig:"PUBLIC_BUCKET_DOMAIN" default:"storage.lucity.app"`

	// Edge (CDN in front of the workload gateway)
	EdgeOriginURL         string `envconfig:"EDGE_ORIGIN_URL"`
	EdgeHeaderSecret      string `envconfig:"EDGE_HEADER_SECRET"`
	EdgeHeaderEnforced    bool   `envconfig:"EDGE_HEADER_ENFORCED" default:"false"`
	EdgeWildcardTLSSecret string `envconfig:"EDGE_WILDCARD_TLS_SECRET" default:"lucity-app-wildcard-tls"`

	// Database backups. One archive bucket for the whole platform, addressed by a
	// per-workspace prefix. Retention and schedule are product decisions and live
	// in internal/resources, not here.
	DatabaseBackupEnabled         bool   `envconfig:"DATABASE_BACKUP_ENABLED" default:"false"`
	DatabaseBackupEndpoint        string `envconfig:"DATABASE_BACKUP_S3_ENDPOINT"`
	DatabaseBackupBucket          string `envconfig:"DATABASE_BACKUP_S3_BUCKET"`
	DatabaseBackupRegion          string `envconfig:"DATABASE_BACKUP_S3_REGION"`
	DatabaseBackupAccessKeyID     string `envconfig:"DATABASE_BACKUP_S3_ACCESS_KEY_ID"`
	DatabaseBackupSecretAccessKey string `envconfig:"DATABASE_BACKUP_S3_SECRET_ACCESS_KEY"`

	// Agents. Their model and connector calls go through the shared agent
	// gateway, whose access log VictoriaLogs keeps for the audit.
	AgentsEnabled   bool   `envconfig:"AGENTS_ENABLED" default:"false"`
	AgentsNamespace string `envconfig:"AGENTS_NAMESPACE" default:"lucity-agents"`
	AgentImage      string `envconfig:"AGENT_IMAGE"`
	AgentGatewayURL string `envconfig:"AGENT_GATEWAY_URL" default:"http://lucity-agent-gateway.lucity-agents.svc.cluster.local"`
	VictoriaLogsURL string `envconfig:"VICTORIA_LOGS_URL"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "run-deploy" {
		runDeploy()
		return
	}

	var config Config
	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if config.EdgeHeaderEnforced && config.EdgeHeaderSecret == "" {
		slog.Error("EDGE_HEADER_ENFORCED requires EDGE_HEADER_SECRET")
		os.Exit(1)
	}

	if config.AgentsEnabled && config.AgentImage == "" {
		slog.Error("AGENTS_ENABLED requires AGENT_IMAGE")
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	ctx, cancel := graceful.Context()
	defer cancel()

	apiAudience := config.OIDCClientID
	if config.OIDCAudience != "" {
		apiAudience = config.OIDCAudience
	}

	var oidcHTTP *http.Client
	if config.OIDCDiscoveryURL != "" {
		client, err := newIssuerRewriteClient(config.OIDCIssuerURL, config.OIDCDiscoveryURL)
		if err != nil {
			slog.Error("failed to create OIDC HTTP client", "error", err)
			os.Exit(1)
		}
		oidcHTTP = client
	}
	oidcProvider := &oidc.Provider{
		Endpoint:     strings.TrimSuffix(config.OIDCIssuerURL, "/oidc"),
		ClientID:     config.OIDCClientID,
		Audience:     apiAudience,
		DirectSignIn: directSignIn,
		Scopes:       loginScopes,
		HTTP:         oidcHTTP,
	}

	verifier, err := auth.NewVerifier(ctx, config.OIDCIssuerURL, apiAudience)
	if err != nil {
		slog.Error("failed to create JWT verifier", "error", err)
		os.Exit(1)
	}

	k8sClient, dynClient, _, restConfig, err := buildKubeClients()
	if err != nil {
		slog.Error("failed to build kube clients", "error", err)
		os.Exit(1)
	}

	var billingClient *billing.Client
	if config.StripeSecretKey != "" {
		metered := []string{
			config.StripeEcoCPUPriceID,
			config.StripeEcoMemPriceID,
			config.StripeEcoDiskPriceID,
			config.StripeProdCPUPriceID,
			config.StripeProdMemPriceID,
			config.StripeProdDiskPriceID,
			config.StripeStoragePriceID,
			config.StripeEgressPriceID,
		}

		if config.StripeHobbyPriceID == "" || config.StripeProPriceID == "" || slices.Contains(metered, "") {
			slog.Error("billing needs all ten STRIPE_*_PRICE_ID settings with STRIPE_SECRET_KEY")
			os.Exit(1)
		}

		billingClient = billing.New(config.StripeSecretKey, billing.Prices{
			Hobby:   config.StripeHobbyPriceID,
			Pro:     config.StripeProPriceID,
			Metered: metered,
		})
		slog.Info("billing enabled")
	} else {
		slog.Info("STRIPE_SECRET_KEY not set, billing disabled")
	}

	githubApp, err := ghpkg.NewApp(config.GitHubAppID, config.GitHubClientID, config.GitHubClientSecret, config.GitHubPrivateKeyPath)

	if err != nil {
		slog.Error("failed to create github app", "error", err)
		os.Exit(1)
	}

	slog.Info("github app initialized", "app_id", config.GitHubAppID)

	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	verifier = verifier.WithOrgResolver(newOrgResolver(logtoClient))

	loadBalancerHostname := "lb." + config.WorkloadDomain

	secure := secureCookies(config.DashboardURL)
	sessionStore := newSessionStore(kvstore.NewMemory[sessionValue](), oidcProvider, logtoClient)
	sessionCodec := session.NewCodec(config.SessionSecret, sessionCookieName, secure, sessionCookieMaxAge)

	platformClient := platformK8s.New(k8sClient, dynClient)

	directoryClient, err := directoryLogto.New(logtoClient)

	if err != nil {
		slog.Error("failed to create directory client", "error", err)
		os.Exit(1)
	}

	jobsCache, err := jobs.NewCache(k8sClient, config.BuildNamespace, config.SystemNamespace)

	if err != nil {
		slog.Error("failed to create job cache", "error", err)
		os.Exit(1)
	}

	jobsClient := buildjobK8s.New(k8sClient, jobsCache, config.BuildNamespace, config.RegistryPushURL, config.RegistryAuthSecret, config.BuildImage, config.BuildkitTLSSecret, config.BuildkitServerName, config.BuildJobPlacement)

	deployJobsClient := deployjobK8s.New(k8sClient, jobsCache, deployjobK8s.Config{
		Namespace:            config.SystemNamespace,
		Image:                config.DeployImage,
		ServiceAccount:       config.DeployServiceAccount,
		BuildNamespace:       config.BuildNamespace,
		RegistryPullURL:      config.RegistryPullURL,
		GatewayName:          config.GatewayName,
		GatewayNS:            config.GatewayNamespace,
		GatewayHTTPSListener: config.GatewayHTTPSListener,
		ClusterIssuer:        config.CustomDomainClusterIssuer,
		Backups: deployjobK8s.BackupConfig{
			Enabled:  config.DatabaseBackupEnabled,
			Endpoint: config.DatabaseBackupEndpoint,
			Bucket:   config.DatabaseBackupBucket,
		},
		EdgeHeader:        config.EdgeHeaderEnforced,
		Placement:         config.DeployJobPlacement,
		WorkloadPlacement: config.WorkloadDefaultPlacement,
	})

	pipelineClient := pipeline.New(k8sClient, jobsCache, config.MaxConcurrentReleases, config.MaxConcurrentBuilds)

	scanJobsClient := scanjobK8s.New(k8sClient, jobsCache, scanjobK8s.Config{
		Namespace:             config.BuildNamespace,
		Image:                 config.BuildImage,
		Registry:              config.RegistryPushURL,
		RegistryAuthSecret:    config.RegistryAuthSecret,
		Timeout:               config.ScanTimeout,
		GitleaksWorkers:       config.ScanGitleaksWorkers,
		TrufflehogConcurrency: config.ScanTrufflehogConcurrency,
		Placement:             config.BuildJobPlacement,
	})

	secret, err := k8sClient.CoreV1().Secrets(config.SystemNamespace).Get(ctx, config.RegistryPullSecret, metav1.GetOptions{})

	if err != nil {
		slog.Error("failed to fetch registry pull secret", "error", err)
		os.Exit(1)
	}

	keychain, err := kauth.NewFromPullSecrets(ctx, []corev1.Secret{*secret})

	if err != nil {
		slog.Error("failed to parse registry pull secret", "error", err)
		os.Exit(1)
	}

	version, err := semver.Parse(Version)

	if err != nil {
		slog.Error("failed to parse version", "error", err, "version", Version)
		os.Exit(1)
	}

	source := sourceGH.New(githubApp)

	planner := railpack.New()

	chartRef, err := helmDeployer.LoadChartFromFS(charts.LucityApp, "lucity-app")

	if err != nil {
		slog.Error("failed to load lucity-app chart", "error", err)
		os.Exit(1)
	}

	chartRef.Metadata.Version = version.String()

	var agents helmDeployer.AgentConfig

	if config.AgentsEnabled {
		workspaceChart, err := helmDeployer.LoadChartFromFS(charts.LucityWorkspace, "lucity-workspace")

		if err != nil {
			slog.Error("failed to load lucity-workspace chart", "error", err)
			os.Exit(1)
		}

		workspaceChart.Metadata.Version = version.String()

		agents = helmDeployer.AgentConfig{
			Image:          config.AgentImage,
			GatewayURL:     config.AgentGatewayURL,
			Namespace:      config.AgentsNamespace,
			WorkspaceChart: workspaceChart,
		}
	}

	var workloadDefaultPlacement values.Placement

	if err := workloadDefaultPlacement.Decode(config.WorkloadDefaultPlacement); err != nil {
		slog.Error("failed to parse WORKLOAD_DEFAULT_PLACEMENT", "error", err)
		os.Exit(1)
	}

	var workloadPaidPlacement values.Placement

	if err := workloadPaidPlacement.Decode(config.WorkloadPaidPlacement); err != nil {
		slog.Error("failed to parse WORKLOAD_PAID_PLACEMENT", "error", err)
		os.Exit(1)
	}

	deployerClient, err := helmDeployer.New(chartRef, config.GatewayName, config.GatewayNamespace, config.CustomDomainClusterIssuer, helmDeployer.BackupConfig{
		Enabled:  config.DatabaseBackupEnabled,
		Endpoint: config.DatabaseBackupEndpoint,
		Bucket:   config.DatabaseBackupBucket,
	}, config.RegistryPullURL, workloadDefaultPlacement, agents, helmDeployer.WithEdgeHeader(config.EdgeHeaderEnforced), helmDeployer.WithHTTPSListener(config.GatewayHTTPSListener))

	if err != nil {
		slog.Error("failed to create deployer client", "error", err)
		os.Exit(1)
	}

	environmentClient := environmentK8s.New(k8sClient, dynClient, config.SystemNamespace, config.RegistryPullSecret, config.PodCIDR, config.ServiceCIDR, environmentK8s.BackupCredentials{
		AccessKeyID:     config.DatabaseBackupAccessKeyID,
		SecretAccessKey: config.DatabaseBackupSecretAccessKey,
		Region:          config.DatabaseBackupRegion,
	})

	ovhBackend, err := objectstorageOVH.New(
		config.OVHEndpoint,
		config.OVHAppKey,
		config.OVHAppSecret,
		config.OVHConsumerKey,
		config.OVHProjectID,
		config.OVHRegion,
	)

	if err != nil {
		slog.Error("failed to create object storage backend", "error", err)
		os.Exit(1)
	}

	var (
		storageBackend objectstorage.Backend = ovhBackend
		bunnyClient    *bunny.Client
	)

	if config.BunnyAPIKey != "" {
		bunnyClient = bunny.New(config.BunnyAPIKey)
		storageBackend = objectstorageBunny.New(ovhBackend, bunnyClient, config.PublicBucketDomain)
	} else {
		slog.Warn("BUNNY_API_KEY not set; public buckets disabled")
	}

	var edgeClient edge.Interface = edge.Disabled{}

	if bunnyClient != nil && config.EdgeOriginURL != "" {
		edgeClient = newEdge(bunnyClient, config)
	} else {
		slog.Warn("EDGE_ORIGIN_URL not set; edge disabled")
	}

	objectStorageClient := objectstorage.NewManager(storageBackend, k8sClient)

	hostnameClient := hostname.New(config.WorkloadDomain, loadBalancerHostname, config.IPAddress, edgeClient.EdgeAddresses, k8sClient, dynClient)

	domainsClient := domains.New(hostnameClient, edgeClient, platformClient, deployerClient, k8sClient, config.WorkloadDomain, config.GatewayNamespace, config.EdgeWildcardTLSSecret)

	metricsProvider, err := metrics.New(config.VictoriaMetricsURL)
	if err != nil {
		slog.Error("failed to create metrics provider", "error", err)
		os.Exit(1)
	}

	chartFS, err := fs.Sub(charts.LucityApp, "lucity-app")

	if err != nil {
		slog.Error("failed to open lucity-app chart fs", "error", err)
		os.Exit(1)
	}

	conductorConfig := conductor.Config{
		Version:              Version,
		ChartFS:              chartFS,
		RegistryURL:          config.RegistryURL,
		RegistryPushURL:      config.RegistryURL,
		RegistryPullURL:      config.RegistryPullURL,
		RegistryPullSecret:   keychain,
		WorkloadDomain:       config.WorkloadDomain,
		DatabaseDomain:       config.DatabaseDomain,
		LoadBalancerHostname: loadBalancerHostname,
		LoadBalancerIP:       config.IPAddress,
		GitHubAppSlug:        config.GitHubAppSlug,
		DashboardURL:         config.DashboardURL,
		MaxQueuedReleases:    config.MaxQueuedReleases,
		PaidPlacement:        workloadPaidPlacement,
		AgentsEnabled:        config.AgentsEnabled,
		BackupArchive: backuparchive.New(backuparchive.Config{
			Endpoint:        config.DatabaseBackupEndpoint,
			Bucket:          config.DatabaseBackupBucket,
			Region:          config.DatabaseBackupRegion,
			AccessKeyID:     config.DatabaseBackupAccessKeyID,
			SecretAccessKey: config.DatabaseBackupSecretAccessKey,
		}),
	}

	scanReportClient := scanreport.New(scanreport.Config{
		Endpoint:     config.RegistryPullURL,
		DialEndpoint: config.RegistryURL,
		Keychain:     keychain,
	})

	vulnerabilitiesClient := vulnerabilities.New(vulnerabilities.Config{
		Endpoint:     config.RegistryPullURL,
		DialEndpoint: config.RegistryURL,
		Keychain:     keychain,
	})

	agentRunnerClient, err := agentrunner.New(restConfig, k8sClient)

	if err != nil {
		slog.Error("failed to create agent runner client", "error", err)
		os.Exit(1)
	}

	agentAuditClient := agentaudit.New(config.VictoriaLogsURL, config.AgentsNamespace)

	registry := prometheus.NewRegistry()

	conductor := conductor.New(billingClient, githubApp, logtoClient, directoryClient, platformClient, jobsClient, deployJobsClient, scanJobsClient, scanReportClient, vulnerabilitiesClient, pipelineClient, planner, source, hostnameClient, deployerClient, environmentClient, objectStorageClient, metricsProvider, domainsClient, upstream.New(), agentRunnerClient, agentAuditClient, conductorConfig, registry)

	platformClient.Start(ctx)
	jobsCache.Start(ctx)

	slog.Info("release queue ready", "maxQueuedPerWorkspace", config.MaxQueuedReleases)

	githubActionsAudience := config.GitHubActionsAudience
	if githubActionsAudience == "" {
		githubActionsAudience = originFromURL(config.OIDCCallbackURL)
	}

	graphqlServer := NewGraphQLServer(config.Port, conductor, oidcProvider, sessionStore, sessionCodec, verifier, logtoClient, config.OIDCCallbackURL, config.DashboardURL, config.GitHubAppSlug, githubActionsAudience, config.OIDCIssuerURL, apiAudience, config.OIDCCLIClientID)

	servers := []graceful.Server{graphqlServer, metricsHTTP.NewServer(config.MetricsPort, registry)}

	if githubApp != nil {
		webhookHandler := &webhookhttp.Handler{
			GitHubApp: githubApp,
			Platform:  platformClient,
			Conductor: conductor,
		}
		webhookSrv := webhookhttp.NewServer(config.WebhookPort, config.WebhookSecret, webhookHandler, registry)
		servers = append(servers, webhookSrv)
		slog.Info("webhook receiver ready", "port", config.WebhookPort)
	} else {
		slog.Info("webhook receiver disabled (GitHub App not configured)")
	}

	graceful.Serve(ctx, servers...)
}

func buildKubeClients() (kubernetes.Interface, dynamic.Interface, metadata.Interface, *rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cfg.QPS = 50
	cfg.Burst = 100
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	meta, err := metadata.NewForConfig(cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return k8s, dyn, meta, cfg, nil
}
