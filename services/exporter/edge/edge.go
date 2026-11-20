package edge

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/zeitlos/lucity/pkg/bunny"
	"github.com/zeitlos/lucity/pkg/metrics"
)

const (
	loop           = "edge_zones"
	refreshTimeout = time.Minute
)

var zonesDesc = prometheus.NewDesc(
	"lucity_edge_zones",
	"Zones in use in the edge CDN account.",
	nil,
	nil,
)

type Collector struct {
	bunny    *bunny.Client
	loops    *metrics.Loops
	interval time.Duration

	mu        sync.RWMutex
	zones     int
	refreshed bool

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func New(client *bunny.Client, loops *metrics.Loops, interval time.Duration) *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	return &Collector{
		bunny:    client,
		loops:    loops,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
}

func (c *Collector) Label() string { return "Edge Zone Collector" }

func (c *Collector) Start() error {
	defer close(c.done)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		start := time.Now()
		err := c.refresh(c.ctx)
		c.loops.Observe(loop, start, err)

		if err != nil && c.ctx.Err() == nil {
			slog.Warn("edge zone refresh failed", "error", err)
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
	ch <- zonesDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.refreshed {
		return
	}

	ch <- prometheus.MustNewConstMetric(zonesDesc, prometheus.GaugeValue, float64(c.zones))
}

func (c *Collector) refresh(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()

	zones, err := c.bunny.PullZones(ctx)

	if err != nil {
		return err
	}

	c.mu.Lock()
	c.zones = len(zones)
	c.refreshed = true
	c.mu.Unlock()

	return nil
}
