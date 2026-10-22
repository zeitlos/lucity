package kubernetes

import (
	"context"
	"sort"

	"github.com/zeitlos/lucity/pkg/jobs"
	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"

	k8slabels "k8s.io/apimachinery/pkg/labels"
)

func (c *Client) List(ctx context.Context, service platform.ServiceID) ([]deployjob.Job, error) {
	selector := k8slabels.SelectorFromSet(deployJobLabels(service, ""))

	cached, err := c.jobs.Deploys(ctx, selector)

	if err != nil {
		return nil, err
	}

	jobs.SortNewestFirst(cached)

	result := make([]deployjob.Job, 0, len(cached))

	for _, job := range cached {
		result = append(result, toJob(*job))
	}

	sort.SliceStable(result, func(i, j int) bool {
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
