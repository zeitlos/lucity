package kubernetes

import (
	"context"
	"fmt"
	"io"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
	"github.com/zeitlos/lucity/services/conductor/internal/jobs"
)

func (c *Client) Logs(ctx context.Context, id buildjob.BuildID) (io.ReadCloser, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(ctx, id.Name, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	if job.Labels[labels.Workspace] != id.Workspace {
		return nil, fmt.Errorf("build %q not found", id.Name)
	}

	return jobs.StreamLogs(ctx, c.kubernetes, c.namespace, id.Name)
}
