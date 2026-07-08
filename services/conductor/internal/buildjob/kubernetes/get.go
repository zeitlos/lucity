package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/jobs"

	"github.com/google/go-containerregistry/pkg/name"
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

	if job.Labels[labels.Workspace] != id.Workspace {
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
	message, err := jobs.TerminationMessage(ctx, c.kubernetes, c.namespace, jobName)

	if err != nil {
		return nil, err
	}

	if message == "" {
		return nil, nil
	}

	var result buildResult

	if err := json.Unmarshal([]byte(message), &result); err != nil {
		return nil, nil
	}

	return parseDigestRefs(result.Images), nil
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
