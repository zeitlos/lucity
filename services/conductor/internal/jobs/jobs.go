package jobs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StreamLogs waits for a Job's pod to start and streams its container logs.
func StreamLogs(ctx context.Context, client kubernetes.Interface, namespace, jobName string) (io.ReadCloser, error) {
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	for {
		pods, err := client.CoreV1().Pods(namespace).List(waitCtx, meta.ListOptions{
			LabelSelector: "job-name=" + jobName,
		})

		if err != nil {
			return nil, err
		}

		if len(pods.Items) > 0 && len(pods.Items[0].Spec.Containers) > 0 {
			pod := pods.Items[0]
			container := pod.Spec.Containers[0]

			switch pod.Status.Phase {
			case core.PodRunning, core.PodSucceeded, core.PodFailed:
				req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &core.PodLogOptions{
					Follow:    true,
					Container: container.Name,
				})

				return req.Stream(ctx)
			}
		}

		select {
		case <-waitCtx.Done():
			if errors.Is(waitCtx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("logs for job %q timed out waiting for pod", jobName)
			}

			return nil, waitCtx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

// TerminationMessage returns the first non-empty container termination
// message of the Job's pods, or "" if none is available yet.
func TerminationMessage(ctx context.Context, client kubernetes.Interface, namespace, jobName string) (string, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})

	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		for _, status := range pod.Status.ContainerStatuses {
			if status.State.Terminated != nil && status.State.Terminated.Message != "" {
				return status.State.Terminated.Message, nil
			}
		}
	}

	return "", nil
}
