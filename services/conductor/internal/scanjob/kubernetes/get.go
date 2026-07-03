package kubernetes

import (
	"context"
	"fmt"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/platform"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
)

func (c *Client) Get(ctx context.Context, id scanjob.ScanID) (*scanjob.Job, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).Get(ctx, id.Name, meta.GetOptions{})

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("scan %q not found", id.Name)
		}

		return nil, err
	}

	if job.Labels[labels.Workspace] != id.Workspace {
		return nil, fmt.Errorf("scan %q not found", id.Name)
	}

	return new(c.toJob(ctx, *job)), nil
}

func (c *Client) List(ctx context.Context, service platform.ServiceID) ([]scanjob.Job, error) {
	selector := k8slabels.SelectorFromSet(scanJobLabels(service, ""))

	jobs, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).List(ctx, meta.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return nil, err
	}

	result := make([]scanjob.Job, 0, len(jobs.Items))

	for i := range jobs.Items {
		result = append(result, c.toJob(ctx, jobs.Items[i]))
	}

	sort.Slice(result, func(i, j int) bool {
		if !result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].CreatedAt.After(result[j].CreatedAt)
		}

		return result[i].ID.Name < result[j].ID.Name
	})

	return result, nil
}
