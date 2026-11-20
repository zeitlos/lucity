package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Loops records how long each pass of a background loop took and when the
// loop last passed without an error.
type Loops struct {
	duration    *prometheus.HistogramVec
	lastSuccess *prometheus.GaugeVec
}

func NewLoops(registry prometheus.Registerer) *Loops {
	loops := &Loops{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "lucity_loop_duration_seconds",
			Help:    "Duration of a background loop pass.",
			Buckets: prometheus.ExponentialBuckets(0.01, 4, 9),
		}, []string{"loop", "outcome"}),
		lastSuccess: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lucity_loop_last_success_timestamp_seconds",
			Help: "Unix time of the last background loop pass without an error.",
		}, []string{"loop"}),
	}

	registry.MustRegister(loops.duration, loops.lastSuccess)

	return loops
}

// Observe records a pass of loop that began at start and ended with err.
func (l *Loops) Observe(loop string, start time.Time, err error) {
	outcome := "success"

	if err != nil {
		outcome = "failure"
	}

	l.duration.WithLabelValues(loop, outcome).Observe(time.Since(start).Seconds())

	if err == nil {
		l.lastSuccess.WithLabelValues(loop).SetToCurrentTime()
	}
}
