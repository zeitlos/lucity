package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/builder"
	"github.com/zeitlos/lucity/pkg/deployer"
	"github.com/zeitlos/lucity/pkg/packager"
	"github.com/zeitlos/lucity/pkg/tenant"
)

// Pipeline orchestrates build+deploy for webhook-triggered CI/CD.
// This is a simplified, fire-and-forget version of the gateway's deploy flow.
type Pipeline struct {
	Builder         builder.BuilderServiceClient
	Packager        packager.PackagerServiceClient
	Deployer        deployer.DeployerServiceClient
	RegistryPushURL string
}

// Run executes the full build+deploy pipeline for a service.
// Blocks until completion — callers should run this in a goroutine.
func (p *Pipeline) Run(ctx context.Context, project, service, environment, gitRef, sourceURL, contextPath string) {
	ws := tenant.FromContext(ctx)
	log := slog.With("project", project, "service", service, "environment", environment, "gitRef", gitRef)
	log.Info("pipeline: starting build")

	registry := deriveImagePath(p.RegistryPushURL, ws, project, service)

	buildResp, err := p.Builder.StartBuild(ctx, &builder.StartBuildRequest{
		SourceUrl:   sourceURL,
		GitRef:      gitRef,
		Service:     service,
		Registry:    registry,
		ContextPath: contextPath,
	})
	if err != nil {
		log.Error("pipeline: failed to start build", "error", err)
		return
	}

	log.Info("pipeline: build started", "buildId", buildResp.BuildId)

	// Poll build status until terminal.
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		status, err := p.Builder.BuildStatus(ctx, &builder.BuildStatusRequest{BuildId: buildResp.BuildId})
		if err != nil {
			log.Error("pipeline: failed to poll build status", "error", err)
			return
		}

		switch status.Phase {
		case builder.BuildPhase_BUILD_PHASE_SUCCEEDED:
			log.Info("pipeline: build succeeded", "imageRef", status.ImageRef)
			p.finalize(ctx, log, project, service, environment, status.ImageRef, status.Digest)
			return
		case builder.BuildPhase_BUILD_PHASE_FAILED:
			log.Error("pipeline: build failed", "error", status.Error)
			return
		}
	}

	log.Error("pipeline: build timed out")
}

func (p *Pipeline) finalize(ctx context.Context, log *slog.Logger, project, service, environment, imageRef, digest string) {
	tag := extractTag(imageRef)

	log.Info("pipeline: updating gitops repo", "tag", tag)
	_, err := p.Packager.UpdateImageTag(ctx, &packager.UpdateImageTagRequest{
		Project:     project,
		Environment: environment,
		Service:     service,
		Tag:         tag,
		Digest:      digest,
	})
	if err != nil {
		log.Error("pipeline: failed to update image tag", "error", err)
		return
	}

	log.Info("pipeline: triggering sync")
	_, err = p.Deployer.SyncDeployment(ctx, &deployer.SyncDeploymentRequest{
		Project:     project,
		Environment: environment,
	})
	if err != nil {
		log.Warn("pipeline: sync trigger failed (auto-sync will pick it up)", "error", err)
	}

	log.Info("pipeline: deploy complete", "tag", tag)
}

// deriveImagePath builds a workspace-scoped registry image path.
func deriveImagePath(registryURL, workspace, project, service string) string {
	return fmt.Sprintf("%s/%s/%s/%s", registryURL, workspace, project, service)
}

// extractTag extracts the tag from a fully-qualified image reference.
func extractTag(imageRef string) string {
	if i := strings.LastIndex(imageRef, ":"); i >= 0 {
		if j := strings.LastIndex(imageRef, "/"); i > j {
			return imageRef[i+1:]
		}
	}
	return imageRef
}
