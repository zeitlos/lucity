package cockpit

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/victoria"
)

// LimitGroup holds the limits of one provider or resource.
type LimitGroup struct {
	Name   string
	Limits []Limit
}

type LimitUnit string

const (
	LimitUnitCount LimitUnit = "count"
	LimitUnitBytes LimitUnit = "bytes"
)

// Limit is how much of something is in use. Cap is nil when the provider's
// cap is unknown. History holds the values of the last 30 days, oldest first.
type Limit struct {
	Name    string
	Unit    LimitUnit
	Value   float64
	Cap     *float64
	History []victoria.Point
}

const (
	limitHistory = 30 * 24 * time.Hour
	limitStep    = 6 * time.Hour

	// pullZoneCap is the number of pull zones the edge CDN allows per account.
	pullZoneCap = 500

	// platformDisks is how many of the fullest platform disks are shown.
	platformDisks = 3

	hetznerVolumes  = `max by (persistentvolume) (last_over_time(kube_persistentvolume_info{csi_driver="csi.hetzner.cloud"}[10m]))`
	attachedVolumes = `label_replace(max by (volumename) (last_over_time(kube_volumeattachment_spec_source_persistentvolume[10m])), "persistentvolume", "$1", "volumename", "(.+)")`
)

// Limits returns the provider limits and the fullest platform disks.
func (c *Client) Limits(ctx context.Context) ([]LimitGroup, error) {
	limits := []struct {
		group string
		limit Limit
		query string
	}{
		{"Hetzner", Limit{Name: "Volumes", Unit: LimitUnitCount}, fmt.Sprintf(`count(%s) or vector(0)`, hetznerVolumes)},
		{"Hetzner", Limit{Name: "Volume space", Unit: LimitUnitBytes}, fmt.Sprintf(`sum(max by (persistentvolume) (last_over_time(kube_persistentvolume_capacity_bytes[10m])) * on (persistentvolume) %s) or vector(0)`, hetznerVolumes)},
		{"Hetzner", Limit{Name: "Detached volumes", Unit: LimitUnitCount}, fmt.Sprintf(`count(%s unless on (persistentvolume) %s) or vector(0)`, hetznerVolumes, attachedVolumes)},
		{"Bunny", Limit{Name: "Pull zones", Unit: LimitUnitCount, Cap: capOf(pullZoneCap)}, `max(lucity_edge_zones)`},
	}

	var disks []Limit

	group, ctx := errgroup.WithContext(ctx)

	for i := range limits {
		group.Go(func() error {
			value, err := c.value(ctx, strings.ToLower(limits[i].limit.Name), limits[i].query)
			limits[i].limit.Value = value

			return err
		})

		group.Go(func() error {
			history, err := c.history(ctx, limits[i].limit.Name, limits[i].query)
			limits[i].limit.History = history

			return err
		})
	}

	group.Go(func() error {
		var err error
		disks, err = c.platformDisks(ctx)

		return err
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	var result []LimitGroup

	for _, limit := range limits {
		if len(result) == 0 || result[len(result)-1].Name != limit.group {
			result = append(result, LimitGroup{Name: limit.group})
		}

		result[len(result)-1].Limits = append(result[len(result)-1].Limits, limit.limit)
	}

	return append(result, LimitGroup{Name: "Platform disks", Limits: disks}), nil
}

type disk struct {
	name      string
	selector  string
	capacity  float64
	available float64
}

func (d *disk) used() float64 {
	return d.capacity - d.available
}

// platformDisks returns the fullest volumes claimed outside the workspaces.
func (c *Client) platformDisks(ctx context.Context) ([]Limit, error) {
	namespaces, err := c.kubernetes.CoreV1().Namespaces().List(ctx, meta.ListOptions{LabelSelector: "!" + labels.Workspace})

	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	platform := map[string]bool{}
	var names []string

	for _, namespace := range namespaces.Items {
		platform[namespace.Name] = true
		names = append(names, namespace.Name)
	}

	pods, err := c.kubernetes.CoreV1().Pods("").List(ctx, meta.ListOptions{FieldSelector: "status.phase=Running"})

	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	disks := map[string]*disk{}

	for _, pod := range pods.Items {
		if !platform[pod.Namespace] {
			continue
		}

		claims := claimedVolumes(pod)

		for _, volume := range claims {
			workload, selector := podWorkload(pod)
			name := strings.TrimPrefix(strings.TrimPrefix(workload, "lucity-infra-"), "lucity-")

			if len(claims) > 1 {
				name += " " + volume
			}

			disks[pod.Namespace+"/"+pod.Name+"/"+volume] = &disk{
				name:     name,
				selector: fmt.Sprintf(`k8s_namespace_name=%q, %s, k8s_volume_name=%q`, pod.Namespace, selector, volume),
			}
		}
	}

	if len(disks) == 0 {
		return []Limit{}, nil
	}

	filter := fmt.Sprintf(`{k8s_namespace_name=~%q}`, strings.Join(names, "|"))

	for _, metric := range []string{"k8s_volume_capacity_bytes", "k8s_volume_available_bytes"} {
		samples, err := c.metrics.QueryVector(ctx, metric+filter, time.Now())

		if err != nil {
			return nil, fmt.Errorf("query platform disks: %w", err)
		}

		for _, sample := range samples {
			target := disks[sample.Labels["k8s_namespace_name"]+"/"+sample.Labels["k8s_pod_name"]+"/"+sample.Labels["k8s_volume_name"]]

			if target == nil {
				continue
			}

			if metric == "k8s_volume_capacity_bytes" {
				target.capacity = sample.Value
			} else {
				target.available = sample.Value
			}
		}
	}

	var measured []*disk

	for _, d := range disks {
		if d.capacity > 0 {
			measured = append(measured, d)
		}
	}

	sort.Slice(measured, func(i, j int) bool {
		return measured[i].used()/measured[i].capacity > measured[j].used()/measured[j].capacity
	})

	measured = measured[:min(len(measured), platformDisks)]
	result := make([]Limit, len(measured))

	group, ctx := errgroup.WithContext(ctx)

	for i, d := range measured {
		result[i] = Limit{Name: d.name, Unit: LimitUnitBytes, Value: d.used(), Cap: capOf(d.capacity)}

		group.Go(func() error {
			history, err := c.history(ctx, d.name, fmt.Sprintf(`max(k8s_volume_capacity_bytes{%[1]s} - k8s_volume_available_bytes{%[1]s})`, d.selector))
			result[i].History = history

			return err
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func claimedVolumes(pod core.Pod) []string {
	var names []string

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			names = append(names, volume.Name)
		}
	}

	return names
}

// podWorkload names what runs the pod, a deployment or the pod itself, with
// the metrics selector that follows it across restarts.
func podWorkload(pod core.Pod) (string, string) {
	if owner := meta.GetControllerOf(&pod); owner != nil && owner.Kind == "ReplicaSet" {
		if cut := strings.LastIndex(owner.Name, "-"); cut > 0 {
			deployment := owner.Name[:cut]

			return deployment, fmt.Sprintf(`k8s_deployment_name=%q`, deployment)
		}
	}

	return pod.Name, fmt.Sprintf(`k8s_pod_name=%q`, pod.Name)
}

func (c *Client) history(ctx context.Context, name, query string) ([]victoria.Point, error) {
	end := time.Now()
	series, err := c.metrics.QueryRange(ctx, query, end.Add(-limitHistory), end, limitStep)

	if err != nil {
		return nil, fmt.Errorf("query %s over 30 days: %w", strings.ToLower(name), err)
	}

	if len(series) == 0 {
		return []victoria.Point{}, nil
	}

	return series[0].Points, nil
}

func capOf(value float64) *float64 {
	return &value
}
