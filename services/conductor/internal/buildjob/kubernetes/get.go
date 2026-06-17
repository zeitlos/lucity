package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"

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

	return new(toJob(*job)), nil
}
