package kubernetes

import (
	"context"
	"fmt"

	"github.com/zeitlos/lucity/pkg/platform"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Instances returns the pods of the service that exist right now.
func (c *Client) Instances(ctx context.Context, id platform.ServiceID) ([]platform.Instance, error) {
	pods, err := c.kubernetes.CoreV1().Pods(id.Namespace()).List(ctx, meta.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", serviceLabel, id.Name),
	})

	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	instances := make([]platform.Instance, 0, len(pods.Items))

	for _, pod := range pods.Items {
		instances = append(instances, platform.Instance{Name: pod.Name, State: instanceState(pod)})
	}

	return instances, nil
}

func instanceState(pod core.Pod) platform.InstanceState {
	if pod.DeletionTimestamp != nil {
		return platform.InstanceStopping
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == core.PodReady && condition.Status == core.ConditionTrue {
			return platform.InstanceRunning
		}
	}

	return platform.InstanceStarting
}
