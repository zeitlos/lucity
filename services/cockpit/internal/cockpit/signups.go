package cockpit

import (
	"context"
	"fmt"
	"math"
	"time"
)

const history = 100 * 24 * time.Hour

// SignupHour is the number of signups in the hour that starts at Start.
type SignupHour struct {
	Start time.Time
	Count int
}

// Signups returns the signups per hour from the hour of since until now,
// oldest first, reaching back at most 100 days. Hours the metrics don't cover
// are left out.
func (c *Client) Signups(ctx context.Context, since time.Time) ([]SignupHour, error) {
	start, current, now := hours(since)

	var result []SignupHour

	if start.Before(current) {
		series, err := c.metrics.QueryRange(ctx, "sum(increase(lucity_signups_total[1h]))", start.Add(time.Hour), current, time.Hour)

		if err != nil {
			return nil, fmt.Errorf("query signups: %w", err)
		}

		for _, s := range series {
			for _, point := range s.Points {
				result = append(result, SignupHour{Start: point.Time.Add(-time.Hour), Count: int(math.Round(point.Value))})
			}
		}
	}

	query := fmt.Sprintf("sum(increase(lucity_signups_total[%ds]))", max(int(now.Sub(current).Seconds()), 1))
	samples, err := c.metrics.QueryVector(ctx, query, now)

	if err != nil {
		return nil, fmt.Errorf("query signups this hour: %w", err)
	}

	for _, sample := range samples {
		result = append(result, SignupHour{Start: current, Count: int(math.Round(sample.Value))})
	}

	return result, nil
}

// hours returns the start of the hour of since, clamped to the history the
// cockpit reads, the start of the current hour and now.
func hours(since time.Time) (time.Time, time.Time, time.Time) {
	now := time.Now()

	if earliest := now.Add(-history); since.Before(earliest) {
		since = earliest
	}

	return since.Truncate(time.Hour), now.Truncate(time.Hour), now
}
