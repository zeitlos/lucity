package pipeline

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sort"
	"sync"
	"time"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/jobs"
	"github.com/zeitlos/lucity/pkg/labels"
)

type Interface interface {
	Reconcile(ctx context.Context) error
	QueuedRuns(ctx context.Context, workspace string) (int, error)
}

const labelComponent = labels.Prefix + "component"

var (
	buildSelector  = mustSelector(labelComponent + " in (build, scan)")
	deploySelector = mustSelector(labelComponent + "=deploy")
)

var errJobsNotSynced = errors.New("job cache has not synced yet")

const (
	builderSlots    = 1
	builderSpill    = 2
	builderWarmWait = time.Minute
)

type Client struct {
	kubernetes    kubernetes.Interface
	jobs          *jobs.Cache
	builders      *jobs.Builders
	maxConcurrent int

	mu      sync.Mutex
	resumed map[string]resumption
}

type resumption struct {
	resourceVersion string
	builder         string
}

// New returns an admission client. builders may be nil for clients that only
// count queued runs; such a client never admits a build.
func New(kubernetes kubernetes.Interface, jobs *jobs.Cache, builders *jobs.Builders, maxConcurrent int) *Client {
	return &Client{
		kubernetes:    kubernetes,
		jobs:          jobs,
		builders:      builders,
		maxConcurrent: maxConcurrent,
		resumed:       map[string]resumption{},
	}
}

var _ Interface = (*Client)(nil)

type runState int

const (
	runTerminal runState = iota
	runQueued
	runActive
)

type job struct {
	namespace       string
	name            string
	resourceVersion string
	build           bool
	source          string
	builder         string
	suspended       bool
	terminal        bool
}

type run struct {
	workspace   string
	environment string
	key         string
	createdAt   time.Time
	jobs        []job
}

func (r run) state() runState {
	queued := false

	for _, j := range r.jobs {
		if j.terminal {
			continue
		}

		if !j.suspended {
			return runActive
		}

		queued = true
	}

	if queued {
		return runQueued
	}

	return runTerminal
}

func (r run) pendingBuild() (job, bool) {
	for _, j := range r.jobs {
		if j.build && j.suspended && !j.terminal {
			return j, true
		}
	}

	return job{}, false
}

// Reconcile admits queued release runs: at most one active run per environment
// (the Helm-release boundary, so concurrent applies to the same release can't
// race), oldest first, bounded by the global concurrency cap. A run that still
// has to build also needs a free builder: its source is ranked against the
// ready builders, so builds of one source keep landing on the builder that
// holds their warm cache. Deploy-only runs pass waiting builds.
func (c *Client) Reconcile(ctx context.Context) error {
	if !c.jobs.HasSynced() {
		return nil
	}

	runs, err := c.runs(ctx)

	if err != nil {
		return err
	}

	ready, err := c.readyBuilders()

	if err != nil {
		return err
	}

	load := builderLoad(runs)
	active := 0
	activeEnvironments := map[string]bool{}
	queuedByEnvironment := map[string][]run{}

	for _, r := range runs {
		switch r.state() {
		case runActive:
			active++
			activeEnvironments[r.environment] = true

			builder := ""

			if build, ok := r.pendingBuild(); ok {
				builder, _, ok = place(build.source, builderWarmWait, ready, load)

				if !ok {
					continue
				}
			}

			if err := c.resume(ctx, r, builder); err != nil {
				return fmt.Errorf("resume partially admitted run %q: %w", r.key, err)
			}

			if builder != "" {
				load[builder]++
			}
		case runQueued:
			queuedByEnvironment[r.environment] = append(queuedByEnvironment[r.environment], r)
		}
	}

	var candidates []run

	for environment, queue := range queuedByEnvironment {
		if !activeEnvironments[environment] {
			candidates = append(candidates, queue[0])
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if !candidates[i].createdAt.Equal(candidates[j].createdAt) {
			return candidates[i].createdAt.Before(candidates[j].createdAt)
		}

		return candidates[i].key < candidates[j].key
	})

	for _, candidate := range candidates {
		if c.maxConcurrent > 0 && active >= c.maxConcurrent {
			break
		}

		builder := ""
		choice := 0

		if build, ok := candidate.pendingBuild(); ok {
			builder, choice, ok = place(build.source, time.Since(candidate.createdAt), ready, load)

			if !ok {
				continue
			}
		}

		if err := c.resume(ctx, candidate, builder); err != nil {
			return fmt.Errorf("admit run %q: %w", candidate.key, err)
		}

		active++

		if builder != "" {
			load[builder]++
			slog.Info("build placed", "run", candidate.key, "builder", builder, "choice", choice)
		}
	}

	return nil
}

