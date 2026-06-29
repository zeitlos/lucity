package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/zeitlos/lucity/pkg/victoria"
)

const podVolumePrefix = "volume-"

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
	Points []Point
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

func (p *Provider) VolumeStorageUsed(ctx context.Context, namespace, volumeName string, window Window) (Series, error) {
	params, ok := windows[window]
	if !ok {
		return Series{}, fmt.Errorf("unknown window %q", window)
	}

	end := time.Now().UTC()
	start := end.Add(-params.duration)

	selector := fmt.Sprintf(`k8s_namespace_name=%q,k8s_volume_name=%q`, namespace, podVolumePrefix+volumeName)
	query := fmt.Sprintf(
		`max by (k8s_volume_name) (k8s_volume_capacity_bytes{%[1]s}) - max by (k8s_volume_name) (k8s_volume_available_bytes{%[1]s})`,
		selector,
	)

	series, err := p.vm.QueryRange(ctx, query, start, end, params.step)
	if err != nil {
		return Series{}, err
	}

	if len(series) == 0 {
		return Series{}, nil
	}

	points := make([]Point, 0, len(series[0].Points))
	for _, pt := range series[0].Points {
		value := pt.Value
		points = append(points, Point{Timestamp: pt.Time, Value: &value})
	}

	return Series{Points: points}, nil
}
