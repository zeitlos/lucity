package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/services/conductor/internal/buildjob"
)

func (c *Client) Logs(ctx context.Context, id buildjob.BuildID) (io.ReadCloser, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	for {
		job, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(waitCtx, id.Name, meta.GetOptions{})

		if err != nil {
			return nil, err
		}

		if job.Labels[labelWorkspace] != id.Workspace {
			return nil, fmt.Errorf("build %q not found", id.Name)
		}

		pods, err := c.kubernetes.CoreV1().Pods(c.namespace).List(waitCtx, meta.ListOptions{
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
				req := c.kubernetes.CoreV1().Pods(c.namespace).GetLogs(pod.Name, &core.PodLogOptions{
					Follow:    true,
					Container: container.Name,
				})

				return req.Stream(ctx)
			}
		}

		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("build logs for %q timed out waiting for pod", id.Name)
			}

			return nil, waitCtx.Err()
		case <-time.After(500 * time.Millisecond):
			// Wait 500ms, then try again.
		}
	}
}
