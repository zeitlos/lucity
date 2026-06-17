package kubernetes

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/zeitlos/lucity/services/conductor/internal/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Logs streams pod logs for every pod backing the service. Lines are
// prefixed with the pod's short ID when more than one pod is present. The
// channel closes once all pod streams end or ctx is cancelled.
func (c *Client) Logs(ctx context.Context, id platform.ServiceID, tail int) (<-chan platform.LogEntry, error) {
	namespace := id.Namespace()

	pods, err := c.kubernetes.CoreV1().Pods(namespace).List(ctx, meta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", serviceLabel, id.Name),
	})

	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for service %q in %q", id.Name, namespace)
	}

	multiplePods := len(pods.Items) > 1

	tailLines := int64(1000)

	if tail > 0 {
		tailLines = int64(tail)
	}

	out := make(chan platform.LogEntry, 128)

	var wg sync.WaitGroup

	for _, pod := range pods.Items {
		wg.Add(1)

		go func(podName string) {
			defer wg.Done()
			c.streamPodLogs(ctx, namespace, podName, id.Name, tailLines, multiplePods, out)
		}(pod.Name)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out, nil
}

func (c *Client) streamPodLogs(ctx context.Context, namespace, pod, container string, tail int64, prefix bool, out chan<- platform.LogEntry) {
	stream, err := c.kubernetes.CoreV1().Pods(namespace).GetLogs(pod, &core.PodLogOptions{
		Container: container,
		Follow:    true,
		TailLines: &tail,
	}).Stream(ctx)

	if err != nil {
		slog.Warn("failed to open log stream", "pod", pod, "error", err)
		return
	}

	defer stream.Close()

	short := shortPodID(pod)
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if prefix {
			line = fmt.Sprintf("[%s] %s", short, line)
		}

		select {
		case out <- platform.LogEntry{Pod: pod, Line: line}:
		case <-ctx.Done():
			return
		}
	}
}

func shortPodID(name string) string {
	parts := strings.Split(name, "-")

	if len(parts) == 0 {
		return name
	}

	return parts[len(parts)-1]
}
