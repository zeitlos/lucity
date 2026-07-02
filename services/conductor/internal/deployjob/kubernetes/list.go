package kubernetes

import (
	"context"
	"sort"

	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
)

func (c *Client) List(ctx context.Context, service platform.ServiceID) ([]deployjob.Job, error) {
	selector := k8slabels.SelectorFromSet(deployJobLabels(service, ""))

	jobs, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	result := make([]deployjob.Job, 0, len(jobs.Items))

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
