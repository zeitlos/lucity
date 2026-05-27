package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	for {
		if _, err := c.kubernetes.BatchV1().Jobs(c.namespace).Get(waitCtx, id, meta.GetOptions{}); err != nil {
			return nil, err
		}

		pods, err := c.kubernetes.CoreV1().Pods(c.namespace).List(waitCtx, meta.ListOptions{
			LabelSelector: "job-name=" + id,
		})

		if err != nil {
			return nil, err
		}

		if len(pods.Items) > 0 {
			pod := pods.Items[0]

			switch pod.Status.Phase {
			case core.PodRunning, core.PodSucceeded, core.PodFailed:
				req := c.kubernetes.CoreV1().Pods(c.namespace).GetLogs(pod.Name, &core.PodLogOptions{
					Follow:    true,
					Container: "builder",
				})

				return req.Stream(ctx)
			}
		}

		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("build logs for %q timed out waiting for pod", id)
			}

			return nil, waitCtx.Err()
		case <-time.After(500 * time.Millisecond):
			// Wait 500ms, then try again.
		}
	}
}
