package webhook

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	inprocdeployer "github.com/zeitlos/lucity/services/conductor/internal/inproc/deployer"
	inprocpackager "github.com/zeitlos/lucity/services/conductor/internal/inproc/packager"
	"github.com/zeitlos/lucity/services/conductor/internal/source"
)

// Pipeline orchestrates build+deploy for webhook-triggered CI/CD.
// This is a simplified, fire-and-forget version of the dashboard's deploy flow.
type Pipeline struct {
	Buildjob buildjob.Interface
	Source   source.Interface
	Packager *inprocpackager.Client
	Deployer *inprocdeployer.Client
}

// Run executes the full build+deploy pipeline for a service.
// Blocks until completion — callers should run this in a goroutine.
func (p *Pipeline) Run(ctx context.Context, workspace, project, service, environment, commitSHA, sourceURL, contextPath string) {
	log := slog.With(
		"workspace", workspace,
		"project", project,
		"service", service,
		"environment", environment,
		"commit", commitSHA,
	)
	log.Info("pipeline: starting build")

	token, err := p.Source.Token(ctx, sourceURL)

	if err != nil {
		log.Error("pipeline: failed to mint source token", "error", err)
		return
	}

	imageName := workspace + "/" + project + "/" + service

	build, err := p.Buildjob.Start(ctx, buildjob.StartOptions{
		Workspace:        workspace,
		RepoURL:          sourceURL,
		Commit:           commitSHA,
		ContextPath:      contextPath,
		TargetImageNames: []string{imageName},
		Token:            token,
	})

	if err != nil {
		log.Error("pipeline: failed to start build", "error", err)
		return
	}

	log = log.With("buildId", build.ID)
	log.Info("pipeline: build started")

	deadline := time.Now().Add(30 * time.Minute)

	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)

		job, err := p.Buildjob.Get(ctx, build.ID)

		if err != nil {
			log.Error("pipeline: failed to poll build", "error", err)
			return
		}

		switch job.Status {
		case buildjob.StatusSucceeded:
			if len(job.ImageRefs) == 0 {
				log.Error("pipeline: build succeeded but produced no image refs")
				return
			}

			tag := extractTag(job.ImageRefs[0])

			log.Info("pipeline: build succeeded", "tag", tag)
			p.finalize(ctx, log, workspace, project, service, environment, tag)

			return

		case buildjob.StatusFailed, buildjob.StatusCancelled:
			log.Warn("pipeline: build did not succeed", "status", string(job.Status))
			return
		}
	}

	log.Error("pipeline: build timed out")
}

func (p *Pipeline) finalize(ctx context.Context, log *slog.Logger, workspace, project, service, environment, tag string) {
	log.Info("pipeline: updating gitops repo")

	if err := p.Packager.UpdateImageTag(ctx, workspace, project, environment, service, tag, "", ""); err != nil {
		log.Error("pipeline: failed to update image tag", "error", err)
		return
	}

	log.Info("pipeline: triggering sync")

	if _, err := p.Deployer.SyncDeployment(ctx, workspace, project, environment); err != nil {
		log.Warn("pipeline: sync trigger failed (auto-sync will pick it up)", "error", err)
	}

	log.Info("pipeline: deploy complete")
}

// extractTag extracts the tag from a fully-qualified image reference.
// Port-safe: ignores colons that come before the last slash.
func extractTag(imageRef string) string {
	slashIdx := strings.LastIndex(imageRef, "/")
	colonIdx := strings.LastIndex(imageRef, ":")

	if colonIdx > slashIdx {
		return imageRef[colonIdx+1:]
	}

	return ""
}
