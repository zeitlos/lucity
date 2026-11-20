package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	batch "k8s.io/api/batch/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"

	helmDeployer "github.com/zeitlos/lucity/pkg/deployer/helm"
	"github.com/zeitlos/lucity/pkg/jobs"
	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/pkg/pipeline"
)

const labelComponent = labels.Prefix + "component"

var (
	helmReleasesDesc = prometheus.NewDesc(
		"lucity_helm_releases",
		"Helm releases by chart and the status of their newest revision.",
		[]string{"chart", "status"},
		nil,
	)
	releasesDesc = prometheus.NewDesc(
		"lucity_releases",
		"Releases in flight by status.",
		[]string{"status"},
		nil,
	)
)

type collector struct {
	appReleases       *helmDeployer.ReleaseCache
	workspaceReleases *helmDeployer.WorkspaceReleaseStatuses
	pipeline          *pipeline.Client
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- helmReleasesDesc
	ch <- releasesDesc
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	if statuses, ok := c.appReleases.Statuses(); ok {
		collectHelmReleases(ch, "app", statuses)
	}

	if c.workspaceReleases != nil {
		if statuses, ok := c.workspaceReleases.Statuses(); ok {
			collectHelmReleases(ch, "workspace", statuses)
		}
	}

	releases, err := c.pipeline.Releases(context.Background())

	if err != nil {
		return
	}

	ch <- prometheus.MustNewConstMetric(releasesDesc, prometheus.GaugeValue, float64(releases.Queued), "queued")
	ch <- prometheus.MustNewConstMetric(releasesDesc, prometheus.GaugeValue, float64(releases.Building), "building")
	ch <- prometheus.MustNewConstMetric(releasesDesc, prometheus.GaugeValue, float64(releases.Deploying), "deploying")
}

func collectHelmReleases(ch chan<- prometheus.Metric, chart string, statuses map[string]int) {
	for status, count := range statuses {
		ch <- prometheus.MustNewConstMetric(helmReleasesDesc, prometheus.GaugeValue, float64(count), chart, status)
	}
}

type jobMetrics struct {
	cache    *jobs.Cache
	queue    *prometheus.HistogramVec
	duration *prometheus.HistogramVec
}

func registerJobMetrics(cache *jobs.Cache, registry prometheus.Registerer) error {
	m := &jobMetrics{
		cache: cache,
		queue: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "lucity_job_queue_seconds",
			Help:    "Time a build, scan or deploy Job waited for admission.",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 14),
		}, []string{"component"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "lucity_job_duration_seconds",
			Help:    "Time a build, scan or deploy Job ran until it succeeded or failed.",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 14),
		}, []string{"component", "outcome"}),
	}

	registry.MustRegister(m.queue, m.duration)

	return cache.OnFinished(m.observe)
}

func (m *jobMetrics) observe(job *batch.Job) {
	finishedAt, failed, _ := jobs.Finished(job)

	if job.Status.StartTime == nil {
		return
	}

	component := job.Labels[labelComponent]
	started := job.Status.StartTime.Time

	m.queue.WithLabelValues(component).Observe(started.Sub(job.CreationTimestamp.Time).Seconds())

	if component == "deploy" {
		if built, ok := m.buildFinished(job); ok && built.After(started) {
			started = built
		}
	}

	outcome := "succeeded"

	if failed {
		outcome = "failed"
	}

	m.duration.WithLabelValues(component, outcome).Observe(finishedAt.Sub(started).Seconds())
}

func (m *jobMetrics) buildFinished(job *batch.Job) (time.Time, bool) {
	release := job.Labels[labels.Release]

	if release == "" {
		return time.Time{}, false
	}

	selector := k8slabels.SelectorFromSet(k8slabels.Set{
		labels.Workspace: job.Labels[labels.Workspace],
		labels.Release:   release,
		labelComponent:   "build",
	})

	builds, err := m.cache.Builds(context.Background(), selector)

	if err != nil {
		return time.Time{}, false
	}

	for _, build := range builds {
		if at, _, done := jobs.Finished(build); done {
			return at, true
		}
	}

	return time.Time{}, false
}
