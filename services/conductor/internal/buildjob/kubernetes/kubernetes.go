package kubernetes

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/url"
	"path"
	"strings"

	"github.com/zeitlos/lucity/pkg/github"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

	"github.com/google/go-containerregistry/pkg/name"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	labelWorkspace    = "lucity.dev/workspace"
	labelRepoHash     = "lucity.dev/source-repo-hash"
	labelSourceCommit = "lucity.dev/source-commit"
	labelContextHash  = "lucity.dev/source-context-hash"
	labelRelease      = "lucity.dev/release"
)

const (
	annotationSourceRepo    = "lucity.dev/source-repo"
	annotationContext       = "lucity.dev/source-context"
	annotationTargets       = "lucity.dev/target-image-refs"
	annotationTriggeredBy   = "lucity.dev/triggered-by"
	annotationCancelledAt   = "lucity.dev/cancelled-at"
	annotationCommitMessage = "lucity.dev/commit-message"
)

type Client struct {
	namespace          string
	registry           string
	registryAuthSecret string
	buildKitAddr       string
	buildRunnerImage   string
	buildKitTLSSecret  string
	buildKitServerName string

	kubernetes kubernetes.Interface
	github     *github.App
}

func New(kubernetes kubernetes.Interface, namespace, registry, registryAuthSecret, buildRunnerImage, buildKitTLSSecret, buildKitServerName string) *Client {
	return &Client{
		namespace:          namespace,
		registry:           registry,
		registryAuthSecret: registryAuthSecret,
		buildKitAddr:       "tcp://lucity-buildkit:1234",
		buildRunnerImage:   buildRunnerImage,
		buildKitTLSSecret:  buildKitTLSSecret,
		buildKitServerName: buildKitServerName,
		kubernetes:         kubernetes,
	}
}

var _ buildjob.Interface = (*Client)(nil)

func toJob(job batch.Job) buildjob.Job {
	build := buildjob.Job{
		ID:            buildjob.BuildID{Workspace: job.Labels[labelWorkspace], Name: job.Name},
		Status:        buildStatus(job),
		SourceURL:     job.Annotations[annotationSourceRepo],
		Commit:        job.Labels[labelSourceCommit],
		CommitMessage: job.Annotations[annotationCommitMessage],
		ContextPath:   job.Annotations[annotationContext],
		TriggeredBy:   job.Annotations[annotationTriggeredBy],
		ReleaseID:     job.Labels[labelRelease],
		ImageRefs:     make(map[string]name.Reference),
	}

	if job.Status.StartTime != nil {
		build.StartedAt = new(job.Status.StartTime.Time)
	}

	if job.Status.CompletionTime != nil {
		build.FinishedAt = new(job.Status.CompletionTime.Time)
	}

	if refs := job.Annotations[annotationTargets]; refs != "" {
		for _, ref := range strings.Split(refs, ",") {
			parsed, err := name.ParseReference(ref)

			if err != nil {
				slog.Warn("failed to parse image ref on build job", "error", err, "ref", ref, "job", job.Name, "namespace", job.Namespace)
				continue
			}

			build.ImageRefs[parsed.Context().RepositoryStr()] = parsed
		}
	}

	return build
}

func buildStatus(job batch.Job) buildjob.Status {
	_, cancelled := job.Annotations[annotationCancelledAt]

	for _, c := range job.Status.Conditions {
		if c.Status != core.ConditionTrue {
			continue
		}

		switch c.Type {
		case batch.JobComplete:
			return buildjob.StatusSucceeded
		case batch.JobFailed:
			if cancelled {
				return buildjob.StatusCancelled
			}

			return buildjob.StatusFailed
		}
	}

	if cancelled {
		// Cancel issued, deadline patch in flight, controller hasn't transitioned yet
		return buildjob.StatusCancelling
	}

	if job.Status.Active > 0 {
		return buildjob.StatusRunning
	}

	return buildjob.StatusQueued
}

func normalizeRepoURL(u url.URL) string {
	normalized := u.Host + strings.TrimSuffix(u.Path, ".git")
	normalized = strings.ToLower(normalized)

	return strings.TrimSuffix(normalized, "/")
}

func normalizeContextPath(contextPath string) string {
	normalized := path.Clean(strings.TrimSpace(contextPath))

	if normalized == "." || normalized == "/" {
		return ""
	}

	return strings.TrimPrefix(normalized, "/")
}

func repoURLHash(u url.URL) string {
	hash := sha256.Sum256([]byte(normalizeRepoURL(u)))
	return hex.EncodeToString(hash[:8])
}

func contextHash(contextPath string) string {
	hash := sha256.Sum256([]byte(normalizeContextPath(contextPath)))
	return hex.EncodeToString(hash[:8])
}
