package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/blang/semver/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/zeitlos/lucity/charts"
	"github.com/zeitlos/lucity/pkg/logger"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	buildjobK8s "github.com/zeitlos/lucity/services/conductor/internal/buildjob/kubernetes"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	helmDeployer "github.com/zeitlos/lucity/services/conductor/internal/deployer/helm"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type DeployConfig struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	Workspace   string `envconfig:"DEPLOY_WORKSPACE" required:"true"`
	Project     string `envconfig:"DEPLOY_PROJECT" required:"true"`
	Environment string `envconfig:"DEPLOY_ENVIRONMENT" required:"true"`
	Service     string `envconfig:"DEPLOY_SERVICE" required:"true"`

	BuildName      string `envconfig:"DEPLOY_BUILD_NAME" required:"true"`
	BuildNamespace string `envconfig:"DEPLOY_BUILD_NAMESPACE" required:"true"`

	CommitMessage  string `envconfig:"DEPLOY_COMMIT_MESSAGE"`
	ReleaseID      string `envconfig:"DEPLOY_RELEASE_ID"`
	ReleaseTrigger string `envconfig:"DEPLOY_RELEASE_TRIGGER" default:"manual"`
	ReleaseActor   string `envconfig:"DEPLOY_RELEASE_ACTOR"`

	RegistryPullURL  string `envconfig:"REGISTRY_PULL_URL" required:"true"`
	GatewayName      string `envconfig:"GATEWAY_NAME" default:"lucity-gateway"`
	GatewayNamespace string `envconfig:"GATEWAY_NAMESPACE" default:"lucity-system"`
	ClusterIssuer    string `envconfig:"CUSTOM_DOMAIN_CLUSTER_ISSUER" default:"letsencrypt-http01"`

	// This process runs the same values apply as the conductor, so it needs the
	// same archive-store settings or it would strip archiving on every deploy.
	DatabaseBackupEnabled  bool   `envconfig:"DATABASE_BACKUP_ENABLED" default:"false"`
	DatabaseBackupEndpoint string `envconfig:"DATABASE_BACKUP_S3_ENDPOINT"`
	DatabaseBackupBucket   string `envconfig:"DATABASE_BACKUP_S3_BUCKET"`
}

const buildWaitTimeout = 30 * time.Minute

func runDeploy() {
	var config DeployConfig

	if err := envconfig.Process("", &config); err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Setup(config.LogLevel)

	serviceID := platform.ServiceID{
		Workspace:   config.Workspace,
		Project:     config.Project,
		Environment: config.Environment,
		Name:        config.Service,
	}

	log := slog.With(
		"build", config.BuildName,
		"project", serviceID.Project,
		"service", serviceID.Name,
		"environment", serviceID.Environment,
	)

	ctx, cancel := context.WithTimeout(context.Background(), buildWaitTimeout)
	defer cancel()

	k8sClient, _, err := buildKubeClients()

	if err != nil {
		log.Error("failed to build kube clients", "error", err)
		os.Exit(1)
	}

	builds := buildjobK8s.New(k8sClient, config.BuildNamespace, "", "", "", "", "")
	buildID := buildjob.BuildID{Workspace: config.Workspace, Name: config.BuildName}

	log.Info("deploy: waiting for build")

	build, err := waitForBuild(ctx, builds, buildID, log)

	if err != nil {
		log.Error("deploy: failed waiting for build", "error", err)
		os.Exit(1)
	}

	if build.Status != buildjob.StatusSucceeded {
		log.Warn("deploy: build did not succeed, nothing to deploy", "status", string(build.Status))
		return
	}

	ref, err := build.BuiltImage(serviceID.ImageRepository())

	if err != nil {
		log.Error("deploy: build succeeded but image unusable", "error", err)
		os.Exit(1)
	}

	ref.Repository = config.RegistryPullURL + "/" + ref.Repository

	log.Info("deploy: build succeeded, applying image", "ref", ref.String())

	version, err := semver.Parse(Version)

	if err != nil {
		log.Error("deploy: failed to parse version", "error", err, "version", Version)
		os.Exit(1)
	}

	chartRef, err := helmDeployer.LoadChartFromFS(charts.LucityApp, "lucity-app")

	if err != nil {
		log.Error("deploy: failed to load lucity-app chart", "error", err)
		os.Exit(1)
	}

	chartRef.Metadata.Version = version.String()

	deployerClient, err := helmDeployer.New(chartRef, config.GatewayName, config.GatewayNamespace, config.ClusterIssuer, helmDeployer.BackupConfig{
		Enabled:  config.DatabaseBackupEnabled,
		Endpoint: config.DatabaseBackupEndpoint,
		Bucket:   config.DatabaseBackupBucket,
	})

	if err != nil {
		log.Error("deploy: failed to create deployer client", "error", err)
		os.Exit(1)
	}

	provenance := deployer.ImageProvenance{
		Commit:        build.Commit,
		CommitMessage: config.CommitMessage,
		BuildID:       buildID.String(),
	}

	release := deployer.ReleaseMeta{
		ID:      config.ReleaseID,
		Trigger: deployer.TriggerKind(config.ReleaseTrigger),
		Actor:   config.ReleaseActor,
	}

	if _, err := deployerClient.Services().SetImage(ctx, serviceID, ref, provenance, release); err != nil {
		log.Error("deploy: set image failed", "error", err)
		os.Exit(1)
	}

	log.Info("deploy: complete")
}

func waitForBuild(ctx context.Context, builds buildjob.Interface, id buildjob.BuildID, log *slog.Logger) (*buildjob.Job, error) {
	for {
		build, err := builds.Get(ctx, id)

		if err != nil {
			return nil, err
		}

		switch build.Status {
		case buildjob.StatusSucceeded, buildjob.StatusFailed, buildjob.StatusCancelled:
			return build, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
