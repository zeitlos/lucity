package kubernetes

import (
	"context"
	"fmt"
	"io"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/jobs"
	"github.com/zeitlos/lucity/services/conductor/internal/scanjob"
)

func (c *Client) Logs(ctx context.Context, id scanjob.ScanID) (io.ReadCloser, error) {
	job, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).Get(ctx, id.Name, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	if job.Labels[labels.Workspace] != id.Workspace {
		return nil, fmt.Errorf("scan %q not found", id.Name)
	}

	return jobs.StreamLogs(ctx, c.kubernetes, c.config.Namespace, id.Name)
}
