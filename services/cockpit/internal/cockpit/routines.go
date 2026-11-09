package cockpit

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/zeitlos/lucity/pkg/victoria"
)

// Routine is a background loop of the platform, such as the release pass.
type Routine struct {
	Name string
	// Hours holds the passes of the last 24 hours, oldest first. The last
	// hour is the current one, so far.
	Hours []RoutineHour
	// AverageDuration is the average pass over the last 24 hours, nil
	// without passes.
	AverageDuration *time.Duration
	// AverageDurationChange compares AverageDuration with the 7 days before,
	// as a ratio: 0.5 is 50% slower. nil without passes in either window.
	AverageDurationChange *float64
	LastPassedAt          *time.Time
}

type RoutineHour struct {
	Start  time.Time
	Passed int
	Failed int
}

const (
	routinePasses  = `sum by (loop, outcome) (increase(lucity_loop_duration_seconds_count[%s]))`
	routineAverage = `sum by (loop) (increase(lucity_loop_duration_seconds_sum%[1]s)) / sum by (loop) (increase(lucity_loop_duration_seconds_count%[1]s))`
)

// Routines returns every loop that ran in the last 24 hours or passed in the
// last 7 days, by name.
func (c *Client) Routines(ctx context.Context) ([]Routine, error) {
	now := time.Now()
	current := now.Truncate(time.Hour)
	first := current.Add(-23 * time.Hour)

	var hourly []victoria.Series
	var currentHour []victoria.Sample
	var average []victoria.Sample
	var averageBefore []victoria.Sample
	var lastPassed []victoria.Sample

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		series, err := c.metrics.QueryRange(ctx, fmt.Sprintf(routinePasses, "1h"), first.Add(time.Hour), current, time.Hour)
		hourly = series

		return wrap("routine passes", err)
	})

	group.Go(func() error {
		window := fmt.Sprintf("%ds", max(int(now.Sub(current).Seconds()), 1))
		samples, err := c.metrics.QueryVector(ctx, fmt.Sprintf(routinePasses, window), now)
		currentHour = samples

		return wrap("routine passes this hour", err)
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, fmt.Sprintf(routineAverage, "[24h]"), now)
		average = samples

		return wrap("routine durations", err)
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, fmt.Sprintf(routineAverage, "[7d] offset 24h"), now)
		averageBefore = samples

		return wrap("routine durations before", err)
	})

	group.Go(func() error {
		samples, err := c.metrics.QueryVector(ctx, `max by (loop) (last_over_time(lucity_loop_last_success_timestamp_seconds[7d]))`, now)
		lastPassed = samples

		return wrap("routine last passes", err)
	})

	if err := group.Wait(); err != nil {
		return nil, err
	}

	routines := map[string]*Routine{}

	routine := func(name string) *Routine {
		if routines[name] == nil {
			hours := make([]RoutineHour, 24)

			for i := range hours {
				hours[i].Start = first.Add(time.Duration(i) * time.Hour)
			}

			routines[name] = &Routine{Name: name, Hours: hours}
		}

		return routines[name]
	}

	count := func(name, outcome string, hour int, value float64) {
		passes := int(math.Round(value))

		if passes == 0 || hour < 0 || hour > 23 {
			return
		}

		if outcome == "failure" {
			routine(name).Hours[hour].Failed += passes
		} else {
			routine(name).Hours[hour].Passed += passes
		}
	}

	for _, series := range hourly {
		for _, point := range series.Points {
			count(series.Labels["loop"], series.Labels["outcome"], int(point.Time.Add(-time.Hour).Sub(first)/time.Hour), point.Value)
		}
	}

	for _, sample := range currentHour {
		count(sample.Labels["loop"], sample.Labels["outcome"], 23, sample.Value)
	}

	for _, sample := range lastPassed {
		at := time.Unix(int64(sample.Value), 0)
		routine(sample.Labels["loop"]).LastPassedAt = &at
	}

	before := map[string]float64{}

	for _, sample := range averageBefore {
		before[sample.Labels["loop"]] = sample.Value
	}

	for _, sample := range average {
		if math.IsNaN(sample.Value) {
			continue
		}

		target := routine(sample.Labels["loop"])
		duration := time.Duration(sample.Value * float64(time.Second))
		target.AverageDuration = &duration

		if previous := before[sample.Labels["loop"]]; previous > 0 && !math.IsNaN(previous) {
			change := sample.Value/previous - 1
			target.AverageDurationChange = &change
		}
	}

	result := make([]Routine, 0, len(routines))

	for _, routine := range routines {
		result = append(result, *routine)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func wrap(what string, err error) error {
	if err != nil {
		return fmt.Errorf("query %s: %w", what, err)
	}

	return nil
}
