package cockpit

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

const standingsQuery = "sum by (standing) (lucity_workspaces)"

// WorkspaceStanding is how many workspaces were in each billing standing at At.
type WorkspaceStanding struct {
	At      time.Time
	Good    int
	Billing int
	Abuse   int
}

// WorkspaceStandings returns the workspaces per standing at every full hour
// after since and once more now, oldest first, reaching back at most 100 days.
// Hours the metrics don't cover are left out.
func (c *Client) WorkspaceStandings(ctx context.Context, since time.Time) ([]WorkspaceStanding, error) {
	start, current, now := hours(since)
	byTime := map[time.Time]*WorkspaceStanding{}

	add := func(at time.Time, standing string, value float64) {
		entry, ok := byTime[at]

		if !ok {
			entry = &WorkspaceStanding{At: at}
			byTime[at] = entry
		}

		count := int(math.Round(value))

		switch standing {
		case "good":
			entry.Good += count
		case "billing":
			entry.Billing += count
		case "abuse":
			entry.Abuse += count
		}
	}

	if start.Before(current) {
		series, err := c.metrics.QueryRange(ctx, standingsQuery, start.Add(time.Hour), current, time.Hour)

		if err != nil {
			return nil, fmt.Errorf("query workspace standings: %w", err)
		}

		for _, s := range series {
			for _, point := range s.Points {
				add(point.Time, s.Labels["standing"], point.Value)
			}
		}
	}

	samples, err := c.metrics.QueryVector(ctx, standingsQuery, now)

	if err != nil {
		return nil, fmt.Errorf("query workspace standings now: %w", err)
	}

	for _, sample := range samples {
		add(now.UTC(), sample.Labels["standing"], sample.Value)
	}

	result := make([]WorkspaceStanding, 0, len(byTime))

	for _, entry := range byTime {
		result = append(result, *entry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].At.Before(result[j].At)
	})

	return result, nil
}
