package main

import (
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/blang/semver/v4"
	"github.com/zeitlos/lucity/charts"
	webhookhttp "github.com/zeitlos/lucity/services/conductor/internal/api/webhook/http"
	buildjobK8s "github.com/zeitlos/lucity/services/conductor/internal/buildjob/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/conductor"
	helmDeployer "github.com/zeitlos/lucity/services/conductor/internal/deployer/helm"
	deployjobK8s "github.com/zeitlos/lucity/services/conductor/internal/deployjob/kubernetes"
	directoryLogto "github.com/zeitlos/lucity/services/conductor/internal/directory/logto"
	environmentK8s "github.com/zeitlos/lucity/services/conductor/internal/environment/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/gateway"
	"github.com/zeitlos/lucity/services/conductor/internal/hostname"
	"github.com/zeitlos/lucity/services/conductor/internal/metrics"
	"github.com/zeitlos/lucity/services/conductor/internal/objectstorage"
	objectstorageOVH "github.com/zeitlos/lucity/services/conductor/internal/objectstorage/ovh"
	"github.com/zeitlos/lucity/services/conductor/internal/pipeline"
	"github.com/zeitlos/lucity/services/conductor/internal/planner/railpack"
	platformK8s "github.com/zeitlos/lucity/services/conductor/internal/platform/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/resources"
	scanjobK8s "github.com/zeitlos/lucity/services/conductor/internal/scanjob/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/scanreport"
	sourceGH "github.com/zeitlos/lucity/services/conductor/internal/source/github"
	conductorgrpc "github.com/zeitlos/lucity/services/conductor/internal/transport/grpc"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/pkg/cashier"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"

	kauth "github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	GRPCPort    string `envconfig:"GRPC_PORT" default:"9090"`
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9004"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// OIDC
	OIDCIssuerURL    string `envconfig:"OIDC_ISSUER_URL" required:"true"`
	OIDCDiscoveryURL string `envconfig:"OIDC_DISCOVERY_URL"`
	OIDCClientID     string `envconfig:"OIDC_CLIENT_ID" required:"true"`
	OIDCCallbackURL  string `envconfig:"OIDC_CALLBACK_URL" default:"http://localhost:8080/auth/callback"`

	// Auth
	DashboardURL   string `envconfig:"DASHBOARD_URL" default:"http://localhost:5173"`
	SessionSecret  string `envconfig:"SESSION_SECRET" required:"true"`
	AuthTestSecret string `envconfig:"AUTH_TEST_SECRET"`

	// Logto Management API (M2M)
	LogtoEndpoint     string `envconfig:"LOGTO_ENDPOINT" required:"true"`
	LogtoM2MAppID     string `envconfig:"LOGTO_M2M_APP_ID" required:"true"`
	LogtoM2MAppSecret string `envconfig:"LOGTO_M2M_APP_SECRET" required:"true"`

	// External services (still gRPC)
	CashierAddr string `envconfig:"CASHIER_ADDR"`

	// Webhook (GitHub events)
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" required:"true"`

	// Cluster
	GatewayName        string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace   string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	RegistryPullSecret string `envconfig:"REGISTRY_PULL_SECRET" default:"lucity-registry-pull"`

	// Per-env NetworkPolicy needs the cluster's pod and service CIDRs to
	// carve out "internet but not the cluster" egress. These are
	// cluster-wide constants; defaults match kubeadm + K8s defaults.
	PodCIDR     string `envconfig:"POD_CIDR" default:"10.244.0.0/16"`
	ServiceCIDR string `envconfig:"SERVICE_CIDR" default:"10.96.0.0/12"`

	// Builder
	RegistryURL        string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryPullURL    string `envconfig:"REGISTRY_PULL_URL" required:"true"`
	RegistryPushURL    string `envconfig:"REGISTRY_PUSH_URL" required:"true"`
	RegistryAuthSecret string `envconfig:"REGISTRY_AUTH_SECRET" required:"true"`
	BuildImage         string `envconfig:"BUILD_IMAGE"`
	BuildkitAddr       string `envconfig:"BUILDKIT_ADDR"`
	BuildNamespace     string `envconfig:"BUILD_NAMESPACE" default:"lucity-builds"`
	BuildkitTLSSecret  string `envconfig:"BUILDKIT_TLS_SECRET"`
	BuildkitServerName string `envconfig:"BUILDKIT_SERVER_NAME"`

	// Deploy jobs
	DeployImage          string `envconfig:"DEPLOY_IMAGE" required:"true"`
	DeployServiceAccount string `envconfig:"DEPLOY_SERVICE_ACCOUNT"`

	// Secret scan tuning
	ScanTimeout               time.Duration `envconfig:"SCAN_TIMEOUT" default:"60m"`
	ScanGitleaksWorkers       int           `envconfig:"SCAN_GITLEAKS_WORKERS" default:"3"`
	ScanTrufflehogConcurrency int           `envconfig:"SCAN_TRUFFLEHOG_CONCURRENCY" default:"2"`

	MaxConcurrentReleases int `envconfig:"MAX_CONCURRENT_RELEASES" default:"5"`
	MaxQueuedReleases     int `envconfig:"MAX_QUEUED_RELEASES" default:"10"`

	SystemNamespace string `envconfig:"SYSTEM_NAMESPACE" default:"lucity-system"`

	// Observability
	VictoriaMetricsURL string `envconfig:"VICTORIA_METRICS_URL" required:"true"`

	// Keyless CI deploys (GitHub Actions OIDC)
	GitHubActionsAudience string        `envconfig:"GITHUB_ACTIONS_AUDIENCE"`
	CISessionTTL          time.Duration `envconfig:"CI_SESSION_TTL" default:"1h"`

	// GitHub App (for installation tokens + OAuth)
	GitHubAppID            int64  `envconfig:"GITHUB_APP_ID" required:"true"`
	GitHubPrivateKeyPath   string `envconfig:"GITHUB_PRIVATE_KEY_PATH" required:"true"`
	GitHubClientID         string `envconfig:"GITHUB_CLIENT_ID" required:"true"`
	GitHubClientSecret     string `envconfig:"GITHUB_CLIENT_SECRET" required:"true"`
	GitHubOAuthCallbackURL string `envconfig:"GITHUB_OAUTH_CALLBACK_URL" required:"true"`
	GitHubAppSlug          string `envconfig:"GITHUB_APP_SLUG" required:"true"`

	// Domains
	WorkloadDomain            string `envconfig:"WORKLOAD_DOMAIN" required:"true"`
	DatabaseDomain            string `envconfig:"DATABASE_DOMAIN" required:"true"`
	IPAddress                 string `envconfig:"IP_ADDRESS"`
	CustomDomainClusterIssuer string `envconfig:"CUSTOM_DOMAIN_CLUSTER_ISSUER" default:"letsencrypt-http01"`

	// Reconciliation
	ReconcileEnabled bool `envconfig:"RECONCILE_ENABLED" default:"true"`

	// Internal JWT (ES256)
	InternalJWTPrivateKeyPath string `envconfig:"INTERNAL_JWT_PRIVATE_KEY_PATH"`
	InternalJWTPublicKeyPath  string `envconfig:"INTERNAL_JWT_PUBLIC_KEY_PATH"`

	// OVH
	OVHEndpoint    string `envconfig:"OVH_ENDPOINT" default:"ovh-eu"`
	OVHAppKey      string `envconfig:"OVH_APPLICATION_KEY" required:"true"`
	OVHAppSecret   string `envconfig:"OVH_APPLICATION_SECRET" required:"true"`
	OVHConsumerKey string `envconfig:"OVH_CONSUMER_KEY" required:"true"`
	OVHProjectID   string `envconfig:"OVH_PROJECT_ID" required:"true"`
	OVHRegion      string `envconfig:"OVH_REGION" default:"GRA"`
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

	logger.Setup(config.LogLevel)

	ctx, cancel := graceful.Context()
	defer cancel()

	// ---- Auth: OIDC, session JWT, internal-JWT issuer ----
	oidcProvider, err := NewOIDCProvider(ctx, config.OIDCIssuerURL, config.OIDCDiscoveryURL, config.OIDCClientID, config.OIDCCallbackURL)
	if err != nil {
		slog.Error("failed to initialize OIDC provider", "error", err)
		os.Exit(1)
	}

	verifier, err := auth.NewVerifier(ctx, config.OIDCIssuerURL, config.OIDCClientID)
	if err != nil {
		slog.Error("failed to create JWT verifier", "error", err)
		os.Exit(1)
	}

	sessionSecret := config.SessionSecret
	if config.AuthTestSecret != "" {
		sessionSecret = config.AuthTestSecret
		slog.Warn("test token authentication enabled — do not use in production")
	}
	verifier = verifier.WithFallback(hmacValidateFunc(sessionSecret))

	var internalIssuer *auth.Issuer
	if config.InternalJWTPrivateKeyPath != "" {
		internalIssuer, err = auth.NewIssuerFromFile(config.InternalJWTPrivateKeyPath)
		if err != nil {
			slog.Error("failed to create internal JWT issuer", "error", err)
			os.Exit(1)
		}
		slog.Info("internal JWT issuer initialized (ES256)")
	} else {
		slog.Warn("internal JWT not configured — outgoing service-to-service calls use legacy plain metadata headers")
	}

	// ---- Kubernetes clients ----
	k8sClient, dynClient, err := buildKubeClients()
	if err != nil {
		slog.Error("failed to build kube clients", "error", err)
		os.Exit(1)
	}

	// ---- External cashier (still real gRPC) ----
	var cashierClient cashier.CashierServiceClient
	var cashierConn *grpc.ClientConn
	if config.CashierAddr != "" {
		cashierConn, err = grpc.NewClient(config.CashierAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			slog.Error("failed to connect to cashier", "error", err, "addr", config.CashierAddr)
			os.Exit(1)
		}
		defer cashierConn.Close()
		cashierClient = cashier.NewCashierServiceClient(cashierConn)
		slog.Info("billing enabled", "addr", config.CashierAddr)
	} else {
		slog.Info("cashier not configured — billing disabled")
	}

	githubApp, err := ghpkg.NewApp(config.GitHubAppID, config.GitHubClientID, config.GitHubClientSecret, "", config.GitHubOAuthCallbackURL, config.GitHubPrivateKeyPath)

	if err != nil {
		slog.Error("failed to create github app", "error", err)
		os.Exit(1)
	}

	slog.Info("github app initialized", "app_id", config.GitHubAppID)

	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	domainTarget := "lb." + config.WorkloadDomain

	secure := secureCookies(config.DashboardURL)
	tokenRefresher := newTokenRefresher(oidcProvider, secure)

	platformClient := platformK8s.New(k8sClient, dynClient)

	directoryClient, err := directoryLogto.New(logtoClient)

	if err != nil {
		slog.Error("failed to create directory client", "error", err)
		os.Exit(1)
	}

	jobsClient := buildjobK8s.New(k8sClient, config.BuildNamespace, config.RegistryPushURL, config.RegistryAuthSecret, config.BuildImage, config.BuildkitTLSSecret, config.BuildkitServerName)

	deployJobsClient := deployjobK8s.New(k8sClient, deployjobK8s.Config{
		Namespace:       config.SystemNamespace,
		Image:           config.DeployImage,
		ServiceAccount:  config.DeployServiceAccount,
		BuildNamespace:  config.BuildNamespace,
		RegistryPullURL: config.RegistryPullURL,
		GatewayName:     config.GatewayName,
		GatewayNS:       config.GatewayNamespace,
	})

	pipelineClient := pipeline.New(k8sClient, config.BuildNamespace, config.SystemNamespace, config.MaxConcurrentReleases)

	scanJobsClient := scanjobK8s.New(k8sClient, scanjobK8s.Config{
		Namespace:             config.BuildNamespace,
		Image:                 config.BuildImage,
		Registry:              config.RegistryPushURL,
		RegistryAuthSecret:    config.RegistryAuthSecret,
		Timeout:               config.ScanTimeout,
		GitleaksWorkers:       config.ScanGitleaksWorkers,
		TrufflehogConcurrency: config.ScanTrufflehogConcurrency,
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

	deployerClient, err := helmDeployer.New(chartRef, config.GatewayName, config.GatewayNamespace, resources.DefaultCPULimit, resources.DefaultMemoryLimit)

	if err != nil {
		slog.Error("failed to create deployer client", "error", err)
		os.Exit(1)
	}

	hostnameClient := hostname.New(config.WorkloadDomain, domainTarget, config.IPAddress, config.GatewayNamespace, k8sClient, dynClient)

	gatewayClient := gateway.New(dynClient, config.GatewayName, config.GatewayNamespace, config.CustomDomainClusterIssuer)

	environmentClient := environmentK8s.New(k8sClient, dynClient, config.SystemNamespace, config.RegistryPullSecret, config.PodCIDR, config.ServiceCIDR)

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

	objectStorageClient := objectstorage.NewManager(ovhBackend, k8sClient)

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
		LoadBalancerHostname: domainTarget,
		LoadBalancerIP:       config.IPAddress,
		GitHubAppSlug:        config.GitHubAppSlug,
		DashboardURL:         config.DashboardURL,
		MaxQueuedReleases:    config.MaxQueuedReleases,
	}
	scanReportClient := scanreport.New(scanreport.Config{
		Endpoint:     config.RegistryPullURL,
		DialEndpoint: config.RegistryURL,
		Keychain:     keychain,
	})

	conductor := conductor.New(cashierClient, githubApp, logtoClient, tokenRefresher, directoryClient, platformClient, jobsClient, deployJobsClient, scanJobsClient, scanReportClient, pipelineClient, planner, source, hostnameClient, gatewayClient, deployerClient, environmentClient, objectStorageClient, metricsProvider, conductorConfig)

	go runAdmissionReconciler(ctx, pipelineClient)
	slog.Info("release admission ready", "maxConcurrent", config.MaxConcurrentReleases, "maxQueuedPerWorkspace", config.MaxQueuedReleases)

	if config.ReconcileEnabled {
		go runDomainReconciler(ctx, conductor)
		go runServiceReconciler(ctx, conductor)
	} else {
		slog.Warn("reconcilers disabled")
	}

	// ---- Servers ----
	components := []grpcComponent{}
	if cashierConn != nil {
		components = append(components, grpcComponent{name: "cashier", conn: cashierConn})
	}

	githubActionsAudience := config.GitHubActionsAudience
	if githubActionsAudience == "" {
		githubActionsAudience = originFromURL(config.OIDCCallbackURL)
	}

	graphqlServer := NewGraphQLServer(config.Port, conductor, oidcProvider, verifier, logtoClient, internalIssuer, sessionSecret, config.DashboardURL, config.GitHubAppSlug, githubActionsAudience, config.CISessionTTL, components)

	servers := []graceful.Server{graphqlServer}

	// Webhook receiver (GitHub push/PR events). Wire only when GitHub
	// App credentials are configured — otherwise the pipeline can't
	// authenticate to clone source repos.
	if githubApp != nil {
		webhookHandler := &webhookhttp.Handler{
			GitHubApp: githubApp,
			Platform:  platformClient,
			Conductor: conductor,
		}
		webhookSrv := webhookhttp.NewServer(config.WebhookPort, config.WebhookSecret, webhookHandler)
		servers = append(servers, webhookSrv)
		slog.Info("webhook receiver ready", "port", config.WebhookPort)
	} else {
		slog.Info("webhook receiver disabled (GitHub App not configured)")
	}

	if config.InternalJWTPublicKeyPath != "" {
		internalVerifier, err := auth.NewInternalVerifierFromFile(config.InternalJWTPublicKeyPath)
		if err != nil {
			slog.Error("failed to create internal JWT verifier", "error", err)
			os.Exit(1)
		}
		grpcSvc := conductorgrpc.NewService(platformClient, deployerClient)
		grpcSrv := conductorgrpc.NewServer(":"+config.GRPCPort, grpcSvc, internalVerifier)
		slog.Info("conductor gRPC ready", "port", config.GRPCPort)
		servers = append(servers, grpcSrv)
	} else {
		slog.Info("internal JWT public key not set — conductor gRPC server disabled")
	}

	graceful.Serve(ctx, servers...)
}

// buildKubeClients constructs typed + dynamic Kubernetes clients
// using the standard kubeconfig loading rules (in-cluster first,
// then ~/.kube/config).
func buildKubeClients() (kubernetes.Interface, dynamic.Interface, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, &clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	cfg.QPS = 50
	cfg.Burst = 100
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}
	return k8s, dyn, nil
}
