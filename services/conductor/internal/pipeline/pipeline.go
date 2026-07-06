package pipeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/zeitlos/lucity/pkg/labels"
)

type Interface interface {
	Reconcile(ctx context.Context) error
	QueuedRuns(ctx context.Context, workspace string) (int, error)
}

const labelComponent = labels.Prefix + "component"

type Client struct {
	kubernetes      kubernetes.Interface
	buildNamespace  string
	deployNamespace string
	maxConcurrent   int
}

func New(kubernetes kubernetes.Interface, buildNamespace, deployNamespace string, maxConcurrent int) *Client {
	return &Client{
		kubernetes:      kubernetes,
		buildNamespace:  buildNamespace,
		deployNamespace: deployNamespace,
		maxConcurrent:   maxConcurrent,
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
	namespace string
	name      string
	suspended bool
	terminal  bool
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

// Reconcile admits queued release runs: at most one active run per environment
// (the Helm-release boundary, so concurrent applies to the same release can't
// race), oldest first, bounded by the global concurrency cap.
func (c *Client) Reconcile(ctx context.Context) error {
	runs, err := c.runs(ctx)

	if err != nil {
		return err
	}

	active := 0
	activeEnvironments := map[string]bool{}
	queuedByEnvironment := map[string][]run{}

	for _, r := range runs {
		switch r.state() {
		case runActive:
			active++
			activeEnvironments[r.environment] = true

			if err := c.resume(ctx, r); err != nil {
				return fmt.Errorf("resume partially admitted run %q: %w", r.key, err)
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

		if err := c.resume(ctx, candidate); err != nil {
			return fmt.Errorf("admit run %q: %w", candidate.key, err)
		}

		active++
	}

	return nil
}

func (c *Client) QueuedRuns(ctx context.Context, workspace string) (int, error) {
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

func (c *Client) resume(ctx context.Context, r run) error {
	for _, j := range r.jobs {
		if !j.suspended || j.terminal {
			continue
		}

		_, err := c.kubernetes.BatchV1().Jobs(j.namespace).Patch(ctx, j.name, types.StrategicMergePatchType, unsuspendPatch, meta.PatchOptions{})

		if err != nil {
			return fmt.Errorf("unsuspend job %s/%s: %w", j.namespace, j.name, err)
		}
	}

	return nil
}

func (c *Client) runs(ctx context.Context) ([]run, error) {
	builds, err := c.kubernetes.BatchV1().Jobs(c.buildNamespace).List(ctx, meta.ListOptions{
		LabelSelector: labelComponent + " in (build, scan)",
	})

	if err != nil {
		return nil, err
	}

	deploys, err := c.kubernetes.BatchV1().Jobs(c.deployNamespace).List(ctx, meta.ListOptions{
		LabelSelector: labelComponent + "=deploy",
	})

	if err != nil {
		return nil, err
	}

	groups := map[string]*run{}

	add := func(k8sJob batch.Job) {
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

		group.jobs = append(group.jobs, job{
			namespace: k8sJob.Namespace,
			name:      k8sJob.Name,
			suspended: k8sJob.Spec.Suspend != nil && *k8sJob.Spec.Suspend,
			terminal:  jobTerminal(k8sJob),
		})
	}

	for _, item := range builds.Items {
		add(item)
	}

	for _, item := range deploys.Items {
		add(item)
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

func jobTerminal(k8sJob batch.Job) bool {
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
