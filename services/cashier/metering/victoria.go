package metering

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zeitlos/lucity/pkg/victoria"
)

// VMClient queries VictoriaMetrics via the Prometheus-compatible HTTP API.
type VMClient struct {
	client *victoria.Client
}

// NewVMClient connects to VictoriaMetrics at vmURL (e.g. http://lucity-infra-victoria-metrics-single-server:8428).
func NewVMClient(vmURL string) (*VMClient, error) {
	client, err := victoria.New(vmURL)
	if err != nil {
		return nil, err
	}

	// Sanity check: VM exposes vm_app_uptime_seconds for itself.
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping VictoriaMetrics: %w", err)
	}

	return &VMClient{client: client}, nil
}

// CPUByNamespace returns total CPU-seconds consumed per namespace over the given window.
// Uses container_cpu_time_seconds_total (counter) with PromQL increase(), which handles
// container restarts correctly.
func (c *VMClient) CPUByNamespace(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`sum by (k8s_namespace_name) (increase(container_cpu_time_seconds_total[%s]))`,
		promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "k8s_namespace_name")
}

// MemoryByNamespace returns total memory working_set bytes per namespace, computed as
// the sum across containers of each container's per-window average. This correctly
// reflects the bytes-resident the namespace consumed (multi-container pods sum, not avg).
func (c *VMClient) MemoryByNamespace(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`sum by (k8s_namespace_name) (avg_over_time(container_memory_working_set_bytes[%s]))`,
		promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "k8s_namespace_name")
}

// DiskByNamespace returns the persistent volume capacity (bytes) per namespace,
// averaged over the window. Sourced from kube-state-metrics (PVCs only) — emptyDir,
// configMap, secret, and projected volumes are correctly excluded.
func (c *VMClient) DiskByNamespace(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`sum by (namespace) (avg_over_time(kube_persistentvolumeclaim_resource_requests_storage_bytes[%s]))`,
		promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "namespace")
}

func (c *VMClient) EgressByNamespace(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`sum by (namespace) (label_replace(increase(envoy_cluster_upstream_cx_rx_bytes_total{envoy_cluster_name=~"lucity-system/cilium-gateway-.+"}[%s]), "namespace", "$1", "envoy_cluster_name", "lucity-system/cilium-gateway-[^/]+/([a-z0-9-]+)_.+"))`,
		promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "namespace")
}

func (c *VMClient) StorageByWorkspace(ctx context.Context, start, end time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`sum by (workspace) (avg_over_time(lucity_bucket_size_bytes[%s]))`,
		promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "workspace")
}

// queryByLabel runs an instant PromQL query at evalTime and returns a map keyed by
// the named label.
func (c *VMClient) queryByLabel(ctx context.Context, query string, evalTime time.Time, label string) (map[string]float64, error) {
	samples, err := c.client.QueryVector(ctx, query, evalTime)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(samples))
	for _, s := range samples {
		key, ok := s.Labels[label]
		if !ok {
			continue
		}
		out[key] = s.Value
	}
	return out, nil
}

// promDuration renders a Go duration as a Prometheus range vector duration.
// We use seconds for determinism (e.g. 1h → "3600s") since PromQL accepts integer-second strings.
func promDuration(d time.Duration) string {
	return strconv.FormatInt(int64(d.Seconds()), 10) + "s"
}