func (c *Client) readyBuilders() ([]string, error) {
	if c.builders == nil || !c.builders.HasSynced() {
		return nil, nil
	}

	return c.builders.Ready()
}

func builderLoad(runs []run) map[string]int {
	load := map[string]int{}

	for _, r := range runs {
		for _, j := range r.jobs {
			if j.build && !j.suspended && !j.terminal && j.builder != "" {
				load[j.builder]++
			}
		}
	}

	return load
}

// place picks the builder for a build of source. It walks the source's
// ranking and takes the first builder with a free slot, looking only at the
// first few choices until the build has waited builderWarmWait, so a warm
// builder that frees up soon wins over a cold one now.
func place(source string, waited time.Duration, ready []string, load map[string]int) (string, int, bool) {
	for i, builder := range rank(source, ready) {
		if i >= builderSpill && waited < builderWarmWait {
			return "", 0, false
		}

		if load[builder] < builderSlots {
			return builder, i + 1, true
		}
	}

	return "", 0, false
}

func rank(source string, builders []string) []string {
	ranked := slices.Clone(builders)

	slices.SortFunc(ranked, func(a, b string) int {
		return cmp.Compare(score(source, b), score(source, a))
	})

	return ranked
}

func score(source, builder string) uint64 {
	sum := sha256.Sum256([]byte(source + "\x00" + builder))

	return binary.BigEndian.Uint64(sum[:8])
}

func (c *Client) QueuedRuns(ctx context.Context, workspace string) (int, error) {
	if !c.jobs.HasSynced() {
		return 0, errJobsNotSynced
	}

	runs, err := c.runs(ctx)

	if err != nil {
		return 0, err
	}

	count := 0

	for _, r := range runs {
		if r.workspace == workspace && r.state() == runQueued {
			count++
		}
	}

	return count, nil
}

var unsuspendPatch = []byte(`{"spec":{"suspend":false}}`)

// resume unsuspends the run's suspended Jobs. A build Job also gets builder
// written onto it and its pod template, in the same patch: the pod reads it to
// pick its BuildKit daemon, and a suspended Job that never started still
// accepts pod template annotation changes.
func (c *Client) resume(ctx context.Context, r run, builder string) error {
	for _, j := range r.jobs {
		if !j.suspended || j.terminal {
			continue
		}

		patch := unsuspendPatch
		placed := ""

		if j.build {
			if builder == "" {
				continue
			}

			var err error

			if patch, err = builderPatch(builder); err != nil {
				return err
			}

			placed = builder
		}

		c.expectResumed(j, placed)

		_, err := c.kubernetes.BatchV1().Jobs(j.namespace).Patch(ctx, j.name, types.StrategicMergePatchType, patch, meta.PatchOptions{})

		if err != nil {
			c.forgetResumed(j)
			return fmt.Errorf("unsuspend job %s/%s: %w", j.namespace, j.name, err)
		}
	}

	return nil
}

