package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/conductor/internal/deployjob"
)

func (c *Client) Logs(ctx context.Context, id deployjob.DeployID) (io.ReadCloser, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	for {
		job, err := c.kubernetes.BatchV1().Jobs(c.config.Namespace).Get(waitCtx, id.Name, meta.GetOptions{})

		if err != nil {
			return nil, err
		}

		if job.Labels[labels.Workspace] != id.Workspace {
			return nil, fmt.Errorf("deploy %q not found", id.Name)
		}

		pods, err := c.kubernetes.CoreV1().Pods(c.config.Namespace).List(waitCtx, meta.ListOptions{
			LabelSelector: "job-name=" + id.Name,
		})

		if err != nil {
			return nil, err
		}

		if len(pods.Items) > 0 && len(pods.Items[0].Spec.Containers) > 0 {
			pod := pods.Items[0]
			container := pod.Spec.Containers[0]

			switch pod.Status.Phase {
			case core.PodRunning, core.PodSucceeded, core.PodFailed:
				req := c.kubernetes.CoreV1().Pods(c.config.Namespace).GetLogs(pod.Name, &core.PodLogOptions{
					Follow:    true,
					Container: container.Name,
				})

				return req.Stream(ctx)
			}
		}

		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("deploy logs for %q timed out waiting for pod", id.Name)
			}

			return nil, waitCtx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
