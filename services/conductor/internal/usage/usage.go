// Package usage splits a workspace's metered usage by project, environment
// and resource, from the same series cashier bills on.
package usage

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/zeitlos/lucity/pkg/victoria"
	"github.com/zeitlos/lucity/services/conductor/internal/billing"
)

const (
	retention      = 100 * 24 * time.Hour
	ingestionDelay = 5 * time.Minute
	step           = time.Hour
	day            = 24 * time.Hour
	concurrency    = 8
	gibibyte       = 1024 * 1024 * 1024
)

type Kind string

const (
	KindService       Kind = "SERVICE"
	KindDatabase      Kind = "DATABASE"
	KindKeyValueStore Kind = "KEY_VALUE_STORE"
	KindVolume        Kind = "VOLUME"
	KindBucket        Kind = "BUCKET"
	KindAgent         Kind = "AGENT"
)

// Resource is what one service, database, key-value store, volume, bucket
// or agent of an environment cost.
type Resource struct {
	ID    string
	Kind  Kind
	Cents float64
	Daily []float64
}

// Environment is what one environment cost, with the resources it could be
// split into. LastSeen is the last hour it used anything.
type Environment struct {
	Project   string
	ID        string
	LastSeen  time.Time
	Cents     float64
	Daily     []float64
	Resources []Resource
}

// Split is a workspace's usage in one billing period, one value per day.
type Split struct {
	MeasuredUntil time.Time
	Days          int
	Environments  []Environment
}

// Request names the workspace and period to split. Since is when the
// workspace was created: usage before it belongs to an earlier workspace
// with the same ID. Namespaces maps the namespace of every live environment
// to its project and environment ID.
type Request struct {
	Workspace  string
	Period     billing.Period
	Since      time.Time
	Rates      billing.Rates
	Namespaces map[string]Location
}

// Location is the project and environment a namespace belongs to.
type Location struct {
	Project     string
	Environment string
}

type Client struct {
	vm *victoria.Client
}

func New(vmURL string) (*Client, error) {
	vm, err := victoria.New(vmURL)

	if err != nil {
		return nil, err
	}

	return &Client{vm: vm}, nil
}

type sample struct {
	location Location
	resource resourceKey
	meter    billing.Meter
	points   []victoria.Point
}

type resourceKey struct {
	kind Kind
	id   string
}

// Split returns the usage of the request's workspace in its period, up to
// the last full hour.
func (c *Client) Split(ctx context.Context, request Request) (*Split, error) {
	days := int(math.Ceil(request.Period.End.Sub(request.Period.Start).Hours() / 24))
	from := request.Period.Start

	if request.Since.After(from) {
		from = request.Since
	}

	first := ceilHour(from)
	last := time.Now().Add(-ingestionDelay).Truncate(step)

	if end := request.Period.End.Add(-time.Second).Truncate(step); end.Before(last) {
		last = end
	}

	split := &Split{MeasuredUntil: last, Days: days}

	if last.Before(first) {
		split.MeasuredUntil = first
		return split, nil
	}

	namespaces, err := c.namespaces(ctx, request, last)

	if err != nil {
		return nil, err
	}

	queries := c.queries(request.Workspace, namespaces)
	windows := windows(first, last)
	results := make([][]sample, len(queries)*len(windows))
	errs := make([]error, len(results))
	slots := make(chan struct{}, concurrency)
	var group sync.WaitGroup

	for i, query := range queries {
		for j, window := range windows {
			group.Go(func() {
				slots <- struct{}{}
				defer func() { <-slots }()

				index := i*len(windows) + j
				results[index], errs[index] = query(ctx, window[0], window[1])
			})
		}
	}

	group.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	split.Environments = aggregate(slices.Concat(results...), request, days)

	return split, nil
}

type query func(ctx context.Context, first, last time.Time) ([]sample, error)

