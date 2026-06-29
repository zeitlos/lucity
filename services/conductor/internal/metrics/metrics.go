package metrics

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zeitlos/lucity/pkg/victoria"
)

const (
	podVolumePrefix         = "volume-"
	serviceDeploymentPrefix = "lucity-app-"
)

type Kind string

const (
	KindStorageUsed Kind = "storage_used"
	KindCPUUsage    Kind = "cpu_usage"
	KindMemoryUsage Kind = "memory_usage"
)

type Window string

const (
	Window1h  Window = "1h"
	Window6h  Window = "6h"
	Window24h Window = "24h"
	Window7d  Window = "7d"
	Window30d Window = "30d"
)

type windowParams struct {
	duration time.Duration
	step     time.Duration
}

var windows = map[Window]windowParams{
	Window1h:  {time.Hour, time.Minute},
	Window6h:  {6 * time.Hour, 5 * time.Minute},
	Window24h: {24 * time.Hour, 15 * time.Minute},
	Window7d:  {7 * 24 * time.Hour, time.Hour},
	Window30d: {30 * 24 * time.Hour, 6 * time.Hour},
}

type Point struct {
	Timestamp time.Time
	Value     *float64
}

type Series struct {
	Kind    Kind
	Replica string
	Points  []Point
}

type Provider struct {
	vm *victoria.Client
}

func New(vmURL string) (*Provider, error) {
	vm, err := victoria.New(vmURL)
	if err != nil {
		return nil, err
	}
	return &Provider{vm: vm}, nil
}

func (p *Provider) VolumeStorageUsed(ctx context.Context, namespace, volumeName string, window Window) ([]Series, error) {
	params, ok := windows[window]
	if !ok {
		return nil, fmt.Errorf("unknown window %q", window)
	}

	end := time.Now().UTC()
	start := end.Add(-params.duration)

	selector := fmt.Sprintf(`k8s_namespace_name=%q,k8s_volume_name=%q`, namespace, podVolumePrefix+volumeName)
	query := fmt.Sprintf(
		`max by (k8s_volume_name) (k8s_volume_capacity_bytes{%[1]s}) - max by (k8s_volume_name) (k8s_volume_available_bytes{%[1]s})`,
		selector,
	)

	samples, err := p.vm.QueryRange(ctx, query, start, end, params.step)
	if err != nil {
		return nil, err
	}

	if len(samples) == 0 {
		return nil, nil
	}

	return []Series{{Kind: KindStorageUsed, Points: toPoints(samples[0].Points)}}, nil
}

func (p *Provider) ServiceUsage(ctx context.Context, namespace, service string, kinds []Kind, window Window, perReplica bool) ([]Series, error) {
	params, ok := windows[window]
	if !ok {
		return nil, fmt.Errorf("unknown window %q", window)
	}

	end := time.Now().UTC()
	start := end.Add(-params.duration)

	selector := fmt.Sprintf(`k8s_namespace_name=%q,k8s_deployment_name=%q`, namespace, serviceDeploymentPrefix+service)

	var result []Series
	for _, kind := range kinds {
		query := serviceQuery(kind, selector, params.step, perReplica)
		if query == "" {
			continue
		}

		samples, err := p.vm.QueryRange(ctx, query, start, end, params.step)
		if err != nil {
			return nil, err
		}

		for _, s := range samples {
			result = append(result, Series{
				Kind:    kind,
				Replica: s.Labels["k8s_pod_name"],
				Points:  toPoints(s.Points),
			})
		}
	}

	return result, nil
}

func serviceQuery(kind Kind, selector string, step time.Duration, perReplica bool) string {
	grouping := ""
	if perReplica {
		grouping = " by (k8s_pod_name)"
	}

	switch kind {
	case KindCPUUsage:
		return fmt.Sprintf(`sum%s(rate(container_cpu_time_seconds_total{%s}[%s]))`, grouping, selector, promDuration(step))
	case KindMemoryUsage:
		return fmt.Sprintf(`sum%s(container_memory_working_set_bytes{%s})`, grouping, selector)
	default:
		return ""
	}
}

func promDuration(d time.Duration) string {
	return strconv.FormatInt(int64(d.Seconds()), 10) + "s"
}

func toPoints(pts []victoria.Point) []Point {
	out := make([]Point, 0, len(pts))
	for _, pt := range pts {
		value := pt.Value
		out = append(out, Point{Timestamp: pt.Time, Value: &value})
	}
	return out
}
