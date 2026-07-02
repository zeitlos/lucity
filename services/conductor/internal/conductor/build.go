package conductor

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type Build = buildjob.Job
type BuildID = buildjob.BuildID
type Deploy = deployjob.Job
type DeployID = deployjob.DeployID

var _ platform.WorkspaceScoped = BuildID{}
var _ platform.WorkspaceScoped = DeployID{}

func (c *Client) Builds(ctx context.Context, workspace, repoURL, contextPath string) ([]Build, error) {
	return c.buildjob.List(ctx, workspace, repoURL, contextPath)
}

func (c *Client) Build(ctx context.Context, id BuildID) (*Build, error) {
	return c.buildjob.Get(ctx, id)
}

func (c *Client) BuildLogs(ctx context.Context, id BuildID) (<-chan string, error) {
	reader, err := c.buildjob.Logs(ctx, id)

	if err != nil {
		return nil, err
	}

	out := make(chan string, 128)

	go func() {
		defer close(out)
		defer reader.Close()

		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			select {
			case out <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (c *Client) DeployLogs(ctx context.Context, id DeployID) (<-chan string, error) {
	reader, err := c.deployjob.Logs(ctx, id)

	if err != nil {
		return nil, err
	}

	out := make(chan string, 128)

	go func() {
		defer close(out)
		defer reader.Close()

		scanner := bufio.NewScanner(reader)

		for scanner.Scan() {
			select {
			case out <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (c *Client) Deploy(ctx context.Context, serviceID ServiceID, gitRef string) (*Release, error) {
	service, err := c.platform.Service(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	ref := gitRef

	if ref == "" {
		ref = service.Branch
	}

	commit, err := c.source.Commit(ctx, service.SourceURL, ref)

	if err != nil {
		return nil, err
	}

	token, err := c.source.Token(ctx, service.SourceURL)

	if err != nil {
		return nil, err
	}

	if c.config.MaxQueuedReleases > 0 {
		queued, err := c.pipeline.QueuedRuns(ctx, service.ID.Workspace)

		if err != nil {
			slog.WarnContext(ctx, "queued release count failed", "error", err, "workspace", service.ID.Workspace)
		} else if queued >= c.config.MaxQueuedReleases {
			return nil, fmt.Errorf("deployment queue is full (%d queued) — wait for queued releases to finish", queued)
		}
	}

	claims, _ := auth.FromContext(ctx)
	release := deployer.NewRelease(deployer.TriggerManual, actorFromClaims(claims))

	imageName := service.ID.Workspace + "/" + service.ID.Project + "/" + service.Name

	build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
		Workspace:        service.ID.Workspace,
		RepoURL:          service.SourceURL,
		Commit:           commit.SHA,
		CommitMessage:    commit.Message,
		ContextPath:      service.ContextPath,
		TargetImageNames: []string{imageName},
		Token:            token,
		BuildVars:        service.Variables,
		ReleaseID:        release.ID,
	})

	if err != nil {
		return nil, err
	}

	deploy, err := c.startDeploy(ctx, service.ID, build.ID.Name, commit.Message, release)

	if err != nil {
		return nil, fmt.Errorf("start deploy for build %q: %w", build.ID.Name, err)
	}

	result := Release{
		ID:        ReleaseID{Workspace: serviceID.Workspace, Name: release.ID},
		Status:    releaseStatus(build, deploy, nil),
		Source:    gitSource(service.SourceURL, ref, service.ContextPath, Commit{SHA: commit.SHA, Message: commit.Message}),
		Trigger:   ReleaseTrigger{Kind: release.Trigger, Actor: release.Actor},
		Build:     build,
		Deploy:    deploy,
		CreatedAt: time.Now(),
	}

	return &result, nil
}

func (c *Client) startDeploy(ctx context.Context, serviceID platform.ServiceID, buildName, commitMessage string, release deployer.ReleaseMeta) (*Deploy, error) {
	return c.deployjob.Start(ctx, deployjob.StartOptions{
		Service:        serviceID,
		BuildName:      buildName,
		CommitMessage:  commitMessage,
		ReleaseID:      release.ID,
		ReleaseTrigger: string(release.Trigger),
		ReleaseActor:   release.Actor,
	})
}
