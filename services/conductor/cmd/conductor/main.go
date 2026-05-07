// Command conductor is the unified Lucity control-plane binary.
//
// Phase 3d wires the packager / deployer / builder gRPC services
// in-process via a bufconn-backed gRPC pipe. The handler still talks
// over the typed grpc.ClientConn interfaces it always has — the
// transport just no longer crosses a real network. This means:
//
//   - Streaming RPCs (deployer.ServiceLogs, builder.BuildLogs) keep
//     working without per-method adapter code.
//   - Auth interceptors are skipped on the in-process pipe; calls
//     between in-process components are trusted.
//   - Inbound external gRPC (cashier → conductor.SuspendWorkspace)
//     uses a separate, real listener with the internal-JWT verifier.
package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/zeitlos/lucity/pkg/auth"
	builderpb "github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/cashier"
	deployerpb "github.com/zeitlos/lucity/pkg/deployer"
	ghpkg "github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/pkg/graceful"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/pkg/logto"
	packagerpb "github.com/zeitlos/lucity/pkg/packager"

	"github.com/zeitlos/lucity/services/conductor/internal/api/handler"
	webhookpkg "github.com/zeitlos/lucity/services/conductor/internal/api/webhook"
	webhookhttp "github.com/zeitlos/lucity/services/conductor/internal/api/webhook/http"
	"github.com/zeitlos/lucity/services/conductor/internal/builder/build"
	"github.com/zeitlos/lucity/services/conductor/internal/builder/engine"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/argocd"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer/argo/gitops/softserve"
	inprocbuilder "github.com/zeitlos/lucity/services/conductor/internal/inproc/builder"
	inprocdeployer "github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	inprocpackager "github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
	conductorgrpc "github.com/zeitlos/lucity/services/conductor/internal/transport/grpc"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	GRPCPort    string `envconfig:"GRPC_PORT" default:"9090"`    // inbound from cashier (and similar)
	WebhookPort string `envconfig:"WEBHOOK_PORT" default:"9004"` // inbound from GitHub
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// OIDC (PKCE — no client secret needed)
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
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" default:"dev-secret"`

	// Soft-serve (GitOps repo storage)
	SoftServeSSH         string `envconfig:"SOFTSERVE_SSH_ADDR" default:"localhost:23231"`
	SoftServeHTTP        string `envconfig:"SOFTSERVE_HTTP_ADDR" default:"http://localhost:23232"`
	SoftServeClusterHTTP string `envconfig:"SOFTSERVE_CLUSTER_HTTP_ADDR"`
	SoftServeKeyPath     string `envconfig:"SOFTSERVE_SSH_KEY_PATH" required:"true"`
	SoftServeToken       string `envconfig:"SOFTSERVE_TOKEN"`

	// ArgoCD
	ArgocdAddr     string `envconfig:"ARGOCD_ADDR" required:"true"`
	ArgocdToken    string `envconfig:"ARGOCD_TOKEN" required:"true"`
	ArgocdInsecure bool   `envconfig:"ARGOCD_INSECURE" default:"false"`

	// Cluster (deployer)
	GatewayName        string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace   string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	ClusterIssuer      string `envconfig:"CLUSTER_ISSUER" default:"letsencrypt-http01"`
	RegistryPullSecret string `envconfig:"REGISTRY_PULL_SECRET" default:"lucity-registry-pull"`

	// Builder
	RegistryURL         string `envconfig:"REGISTRY_URL" default:"localhost:5000"`
	RegistryImagePrefix string `envconfig:"REGISTRY_IMAGE_PREFIX"`
	RegistryUsername    string `envconfig:"REGISTRY_USERNAME"`
	RegistryPassword    string `envconfig:"REGISTRY_PASSWORD"`
	RegistryAuthSecret  string `envconfig:"REGISTRY_AUTH_SECRET"`
	RegistryInsecure    bool   `envconfig:"REGISTRY_INSECURE" default:"true"`
	WorkDir             string `envconfig:"WORK_DIR" default:"/tmp/lucity-builds"`
	BuildImage          string `envconfig:"BUILD_IMAGE"`
	BuildkitAddr        string `envconfig:"BUILDKIT_ADDR"`
	BuildNamespace      string `envconfig:"BUILD_NAMESPACE" default:"lucity-builds"`

	// GitHub App (for installation tokens + OAuth)
	GitHubAppID            int64  `envconfig:"GITHUB_APP_ID"`
	GitHubPrivateKeyPath   string `envconfig:"GITHUB_PRIVATE_KEY_PATH"`
	GitHubClientID         string `envconfig:"GITHUB_CLIENT_ID"`
	GitHubClientSecret     string `envconfig:"GITHUB_CLIENT_SECRET"`
	GitHubOAuthCallbackURL string `envconfig:"GITHUB_OAUTH_CALLBACK_URL" default:"http://localhost:8080/auth/github/callback"`
	GitHubAppSlug          string `envconfig:"GITHUB_APP_SLUG"`

	// Domains
	WorkloadDomain string `envconfig:"WORKLOAD_DOMAIN" required:"true"`
	IPAddress      string `envconfig:"IP_ADDRESS"`

	// Internal JWT (ES256)
	InternalJWTPrivateKeyPath string `envconfig:"INTERNAL_JWT_PRIVATE_KEY_PATH"`
	InternalJWTPublicKeyPath  string `envconfig:"INTERNAL_JWT_PUBLIC_KEY_PATH"`
}

