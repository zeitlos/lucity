package storage

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/zeitlos/lucity/pkg/labels"
	"github.com/zeitlos/lucity/services/exporter/ovh"
)

const (
	source         = "object_storage"
	refreshTimeout = 2 * time.Minute
)

var (
	sizeDesc = prometheus.NewDesc(
		"lucity_bucket_size_bytes",
		"Bytes stored in a bucket.",
		[]string{"workspace", "project", "environment", "bucket"},
		nil,
	)
	lastSuccessDesc = prometheus.NewDesc(
		"lucity_exporter_last_success_timestamp_seconds",
		"Unix time of the last successful refresh of a usage source.",
		[]string{"source"},
		nil,
	)
)

type bucketID struct {
	workspace   string
	project     string
	environment string
	name        string
}

type Collector struct {
	ovh      *ovh.Client
	region   string
	interval time.Duration

	ids map[string]bucketID

	mu          sync.RWMutex
	sizes       map[bucketID]int64
	lastSuccess time.Time

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func New(client *ovh.Client, region string, interval time.Duration) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		ovh:      client,
		region:   region,
		interval: interval,
		ids:      make(map[string]bucketID),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
}

func (c *Collector) Label() string { return "Object Storage Collector" }

func (c *Collector) Start() error {
	defer close(c.done)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		if err := c.refresh(c.ctx); err != nil && c.ctx.Err() == nil {
			slog.Warn("object storage refresh failed", "error", err)
		}

		select {
		case <-c.ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (c *Collector) Shutdown(ctx context.Context) error {
	c.cancel()

	select {
	case <-c.done:
	case <-ctx.Done():
	}

	return nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sizeDesc
	ch <- lastSuccessDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.lastSuccess.IsZero() {
		return
	}

	for id, bytes := range c.sizes {
		ch <- prometheus.MustNewConstMetric(sizeDesc, prometheus.GaugeValue, float64(bytes), id.workspace, id.project, id.environment, id.name)
	}

	ch <- prometheus.MustNewConstMetric(lastSuccessDesc, prometheus.GaugeValue, float64(c.lastSuccess.Unix()), source)
}

func (c *Collector) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	buckets, err := c.ovh.Buckets(ctx, c.region)

	if err != nil {
		return err
	}

	sizes := make(map[bucketID]int64)
	listed := make(map[string]bool, len(buckets))

	for _, bucket := range buckets {
		listed[bucket.Name] = true
		id, ok := c.ids[bucket.Name]

		if !ok {
			detail, err := c.ovh.Bucket(ctx, c.region, bucket.Name)

			if err != nil {
				return err
			}

			id = bucketID{
				workspace:   detail.Tags[labels.Workspace],
				project:     detail.Tags[labels.Project],
				environment: detail.Tags[labels.Environment],
				name:        detail.Tags[labels.ObjectStorageBucket],
			}
			c.ids[bucket.Name] = id
		}

		if id.workspace == "" {
			continue
		}

		sizes[id] += bucket.ObjectsSize
	}

	for name := range c.ids {
		if !listed[name] {
			delete(c.ids, name)
		}
	}

	c.mu.Lock()
	c.sizes = sizes
	c.lastSuccess = time.Now()
	c.mu.Unlock()

	slog.Debug("object storage refreshed", "buckets", len(buckets), "billable", len(sizes))

	return nil
}
