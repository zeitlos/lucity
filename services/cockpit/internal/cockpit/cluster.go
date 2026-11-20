package cockpit

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/platform"
)

const (
	annotationSourceCommit = "lucity.dev/source-commit"
	annotationSourceRepo   = "lucity.dev/source-repo"
)

func (c *Client) cluster(ctx context.Context) (map[string][]Environment, error) {
	namespaces, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{LabelSelector: labels.Workspace})

	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	byNamespace := map[string]*Environment{}

	for _, namespace := range namespaces.Items {
		byNamespace[namespace.Name] = &Environment{ID: platform.EnvironmentID{
			Workspace: namespace.Labels[labels.Workspace],
			Project:   namespace.Labels[labels.Project],
			Name:      namespace.Labels[labels.Environment],
		}}
	}

	deployments, err := c.kubernetes.AppsV1().Deployments("").List(ctx, meta.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}

	for _, deployment := range deployments.Items {
		environment := byNamespace[deployment.Namespace]

		if environment == nil {
			continue
		}

		service := Service{
			Name:       deployment.Labels[labels.Service],
			Commit:     deployment.Spec.Template.Annotations[annotationSourceCommit],
			Repository: deployment.Annotations[annotationSourceRepo],
		}

		if service.Name == "" {
			service.Name = deployment.Name
		}

		if containers := deployment.Spec.Template.Spec.Containers; len(containers) > 0 {
			service.Image = containers[0].Image
		}

		environment.Services = append(environment.Services, service)
	}

	pods, err := c.kubernetes.CoreV1().Pods("").List(ctx, meta.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	for _, pod := range pods.Items {
		environment := byNamespace[pod.Namespace]

		if environment == nil {
			continue
		}

		switch pod.Status.Phase {
		case core.PodRunning:
			environment.RunningPods++
		case core.PodPending:
			environment.PendingPods++
		default:
			continue
		}

		for _, container := range pod.Spec.Containers {
			environment.MemoryRequested += container.Resources.Requests.Memory().Value()
		}
	}

	result := map[string][]Environment{}

	for _, environment := range byNamespace {
		result[environment.ID.Workspace] = append(result[environment.ID.Workspace], *environment)
	}

	for _, environments := range result {
		sort.Slice(environments, func(i, j int) bool {
			return environments[i].ID.String() < environments[j].ID.String()
		})
	}

	return result, nil
}

func (c *Client) egress(ctx context.Context) (map[string]float64, error) {
	query := `sum by (namespace) (label_replace(increase(envoy_cluster_upstream_cx_rx_bytes_total{envoy_cluster_name=~"lucity-system/cilium-gateway-.+"}[7d]), "namespace", "$1", "envoy_cluster_name", "lucity-system/cilium-gateway-[^/]+/([a-z0-9-]+)_.+"))`

	samples, err := c.metrics.QueryVector(ctx, query, time.Now())

	if err != nil {
		return nil, fmt.Errorf("query egress: %w", err)
	}

	result := make(map[string]float64, len(samples))

	for _, sample := range samples {
		if namespace := sample.Labels["namespace"]; namespace != "" {
			result[namespace] = sample.Value
		}
	}

	return result, nil
}

// EnvironmentLogs returns the last lines of every running pod in the
// environment, each line prefixed with its pod's service.
func (c *Client) EnvironmentLogs(ctx context.Context, id platform.EnvironmentID, lines int64) ([]string, error) {
	pods, err := c.kubernetes.CoreV1().Pods(id.Namespace()).List(ctx, meta.ListOptions{FieldSelector: "status.phase=Running"})

	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	var result []string

	for _, pod := range pods.Items {
		service := pod.Labels[labels.Service]

		if service == "" {
			service = pod.Name
		}

		for _, container := range pod.Spec.Containers {
			raw, err := c.kubernetes.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &core.PodLogOptions{Container: container.Name, TailLines: &lines}).DoRaw(ctx)

			if err != nil {
				return nil, fmt.Errorf("logs of %s: %w", pod.Name, err)
			}

			for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
				if line != "" {
					result = append(result, service+"  "+line)
				}
			}
		}
	}

	return result, nil
}
