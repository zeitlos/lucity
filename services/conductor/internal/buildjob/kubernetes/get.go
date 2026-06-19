package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

	"github.com/google/go-containerregistry/pkg/name"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) Get(ctx context.Context, id buildjob.BuildID) (*buildjob.Job, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(ctx, id.Name, meta.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("build %q not found", id.Name)
		}

		return nil, err
	}

	if job.Labels[labelWorkspace] != id.Workspace {
		return nil, fmt.Errorf("build %q not found", id.Name)
	}

	result := toJob(*job)

	digests, err := c.builtDigests(ctx, id.Name)

	if err != nil {
		slog.WarnContext(ctx, "failed to read build digests", "error", err, "job", id.Name)
	} else {
		result.Digests = digests
	}

	return new(result), nil
}

type buildResult struct {
	Images []string `json:"images"`
}

func (c *Client) builtDigests(ctx context.Context, jobName string) (map[string]string, error) {
	pods, err := c.kubernetes.CoreV1().Pods(c.namespace).List(ctx, meta.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})

	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			message := terminatedMessage(status)

			if message == "" {
				continue
			}

			var result buildResult

			if err := json.Unmarshal([]byte(message), &result); err != nil {
				continue
			}

			return parseDigestRefs(result.Images), nil
		}
	}

	return nil, nil
}

func terminatedMessage(status core.ContainerStatus) string {
	if status.State.Terminated != nil {
		return status.State.Terminated.Message
	}

	return ""
}

func parseDigestRefs(refs []string) map[string]string {
	digests := make(map[string]string, len(refs))

	for _, ref := range refs {
		parsed, err := name.NewDigest(ref)

		if err != nil {
			slog.Warn("failed to parse built digest ref", "error", err, "ref", ref)
			continue
		}

		digests[parsed.Context().RepositoryStr()] = parsed.DigestStr()
	}

	return digests
}