// windows cuts the hours from first to last, both included, into days, so
// no single query asks VictoriaMetrics for a whole period of minute samples.
func windows(first, last time.Time) [][2]time.Time {
	var result [][2]time.Time

	for start := first; !start.After(last); start = start.Add(day) {
		end := start.Add(day - step)

		if end.After(last) {
			end = last
		}

		result = append(result, [2]time.Time{start, end})
	}

	return result
}

func (c *Client) queries(workspace string, namespaces map[string]Location) []query {
	workload := func(selector string) string {
		return fmt.Sprintf(
			`label_replace(label_replace(%s, "workload", "lucity-app-$1-$2", "k8s_pod_name", "lucity-app-(pg|valkey)-(.+)-[0-9]+"), "workload", "lucity-app-pg-$1", "k8s_pod_name", "lucity-app-pg-(.+)-[0-9]+-(initdb|join|import|pgbasebackup|full-recovery|snapshot-recovery)-[a-z0-9]+")`,
			selector,
		)
	}
	grouping := "k8s_namespace_name, project, environment, k8s_deployment_name, workload"

	queries := []query{
		c.podQuery(billing.MeterCPU, fmt.Sprintf(
			`sum by (%s) (%s)`,
			grouping, workload(fmt.Sprintf(`increase(container_cpu_time_seconds_total{workspace=%q}[1h])`, workspace)),
		), 1.0/60),
		c.podQuery(billing.MeterMemory, fmt.Sprintf(
			`avg_over_time((sum by (%s) (%s))[1h:1m])`,
			grouping, workload(fmt.Sprintf(`container_memory_working_set_bytes{workspace=%q}`, workspace)),
		), 60.0/gibibyte),
		c.bucketQuery(fmt.Sprintf(
			`avg_over_time((max by (project, environment, bucket) (lucity_bucket_size_bytes{workspace=%q}))[1h:1m])`,
			workspace,
		)),
	}

	if len(namespaces) == 0 {
		return queries
	}

	names := make([]string, 0, len(namespaces))

	for namespace := range namespaces {
		names = append(names, regexp.QuoteMeta(namespace))
	}

	slices.Sort(names)
	alternatives := strings.Join(names, "|")

	return append(queries,
		c.namespaceQuery(billing.MeterDisk, fmt.Sprintf(
			`avg_over_time((max by (namespace, persistentvolumeclaim) (kube_persistentvolumeclaim_resource_requests_storage_bytes{namespace=~%q}))[1h:1m])`,
			"("+alternatives+")",
		), "persistentvolumeclaim", volumeResource, namespaces),
		c.namespaceQuery(billing.MeterEgress, fmt.Sprintf(
			`sum by (namespace, k8s_service) (label_replace(label_replace(increase(envoy_cluster_upstream_cx_rx_bytes_total{envoy_cluster_name=~%q}[1h]), "namespace", "$1", "envoy_cluster_name", "lucity-system/cilium-gateway-[^/]+/([a-z0-9-]+)_.+"), "k8s_service", "$1", "envoy_cluster_name", "lucity-system/cilium-gateway-[^/]+/[a-z0-9-]+_([a-z0-9-]+)_.+"))`,
			"lucity-system/cilium-gateway-[^/]+/("+alternatives+")_.+",
		), "k8s_service", serviceResource, namespaces),
	)
}

func (c *Client) podQuery(meter billing.Meter, promQL string, scale float64) query {
	return func(ctx context.Context, first, last time.Time) ([]sample, error) {
		series, err := c.vm.QueryRange(ctx, promQL, first, last, step)

		if err != nil {
			return nil, fmt.Errorf("query %s usage: %w", meter, err)
		}

		samples := make([]sample, 0, len(series))

		for _, s := range series {
			samples = append(samples, sample{
				location: Location{Project: s.Labels["project"], Environment: s.Labels["environment"]},
				resource: podResource(s.Labels["k8s_deployment_name"], s.Labels["workload"]),
				meter:    meter,
				points:   scaled(s.Points, scale),
			})
		}

		return samples, nil
	}
}

