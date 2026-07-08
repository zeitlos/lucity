package conductor

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"

	"github.com/zeitlos/lucity/pkg/auth"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/deployer"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
)

type ReleaseStatus string

const (
	ReleaseQueued     ReleaseStatus = "queued"
	ReleaseBuilding   ReleaseStatus = "building"
	ReleaseDeploying  ReleaseStatus = "deploying"
	ReleaseLive       ReleaseStatus = "live"
	ReleaseFailed     ReleaseStatus = "failed"
	ReleaseCancelled  ReleaseStatus = "cancelled"
	ReleaseSuperseded ReleaseStatus = "superseded"
)

type SourceProvider string

const (
	ProviderGitHub    SourceProvider = "github"
	ProviderGitLab    SourceProvider = "gitlab"
	ProviderBitbucket SourceProvider = "bitbucket"
)

type Commit struct {
	SHA     string
	Message string
	URL     string
}

type GitSource struct {
	Provider    SourceProvider
	Repository  string
	URL         string
	Ref         string
	ContextPath string
	Commit      Commit
}

type ReleaseTrigger struct {
	Kind  deployer.TriggerKind
	Actor string
}

type Release struct {
	ID         ReleaseID
	Status     ReleaseStatus
	Source     *GitSource
	Trigger    ReleaseTrigger
	Build      *Build
	Deploy     *Deploy
	Scan       *Scan
	Deployment *Deployment
	CreatedAt  time.Time
}

func (c *Client) Releases(ctx context.Context, service platform.Service) ([]Release, error) {
	serviceID := service.ID
	deployments := service.Deployments

	var builds []Build

	if service.SourceURL != "" {
		var err error
		builds, err = c.buildjob.List(ctx, serviceID)

		if err != nil {
			return nil, err
		}
	}

	deploys, err := c.deployjob.List(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	scans, err := c.scanjob.List(ctx, serviceID)

	if err != nil {
		return nil, err
	}

	type group struct {
		build      *Build
		deploy     *Deploy
		scan       *Scan
		deployment *Deployment
	}

	groups := map[string]*group{}

	groupFor := func(key string) *group {
		if groups[key] == nil {
			groups[key] = &group{}
		}

		return groups[key]
	}

	for i := range deployments {
		deployment := &deployments[i]
		key := deployment.ReleaseID

		if key == "" {
			key = "dep-" + deployment.ID.Hash
		}

		g := groupFor(key)

		if g.deployment == nil || deployment.CreatedAt.After(g.deployment.CreatedAt) {
			g.deployment = deployment
		}
	}

	for i := range builds {
		build := &builds[i]
		key := build.ReleaseID

		if key == "" {
			key = build.ID.Name
		}

		groupFor(key).build = build
	}

	for i := range deploys {
		deploy := &deploys[i]
		key := deploy.ReleaseID

		if key == "" {
			key = deploy.BuildName
		}

		groupFor(key).deploy = deploy
	}

	for i := range scans {
		scan := &scans[i]
		key := scan.ReleaseID

		if key == "" {
			key = scan.BuildName
		}

		g := groupFor(key)

		if g.scan == nil || scan.CreatedAt.After(g.scan.CreatedAt) {
			g.scan = scan
		}
	}

	releases := make([]Release, 0, len(groups))

	for key, g := range groups {
		releases = append(releases, assembleRelease(serviceID.Workspace, key, g.build, g.deploy, g.scan, g.deployment))
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].CreatedAt.After(releases[j].CreatedAt)
	})

	return releases, nil
}

func assembleRelease(workspace, name string, build *Build, deploy *Deploy, scan *Scan, deployment *Deployment) Release {
	return Release{
		ID:         ReleaseID{Workspace: workspace, Name: name},
		Status:     releaseStatus(build, deploy, deployment),
		Source:     releaseSource(build, deployment),
		Trigger:    releaseTrigger(deployment),
		Build:      build,
		Deploy:     deploy,
		Scan:       scan,
		Deployment: deployment,
		CreatedAt:  releaseCreatedAt(build, deploy, deployment),
	}
}

func releaseStatus(build *Build, deploy *Deploy, deployment *Deployment) ReleaseStatus {
	if deployment != nil {
		switch deployment.Status {
		case platform.DeploymentDeploying:
			return ReleaseDeploying
		case platform.DeploymentActive:
			return ReleaseLive
		case platform.DeploymentSuperseded:
			return ReleaseSuperseded
		case platform.DeploymentFailed:
			return ReleaseFailed
		}
	}

	if deploy != nil && deploy.Status == deployjob.StatusFailed && (build == nil || build.Status == buildjob.StatusSucceeded) {
		return ReleaseFailed
	}

	deployInFlight := deploy != nil && (deploy.Status == deployjob.StatusQueued || deploy.Status == deployjob.StatusRunning)

	if build != nil {
		switch build.Status {
		case buildjob.StatusQueued:
			return ReleaseQueued
		case buildjob.StatusRunning, buildjob.StatusCancelling:
			return ReleaseBuilding
		case buildjob.StatusSucceeded:
			if deployInFlight {
				return ReleaseDeploying
			}

			return ReleaseSuperseded
		case buildjob.StatusFailed:
			return ReleaseFailed
		case buildjob.StatusCancelled:
			return ReleaseCancelled
		}
	}

	if deployInFlight {
		return ReleaseDeploying
	}

	return ReleaseSuperseded
}

