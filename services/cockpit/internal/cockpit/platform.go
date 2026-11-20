package cockpit

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
)

// Totals counts what runs on the cluster now. Each trend holds the hourly
// values of the last 24 hours, oldest first.
type Totals struct {
	RunningPods            int
	PendingPods            int
	FailedPods             int
	RunningContainers      int
	RunningWorkspaces      int
	Workspaces             int
	Releases               int
	FailedReleases         int
	FiringAlerts           []string
	RunningPodsTrend       []float64
	RunningContainersTrend []float64
	RunningWorkspacesTrend []float64
	ReleasesTrend          []float64
	FiringAlertsTrend      []float64
}

func (c *Client) Totals(ctx context.Context) (Totals, error) {
	var totals Totals

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return c.countPods(ctx, &totals)
	})

	group.Go(func() error {
		workspaces, err := c.value(ctx, "workspaces", `sum(max by (plan, standing) (lucity_workspaces))`)
		totals.Workspaces = int(math.Round(workspaces))

		return err
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, `max by (status) (lucity_helm_releases{chart="app"})`, time.Now())

		if err != nil {
			return fmt.Errorf("query releases: %w", err)
		}

		for _, sample := range samples {
			totals.Releases += int(math.Round(sample.Value))

			if sample.Labels["status"] == "failed" {
				totals.FailedReleases += int(math.Round(sample.Value))
			}
		}

		return nil
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, `ALERTS{alertstate="firing"}`, time.Now())

		if err != nil {
			return fmt.Errorf("query alerts: %w", err)
		}

		seen := map[string]bool{}
		totals.FiringAlerts = []string{}

		for _, sample := range samples {
			if name := sample.Labels["alertname"]; !seen[name] {
				seen[name] = true
				totals.FiringAlerts = append(totals.FiringAlerts, name)
			}
		}

		sort.Strings(totals.FiringAlerts)

		return nil
	})

	trends := []struct {
		name   string
		query  string
		target *[]float64
	}{
		{"running pods", `sum(last_over_time(kube_pod_status_phase{phase="Running"}[2m]))`, &totals.RunningPodsTrend},
		{"running containers", `count(last_over_time(container_memory_working_set_bytes[2m]))`, &totals.RunningContainersTrend},
		{"running workspaces", `count(count by (workspace) (last_over_time(k8s_pod_memory_working_set_bytes{workspace!=""}[2m])))`, &totals.RunningWorkspacesTrend},
		{"releases", `sum(max by (status) (lucity_helm_releases{chart="app"}))`, &totals.ReleasesTrend},
		{"firing alerts", `count(ALERTS{alertstate="firing"}) or vector(0)`, &totals.FiringAlertsTrend},
	}

	for _, trend := range trends {
		group.Go(func() error {
			values, err := c.trend(ctx, trend.name, trend.query)
			*trend.target = values

			return err
		})
	}

	if err := group.Wait(); err != nil {
		return Totals{}, err
	}

	return totals, nil
}

func (c *Client) countPods(ctx context.Context, totals *Totals) error {
	namespaces, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{LabelSelector: labels.Workspace})

	if err != nil {
		return fmt.Errorf("list namespaces: %w", err)
	}

	workspaceOf := map[string]string{}

	for _, namespace := range namespaces.Items {
		workspaceOf[namespace.Name] = namespace.Labels[labels.Workspace]
	}

	pods, err := c.kubernetes.CoreV1().Pods("").List(ctx, meta.ListOptions{})

	if err != nil {
		return fmt.Errorf("list pods: %w", err)
	}

	running := map[string]bool{}

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case core.PodRunning:
			totals.RunningPods++

			if workspace := workspaceOf[pod.Namespace]; workspace != "" {
				running[workspace] = true
			}
		case core.PodPending:
			totals.PendingPods++
		case core.PodFailed:
			if owner := meta.GetControllerOf(&pod); owner == nil || owner.Kind != "Job" {
				totals.FailedPods++
			}
		}

		for _, status := range pod.Status.ContainerStatuses {
			if status.State.Running != nil {
				totals.RunningContainers++
			}
		}
	}

	totals.RunningWorkspaces = len(running)

	return nil
}

func (c *Client) value(ctx context.Context, name, query string) (float64, error) {
	samples, err := c.metrics.QueryVector(ctx, query, time.Now())

	if err != nil {
		return 0, fmt.Errorf("query %s: %w", name, err)
	}

	if len(samples) == 0 {
		return 0, nil
	}

	return samples[0].Value, nil
}

func (c *Client) trend(ctx context.Context, name, query string) ([]float64, error) {
	end := time.Now()
	series, err := c.metrics.QueryRange(ctx, query, end.Add(-23*time.Hour), end, time.Hour)

	if err != nil {
		return nil, fmt.Errorf("query %s over the last day: %w", name, err)
	}

	values := []float64{}

	if len(series) > 0 {
		for _, point := range series[0].Points {
			values = append(values, point.Value)
		}
	}

	return values, nil
}