func (c *Client) bucketQuery(promQL string) query {
	return func(ctx context.Context, first, last time.Time) ([]sample, error) {
		series, err := c.vm.QueryRange(ctx, promQL, first, last, step)

		if err != nil {
			return nil, fmt.Errorf("query storage usage: %w", err)
		}

		samples := make([]sample, 0, len(series))

		for _, s := range series {
			samples = append(samples, sample{
				location: Location{Project: s.Labels["project"], Environment: s.Labels["environment"]},
				resource: resourceKey{kind: KindBucket, id: s.Labels["bucket"]},
				meter:    billing.MeterStorage,
				points:   scaled(s.Points, 1.0/gibibyte),
			})
		}

		return samples, nil
	}
}

func (c *Client) namespaceQuery(meter billing.Meter, promQL, resourceLabel string, resource func(string) resourceKey, namespaces map[string]Location) query {
	return func(ctx context.Context, first, last time.Time) ([]sample, error) {
		series, err := c.vm.QueryRange(ctx, promQL, first, last, step)

		if err != nil {
			return nil, fmt.Errorf("query %s usage: %w", meter, err)
		}

		samples := make([]sample, 0, len(series))

		for _, s := range series {
			location, ok := namespaces[s.Labels["namespace"]]

			if !ok {
				continue
			}

			samples = append(samples, sample{
				location: location,
				resource: resource(s.Labels[resourceLabel]),
				meter:    meter,
				points:   scaled(s.Points, 1.0/gibibyte),
			})
		}

		return samples, nil
	}
}

func (c *Client) namespaces(ctx context.Context, request Request, at time.Time) (map[string]Location, error) {
	lookback := min(at.Sub(request.Since), retention)
	namespaces := make(map[string]Location, len(request.Namespaces))

	for namespace, location := range request.Namespaces {
		namespaces[namespace] = location
	}

	if lookback <= 0 {
		return namespaces, nil
	}

	samples, err := c.vm.QueryVector(ctx, fmt.Sprintf(
		`group by (k8s_namespace_name, project, environment) (last_over_time(container_memory_working_set_bytes{workspace=%q}[%ds]))`,
		request.Workspace, int64(lookback.Seconds()),
	), at)

	if err != nil {
		return nil, fmt.Errorf("query namespaces: %w", err)
	}

	for _, s := range samples {
		namespace := s.Labels["k8s_namespace_name"]

		if _, ok := namespaces[namespace]; ok || namespace == "" || s.Labels["project"] == "" || s.Labels["environment"] == "" {
			continue
		}

		namespaces[namespace] = Location{Project: s.Labels["project"], Environment: s.Labels["environment"]}
	}

	return namespaces, nil
}

func aggregate(samples []sample, request Request, days int) []Environment {
	environments := map[Location]*Environment{}
	resources := map[Location]map[resourceKey]*Resource{}

	for _, s := range samples {
		if s.location.Project == "" || s.location.Environment == "" {
			continue
		}

		rate := request.Rates[s.meter]
		environment, ok := environments[s.location]

		if !ok {
			environment = &Environment{Project: s.location.Project, ID: s.location.Environment, Daily: make([]float64, days)}
			environments[s.location] = environment
			resources[s.location] = map[resourceKey]*Resource{}
		}

		var resource *Resource

		if s.resource.id != "" {
			resource, ok = resources[s.location][s.resource]

			if !ok {
				resource = &Resource{ID: s.resource.id, Kind: s.resource.kind, Daily: make([]float64, days)}
				resources[s.location][s.resource] = resource
			}
		}

		for _, point := range s.points {
			if point.Value <= 0 || math.IsNaN(point.Value) {
				continue
			}

			cents := point.Value * rate
			index := dayIndex(point.Time, request.Period.Start, days)
			environment.Daily[index] += cents
			environment.Cents += cents

			if point.Time.After(environment.LastSeen) {
				environment.LastSeen = point.Time
			}

			if resource != nil {
				resource.Daily[index] += cents
				resource.Cents += cents
			}
		}
	}

	list := make([]Environment, 0, len(environments))

	for key, environment := range environments {
		for _, resource := range resources[key] {
			if resource.Cents > 0 {
				environment.Resources = append(environment.Resources, *resource)
			}
		}

		slices.SortFunc(environment.Resources, func(a, b Resource) int {
			return compareCents(a.Cents, b.Cents, a.ID, b.ID)
		})

		if environment.Cents > 0 {
			list = append(list, *environment)
		}
	}

	slices.SortFunc(list, func(a, b Environment) int {
		return compareCents(a.Cents, b.Cents, a.Project+"/"+a.ID, b.Project+"/"+b.ID)
	})

	return list
}