func builderPatch(builder string) ([]byte, error) {
	annotations := map[string]string{labels.Builder: builder}

	return json.Marshal(map[string]any{
		"metadata": map[string]any{"annotations": annotations},
		"spec": map[string]any{
			"suspend": false,
			"template": map[string]any{
				"metadata": map[string]any{"annotations": annotations},
			},
		},
	})
}

func (c *Client) runs(ctx context.Context) ([]run, error) {
	builds, err := c.jobs.Builds(ctx, buildSelector)

	if err != nil {
		return nil, err
	}

	deploys, err := c.jobs.Deploys(ctx, deploySelector)

	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	pending := map[string]bool{}
	groups := map[string]*run{}

	add := func(k8sJob *batch.Job) {
		workspace := k8sJob.Labels[labels.Workspace]

		if workspace == "" {
			return
		}

		key := k8sJob.Labels[labels.Release]

		if key == "" {
			key = k8sJob.Name
		}

		environment := workspace

		if project := k8sJob.Labels[labels.Project]; project != "" {
			if env := k8sJob.Labels[labels.Environment]; env != "" {
				environment = workspace + "/" + project + "/" + env
			}
		}

		groupKey := workspace + "/" + key
		group, ok := groups[groupKey]

		if !ok {
			group = &run{workspace: workspace, environment: environment, key: groupKey, createdAt: k8sJob.CreationTimestamp.Time}
			groups[groupKey] = group
		}

		if k8sJob.CreationTimestamp.Time.Before(group.createdAt) {
			group.createdAt = k8sJob.CreationTimestamp.Time
		}

		suspended := k8sJob.Spec.Suspend != nil && *k8sJob.Spec.Suspend
		builder := k8sJob.Annotations[labels.Builder]
		jobKey := k8sJob.Namespace + "/" + k8sJob.Name

		if expected, ok := c.resumed[jobKey]; ok && expected.resourceVersion == k8sJob.ResourceVersion {
			suspended = false
			pending[jobKey] = true

			if expected.builder != "" {
				builder = expected.builder
			}
		}

		build := k8sJob.Labels[labelComponent] == "build"
		source := ""

		if build {
			source = workspace + "\x00" + k8sJob.Annotations[labels.SourceRepo] + "\x00" + k8sJob.Annotations[labels.SourceContext]
		}

		group.jobs = append(group.jobs, job{
			namespace:       k8sJob.Namespace,
			name:            k8sJob.Name,
			resourceVersion: k8sJob.ResourceVersion,
			build:           build,
			source:          source,
			builder:         builder,
			suspended:       suspended,
			terminal:        jobTerminal(k8sJob),
		})
	}

	for _, item := range builds {
		add(item)
	}

	for _, item := range deploys {
		add(item)
	}

	for key := range c.resumed {
		if !pending[key] {
			delete(c.resumed, key)
		}
	}

	runs := make([]run, 0, len(groups))

	for _, group := range groups {
		runs = append(runs, *group)
	}

	sort.Slice(runs, func(i, j int) bool {
		if !runs[i].createdAt.Equal(runs[j].createdAt) {
			return runs[i].createdAt.Before(runs[j].createdAt)
		}

		return runs[i].key < runs[j].key
	})

	return runs, nil
}

func (c *Client) expectResumed(j job, builder string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.resumed[j.namespace+"/"+j.name] = resumption{resourceVersion: j.resourceVersion, builder: builder}
}

func (c *Client) forgetResumed(j job) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.resumed, j.namespace+"/"+j.name)
}

func mustSelector(selector string) k8slabels.Selector {
	parsed, err := k8slabels.Parse(selector)

	if err != nil {
		panic(err)
	}

	return parsed
}

func jobTerminal(k8sJob *batch.Job) bool {
	for _, condition := range k8sJob.Status.Conditions {
		if condition.Status != core.ConditionTrue {
			continue
		}

		if condition.Type == batch.JobComplete || condition.Type == batch.JobFailed {
			return true
		}
	}

	return false
}
