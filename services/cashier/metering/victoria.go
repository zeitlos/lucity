package metering

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// VMClient queries VictoriaMetrics via the Prometheus-compatible HTTP API.
type VMClient struct {
	url    string
	client *http.Client
}

// NewVMClient connects to VictoriaMetrics at vmURL (e.g. http://lucity-infra-victoria-metrics-single-server:8428).
func NewVMClient(vmURL string) (*VMClient, error) {
	if vmURL == "" {
		return nil, fmt.Errorf("VictoriaMetrics URL is empty")
	}

	c := &VMClient{
		url:    strings.TrimRight(vmURL, "/"),
		client: &http.Client{Timeout: 30 * time.Second},
	}

	// Sanity check: VM exposes vm_app_uptime_seconds for itself.
	if _, err := c.queryVector(context.Background(), "vm_app_uptime_seconds", time.Now()); err != nil {
		return nil, fmt.Errorf("failed to ping VictoriaMetrics: %w", err)
	}
	return c, nil
}

// CPUByNamespace returns total CPU-seconds consumed per namespace over the given window.
// Uses container_cpu_time_seconds_total (counter) with PromQL increase(), which handles
// container restarts correctly.
func (c *VMClient) CPUByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}
	q := fmt.Sprintf(
		`sum by (k8s_namespace_name) (increase(container_cpu_time_seconds_total{k8s_namespace_name=~"%s"}[%s]))`,
		nsRegex(namespaces), promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "k8s_namespace_name")
}

// MemoryByNamespace returns total memory working_set bytes per namespace, computed as
// the sum across containers of each container's per-window average. This correctly
// reflects the bytes-resident the namespace consumed (multi-container pods sum, not avg).
func (c *VMClient) MemoryByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}
	q := fmt.Sprintf(
		`sum by (k8s_namespace_name) (avg_over_time(container_memory_working_set_bytes{k8s_namespace_name=~"%s"}[%s]))`,
		nsRegex(namespaces), promDuration(end.Sub(start)),
	)
	return c.queryByLabel(ctx, q, end, "k8s_namespace_name")
}

// DiskByNamespace returns total persistent volume capacity (bytes) per namespace.
// Sourced from kube-state-metrics (PVCs only) — emptyDir, configMap, secret, and
// projected volumes are correctly excluded.
func (c *VMClient) DiskByNamespace(ctx context.Context, namespaces []string, start, end time.Time) (map[string]float64, error) {
	if len(namespaces) == 0 {
		return nil, nil
	}
	q := fmt.Sprintf(
		`sum by (namespace) (kube_persistentvolumeclaim_resource_requests_storage_bytes{namespace=~"%s"})`,
		nsRegex(namespaces),
	)
	return c.queryByLabel(ctx, q, end, "namespace")
}

// queryByLabel runs an instant PromQL query at evalTime and returns a map keyed by
// the named label.
func (c *VMClient) queryByLabel(ctx context.Context, promql string, evalTime time.Time, label string) (map[string]float64, error) {
	res, err := c.queryVector(ctx, promql, evalTime)
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(res))
	for _, r := range res {
		key, ok := r.Metric[label]
		if !ok {
			continue
		}
		out[key] = r.Value
	}
	return out, nil
}

type vmSample struct {
	Metric map[string]string
	Value  float64
}

// queryVector executes /api/v1/query and returns the vector result.
func (c *VMClient) queryVector(ctx context.Context, promql string, evalTime time.Time) ([]vmSample, error) {
	form := url.Values{}
	form.Set("query", promql)
	form.Set("time", strconv.FormatInt(evalTime.Unix(), 10))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/api/v1/query", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]any            `json:"value"`
			} `json:"result"`
		} `json:"data"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body=%s)", err, string(body))
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics error: %s: %s", parsed.ErrorType, parsed.Error)
	}

	out := make([]vmSample, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		s, ok := r.Value[1].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected value type %T in response", r.Value[1])
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %q: %w", s, err)
		}
		out = append(out, vmSample{Metric: r.Metric, Value: v})
	}
	return out, nil
}

// nsRegex builds an anchored alternation matcher: ^(ns1|ns2|...)$.
// Namespaces are DNS labels ([a-z0-9-]) so no regex metachar escaping is needed.
func nsRegex(namespaces []string) string {
	return "^(" + strings.Join(namespaces, "|") + ")$"
}

// promDuration renders a Go duration as a Prometheus range vector duration.
// We use seconds for determinism (e.g. 1h → "3600s") since PromQL accepts integer-second strings.
func promDuration(d time.Duration) string {
	return strconv.FormatInt(int64(d.Seconds()), 10) + "s"
}