var (
	databasePodWorkload   = regexp.MustCompile(`^lucity-app-pg-(.+)$`)
	keyValueStoreWorkload = regexp.MustCompile(`^lucity-app-valkey-(.+)$`)
	databaseClaim         = regexp.MustCompile(`^lucity-app-pg-(.+)-[0-9]+(-wal)?$`)
	keyValueStoreClaim    = regexp.MustCompile(`^data-lucity-app-valkey-(.+)-[0-9]+$`)
	databaseServiceName   = regexp.MustCompile(`^lucity-app-pg-(.+)-(rw|ro|r)$`)
	agentPrefix           = "lucity-app-agent-"
	volumePrefix          = "lucity-app-volume-"
	serviceWorkloadPrefix = "lucity-app-"
)

func podResource(deployment, workload string) resourceKey {
	if id, ok := strings.CutPrefix(deployment, agentPrefix); ok {
		return resourceKey{kind: KindAgent, id: id}
	}

	if id, ok := strings.CutPrefix(deployment, serviceWorkloadPrefix); ok {
		return resourceKey{kind: KindService, id: id}
	}

	if match := databasePodWorkload.FindStringSubmatch(workload); match != nil {
		return resourceKey{kind: KindDatabase, id: match[1]}
	}

	if match := keyValueStoreWorkload.FindStringSubmatch(workload); match != nil {
		return resourceKey{kind: KindKeyValueStore, id: match[1]}
	}

	return resourceKey{}
}

func volumeResource(claim string) resourceKey {
	if id, ok := strings.CutPrefix(claim, volumePrefix); ok {
		return resourceKey{kind: KindVolume, id: id}
	}

	if id, ok := strings.CutPrefix(claim, agentPrefix); ok {
		return resourceKey{kind: KindAgent, id: id}
	}

	if match := databaseClaim.FindStringSubmatch(claim); match != nil {
		return resourceKey{kind: KindDatabase, id: match[1]}
	}

	if match := keyValueStoreClaim.FindStringSubmatch(claim); match != nil {
		return resourceKey{kind: KindKeyValueStore, id: match[1]}
	}

	return resourceKey{}
}

func serviceResource(name string) resourceKey {
	if id, ok := strings.CutPrefix(name, agentPrefix); ok {
		return resourceKey{kind: KindAgent, id: id}
	}

	if match := databaseServiceName.FindStringSubmatch(name); match != nil {
		return resourceKey{kind: KindDatabase, id: match[1]}
	}

	if id, ok := strings.CutPrefix(name, serviceWorkloadPrefix); ok {
		return resourceKey{kind: KindService, id: id}
	}

	return resourceKey{}
}

func scaled(points []victoria.Point, factor float64) []victoria.Point {
	out := make([]victoria.Point, len(points))

	for i, point := range points {
		out[i] = victoria.Point{Time: point.Time, Value: point.Value * factor}
	}

	return out
}

func dayIndex(at, start time.Time, days int) int {
	index := int(at.Add(-time.Second).Sub(start) / day)

	return max(0, min(days-1, index))
}

func ceilHour(t time.Time) time.Time {
	truncated := t.Truncate(step)

	if truncated.Equal(t) {
		return t
	}

	return truncated.Add(step)
}

func compareCents(a, b float64, nameA, nameB string) int {
	if a != b {
		if a > b {
			return -1
		}

		return 1
	}

	return strings.Compare(nameA, nameB)
}
