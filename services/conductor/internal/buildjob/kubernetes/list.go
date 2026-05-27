package kubernetes

import (
	"context"
	"net/url"
	"sort"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (c *Client) List(ctx context.Context, workspaceID, repoURL, contextPath string) ([]buildjob.Job, error) {
	parsedURL, err := url.Parse(repoURL)

	if err != nil {
		return nil, err
	}

	// TODO: Figure out if we can figure by "labelTriggerRef".
	selector := labels.SelectorFromSet(labels.Set{
		labelWorkspace:   workspaceID,
		labelRepoHash:    repoURLHash(*parsedURL),
		labelContextHash: contextHash(contextPath),
	})

	jobs, err := c.kubernetes.BatchV1().Jobs(c.namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	result := make([]buildjob.Job, 0, len(jobs.Items))

	for i := range jobs.Items {
		result = append(result, toJob(jobs.Items[i]))
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].StartedAt == nil {
			return false
		}

		if result[j].StartedAt == nil {
			return true
		}

		return result[i].StartedAt.After(*result[j].StartedAt)
	})

	return result, nil
}