func releaseSource(build *Build, deployment *Deployment) *GitSource {
	if deployment != nil && deployment.SourceURL != "" {
		return gitSource(deployment.SourceURL, deployment.Ref, deployment.ContextPath, Commit{
			SHA:     deployment.Commit,
			Message: deployment.CommitMessage,
		})
	}

	if build != nil && build.SourceURL != "" {
		return gitSource(build.SourceURL, "", build.ContextPath, Commit{SHA: build.Commit, Message: build.CommitMessage})
	}

	return nil
}

func releaseTrigger(deployment *Deployment) ReleaseTrigger {
	if deployment != nil && deployment.ReleaseTrigger != "" {
		return ReleaseTrigger{
			Kind:  deployer.TriggerKind(deployment.ReleaseTrigger),
			Actor: deployment.ReleaseActor,
		}
	}

	return ReleaseTrigger{Kind: deployer.TriggerManual}
}

func releaseCreatedAt(build *Build, deploy *Deploy, deployment *Deployment) time.Time {
	if build != nil {
		if build.StartedAt != nil {
			return *build.StartedAt
		}

		if !build.CreatedAt.IsZero() {
			return build.CreatedAt
		}
	}

	if deployment != nil {
		return deployment.CreatedAt
	}

	if deploy != nil && deploy.StartedAt != nil {
		return *deploy.StartedAt
	}

	return time.Time{}
}

func gitSource(repoURL, ref, contextPath string, commit Commit) *GitSource {
	if repoURL == "" {
		return nil
	}

	provider := deriveProvider(repoURL)
	commit.URL = commitURL(provider, repoURL, commit.SHA)

	return &GitSource{
		Provider:    provider,
		Repository:  repositoryName(repoURL),
		URL:         strings.TrimSuffix(repoURL, ".git"),
		Ref:         ref,
		ContextPath: contextPath,
		Commit:      commit,
	}
}

func deriveProvider(repoURL string) SourceProvider {
	parsed, err := url.Parse(repoURL)

	if err != nil {
		return ProviderGitHub
	}

	host := strings.ToLower(parsed.Host)

	switch {
	case strings.Contains(host, "gitlab"):
		return ProviderGitLab
	case strings.Contains(host, "bitbucket"):
		return ProviderBitbucket
	default:
		return ProviderGitHub
	}
}

func repositoryName(repoURL string) string {
	parsed, err := url.Parse(repoURL)

	if err != nil {
		return ""
	}

	return strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), ".git")
}

func commitURL(provider SourceProvider, repoURL, sha string) string {
	if repoURL == "" || sha == "" {
		return ""
	}

	base := strings.TrimSuffix(repoURL, ".git")

	if provider == ProviderBitbucket {
		return base + "/commits/" + sha
	}

	return base + "/commit/" + sha
}

func actorFromClaims(claims *auth.Claims) string {
	if claims == nil {
		return ""
	}

	if claims.Name != "" {
		return claims.Name
	}

	if claims.Email != "" {
		return claims.Email
	}

	return claims.Subject
}

type ReleaseID struct {
	Workspace string
	Name      string
}

func ParseReleaseID(s string) (ReleaseID, error) {
	workspace, name, ok := strings.Cut(s, "/")

	if !ok || workspace == "" || name == "" {
		return ReleaseID{}, fmt.Errorf("invalid release id %q", s)
	}

	return ReleaseID{Workspace: workspace, Name: name}, nil
}

func (r ReleaseID) String() string {
	return r.Workspace + "/" + r.Name
}

func (r ReleaseID) WorkspaceID() string {
	return r.Workspace
}

func (r *ReleaseID) UnmarshalGQL(val interface{}) error {
	str, ok := val.(string)

	if !ok {
		return fmt.Errorf("ReleaseID must be a string")
	}

	parsed, err := ParseReleaseID(str)

	if err != nil {
		return err
	}

	*r = parsed

	return nil
}

func (r ReleaseID) MarshalGQL(w io.Writer) {
	graphql.MarshalString(r.String()).MarshalGQL(w)
}

var (
	_ platform.WorkspaceScoped = ReleaseID{}
	_ graphql.Marshaler        = ReleaseID{}
	_ graphql.Unmarshaler      = (*ReleaseID)(nil)
)
