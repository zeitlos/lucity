package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/data"
	inprocbuilder "github.com/zeitlos/lucity/services/conductor/internal/inproc/builder"
	inprocdeployer "github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	inprocpackager "github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
)

// Pipeline orchestrates build+deploy for webhook-triggered CI/CD.
// This is a simplified, fire-and-forget version of the gateway's deploy flow.
type Pipeline struct {
	Builder         *inprocbuilder.Server
	Packager        *inprocpackager.Server
	Deployer        *inprocdeployer.Server
	RegistryPushURL string
}

// Run executes the full build+deploy pipeline for a service.
// Blocks until completion — callers should run this in a goroutine.
func (p *Pipeline) Run(ctx context.Context, workspace, project, service, environment, gitRef, sourceURL, contextPath string) {
	log := slog.With("workspace", workspace, "project", project, "service", service, "environment", environment, "gitRef", gitRef)
	log.Info("pipeline: starting build")

	registry := deriveImagePath(p.RegistryPushURL, workspace, project, service)

	buildID, err := p.Builder.StartBuild(ctx, sourceURL, gitRef, service, registry, contextPath)
	if err != nil {
		log.Error("pipeline: failed to start build", "error", err)
		return
	}

	log.Info("pipeline: build started", "buildId", buildID)

	// Poll build status until terminal.
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		status, err := p.Builder.BuildStatus(ctx, buildID)
		if err != nil {
			log.Error("pipeline: failed to poll build status", "error", err)
			return
		}

		switch status.Phase {
		case data.BuildPhaseSucceeded:
			log.Info("pipeline: build succeeded", "imageRef", status.ImageRef)
			p.finalize(ctx, log, workspace, project, service, environment, status.ImageRef, status.Digest)
			return
		case data.BuildPhaseFailed:
			log.Error("pipeline: build failed", "error", status.Error)
			return
		}
	}

	log.Error("pipeline: build timed out")
}

func (p *Pipeline) finalize(ctx context.Context, log *slog.Logger, workspace, project, service, environment, imageRef, digest string) {
	tag := extractTag(imageRef)

	log.Info("pipeline: updating gitops repo", "tag", tag)
	if err := p.Packager.UpdateImageTag(ctx, workspace, project, environment, service, tag, digest, ""); err != nil {
		log.Error("pipeline: failed to update image tag", "error", err)
		return
	}

	log.Info("pipeline: triggering sync")
	if _, err := p.Deployer.SyncDeployment(ctx, workspace, project, environment); err != nil {
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
