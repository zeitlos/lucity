package conductor

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
	"github.com/zeitlos/lucity/services/conductor/internal/scanreport"
)

type Build = buildjob.Job
type BuildID = buildjob.BuildID
type Deploy = deployjob.Job
type DeployID = deployjob.DeployID
type Scan = scanjob.Job
type ScanID = scanjob.ScanID
type SecretScanReport = scanreport.Report

var _ platform.WorkspaceScoped = BuildID{}
var _ platform.WorkspaceScoped = DeployID{}
var _ platform.WorkspaceScoped = ScanID{}

func (c *Client) Builds(ctx context.Context, service platform.ServiceID) ([]Build, error) {
	return c.buildjob.List(ctx, service)
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
	return c.deploy(ctx, serviceID, gitRef, deployer.TriggerManual)
}

func (c *Client) DeployPush(ctx context.Context, serviceID ServiceID, gitRef string) (*Release, error) {
	return c.deploy(ctx, serviceID, gitRef, deployer.TriggerPush)
}

func (c *Client) deploy(ctx context.Context, serviceID ServiceID, gitRef string, trigger deployer.TriggerKind) (*Release, error) {
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
	release := deployer.NewRelease(trigger, actorFromClaims(claims))

	imageName := service.ID.ImageRepository()

	build, err := c.buildjob.Start(ctx, buildjob.StartOptions{
		Service:          service.ID,
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

	c.startScan(ctx, service.ID, build.ID.Name, service.SourceURL, commit.SHA, token, release.ID)

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

func (c *Client) startScan(ctx context.Context, serviceID platform.ServiceID, buildName, sourceURL, commit, token, releaseID string) {
	_, err := c.scanjob.Start(ctx, scanjob.StartOptions{
		Service:   serviceID,
		BuildName: buildName,
		SourceURL: sourceURL,
		Commit:    commit,
		Token:     token,
		ReleaseID: releaseID,
	})

	if err != nil {
		slog.WarnContext(ctx, "scan start failed", "service", serviceID, "error", err)
	}
}

func (c *Client) ScanLogs(ctx context.Context, id ScanID) (<-chan string, error) {
	reader, err := c.scanjob.Logs(ctx, id)

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

func (c *Client) SecretScanReport(ctx context.Context, serviceID ServiceID) (*SecretScanReport, error) {
	report, err := c.scanreport.Latest(ctx, serviceID)

	if err != nil || report == nil {
		return report, err
	}

	service, err := c.platform.Service(ctx, serviceID)

	if err != nil {
		slog.WarnContext(ctx, "service lookup for finding links failed", "service", serviceID, "error", err)
		return report, nil
	}

	for i := range report.Findings {
		report.Findings[i].URL = findingURL(service.SourceURL, report.Findings[i])
	}

	return report, nil
}

func findingURL(repoURL string, finding scanreport.Finding) string {
	if repoURL == "" || finding.Commit == "" || finding.File == "" {
		return ""
	}

	base := strings.TrimSuffix(repoURL, ".git")
	line := strconv.Itoa(finding.Line)

	switch deriveProvider(repoURL) {
	case ProviderGitLab:
		return base + "/-/blob/" + finding.Commit + "/" + finding.File + "#L" + line
	case ProviderBitbucket:
		return base + "/src/" + finding.Commit + "/" + finding.File + "#lines-" + line
	default:
		return base + "/blob/" + finding.Commit + "/" + finding.File + "#L" + line
	}
}