func main() {
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

	// ---- Soft-serve forge ----
	keyData, err := os.ReadFile(config.SoftServeKeyPath)
	if err != nil {
		slog.Error("failed to read soft-serve key", "error", err, "path", config.SoftServeKeyPath)
		os.Exit(1)
	}
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		slog.Error("failed to parse soft-serve key", "error", err)
		os.Exit(1)
	}
	forge := softserve.New(config.SoftServeSSH, signer, config.SoftServeHTTP, config.SoftServeToken)
	slog.Info("soft-serve forge ready", "ssh", config.SoftServeSSH, "http", config.SoftServeHTTP)

	// ---- ArgoCD client ----
	argoClient := argocd.NewClient(config.ArgocdAddr, config.ArgocdToken, config.ArgocdInsecure)

	// ---- Kubernetes clients ----
	k8sClient, dynClient, err := buildKubeClients()
	if err != nil {
		slog.Error("failed to build kube clients", "error", err)
		os.Exit(1)
	}

	// ---- Build engine ----
	if config.BuildImage == "" {
		slog.Error("BUILD_IMAGE is required (the lucity image used inside Build Job pods)")
		os.Exit(1)
	}
	if err := os.MkdirAll(config.WorkDir, 0o755); err != nil {
		slog.Error("failed to create work dir", "error", err, "path", config.WorkDir)
		os.Exit(1)
	}
	buildEng := engine.NewKubernetesEngine(engine.KubernetesEngineOpts{
		Client:             k8sClient,
		Namespace:          config.BuildNamespace,
		BuildImage:         config.BuildImage,
		BuildkitAddr:       config.BuildkitAddr,
		RegistryURL:        config.RegistryURL,
		RegistryAuthSecret: config.RegistryAuthSecret,
		Insecure:           config.RegistryInsecure,
	})
	buildTracker := build.NewK8sTracker(k8sClient, config.BuildNamespace)

	// ---- Construct in-process server impls ----
	//
	// Cycle: packager Server holds a deployer client; deployer Server
	// holds a packager client. We construct both with nil cross-refs,
	// then wire via SetDeployer / SetPackager once the bufconn-backed
	// clients are available below.
	clusterHTTP := config.SoftServeClusterHTTP
	if clusterHTTP == "" {
		clusterHTTP = config.SoftServeHTTP
	}

	packagerSvc := inprocpackager.NewServer(forge, nil, internalIssuer, config.WorkloadDomain)
	deployerSvc := inprocdeployer.NewServer(argoClient, nil, clusterHTTP, config.SoftServeToken, k8sClient, dynClient, internalIssuer, config.GatewayName, config.GatewayNamespace, config.ClusterIssuer, config.RegistryPullSecret)
	builderSvc := inprocbuilder.NewServer(buildEng, buildTracker, config.RegistryURL, config.RegistryUsername, config.RegistryPassword, config.RegistryInsecure, config.WorkDir)

	// ---- bufconn-backed in-process gRPC ----
	//
	// Register all three server impls on a single grpc.Server reachable
	// only via an in-memory pipe. The handler holds standard gRPC
	// clients that dial through that pipe.
	//
	// No auth interceptors here: in-process is implicitly trusted, and
	// adding the interceptor would require minting an internal JWT for
	// every RPC just to talk to ourselves.
	const bufSize = 4 * 1024 * 1024
	listener := bufconn.Listen(bufSize)
	inprocServer := grpc.NewServer()
	packagerpb.RegisterPackagerServiceServer(inprocServer, packagerSvc)
	deployerpb.RegisterDeployerServiceServer(inprocServer, deployerSvc)
	builderpb.RegisterBuilderServiceServer(inprocServer, builderSvc)
	go func() {
		if err := inprocServer.Serve(listener); err != nil {
			slog.Error("in-process gRPC server stopped", "error", err)
		}
	}()
	defer inprocServer.GracefulStop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("failed to dial in-process gRPC", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	packagerClient := packagerpb.NewPackagerServiceClient(conn)
	deployerClient := deployerpb.NewDeployerServiceClient(conn)
	builderClient := builderpb.NewBuilderServiceClient(conn)

	// Now that the clients exist, break the construction cycle.
	packagerSvc.SetDeployer(deployerClient)
	deployerSvc.SetPackager(packagerClient)

	// Custom-domain reconciliation runs on the deployer service.
	go reconcileCustomDomains(ctx, deployerSvc)

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

	// ---- GitHub App + Logto ----
	var githubApp *ghpkg.App
	if config.GitHubAppID != 0 && config.GitHubPrivateKeyPath != "" {
		githubApp, err = ghpkg.NewApp(config.GitHubAppID, config.GitHubClientID, config.GitHubClientSecret, "", config.GitHubOAuthCallbackURL, config.GitHubPrivateKeyPath)
		if err != nil {
			slog.Error("failed to create github app", "error", err)
			os.Exit(1)
		}
		slog.Info("github app initialized", "app_id", config.GitHubAppID)
	} else {
		slog.Info("github app not configured — repo listing and commit enrichment disabled")
	}

	logtoClient := logto.New(config.LogtoEndpoint, config.LogtoM2MAppID, config.LogtoM2MAppSecret)
	slog.Info("logto management API configured", "endpoint", config.LogtoEndpoint)

	// ---- Handler ----
	registryImagePrefix := config.RegistryImagePrefix
	if registryImagePrefix == "" {
		registryImagePrefix = config.RegistryURL
	}
	domainTarget := "lb." + config.WorkloadDomain

	secure := secureCookies(config.DashboardURL)
	tokenRefresher := newTokenRefresher(oidcProvider, secure)

	api := handler.New(packagerClient, builderClient, deployerClient, cashierClient, internalIssuer, githubApp, logtoClient, tokenRefresher, config.RegistryURL, registryImagePrefix, config.WorkloadDomain, domainTarget, config.IPAddress, config.GitHubAppSlug, config.DashboardURL)

	// ---- Servers ----
	components := []grpcComponent{}
	if cashierConn != nil {
		components = append(components, grpcComponent{name: "cashier", conn: cashierConn})
	}

	graphqlServer := NewGraphQLServer(config.Port, api, oidcProvider, verifier, logtoClient, internalIssuer, sessionSecret, config.DashboardURL, config.GitHubAppSlug, components)

	servers := []graceful.Server{graphqlServer}

	// Webhook receiver (GitHub push/PR events). Wire only when GitHub
	// App credentials are configured — otherwise the pipeline can't
	// authenticate to clone source repos.
	if githubApp != nil {
		webhookHandler := &webhookhttp.Handler{
			GitHubApp:      githubApp,
			InternalIssuer: internalIssuer,
			Pipeline: &webhookpkg.Pipeline{
				Builder:         builderClient,
				Packager:        packagerClient,
				Deployer:        deployerClient,
				RegistryPushURL: config.RegistryURL,
			},
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
		grpcSvc := conductorgrpc.NewService(api)
		grpcSrv := conductorgrpc.NewServer(":"+config.GRPCPort, grpcSvc, internalVerifier)
		slog.Info("conductor gRPC ready", "port", config.GRPCPort)
		servers = append(servers, grpcSrv)
	} else {
		slog.Info("internal JWT public key not set — conductor gRPC server disabled")
	}

	graceful.Serve(ctx, servers...)
}

// reconcileCustomDomains runs the periodic Gateway listener / cert
// reconciliation loop on the in-process deployer service. Was the
// goroutine in services/deployer/cmd/deployer/main.go.
func reconcileCustomDomains(ctx context.Context, dep *inprocdeployer.Server) {
	if err := dep.ReconcileCustomDomains(ctx); err != nil {
		slog.Warn("initial custom domain reconciliation failed", "error", err)
	}
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := dep.ReconcileCustomDomains(ctx); err != nil {
				slog.Warn("custom domain reconciliation failed", "error", err)
			}
		}
	}
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
