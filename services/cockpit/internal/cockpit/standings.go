package cockpit

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// WorkspaceStanding is how many workspaces were in each billing standing at At.
type WorkspaceStanding struct {
	At      time.Time
	Good    int
	Billing int
	Abuse   int
}

// WorkspacePlan is how many workspaces were on each plan at At. None has no
// subscription, Trial has one without a plan.
type WorkspacePlan struct {
	At    time.Time
	None  int
	Trial int
	Hobby int
	Pro   int
}

type workspaceCount struct {
	at     time.Time
	values map[string]int
}

// WorkspaceStandings returns the workspaces per standing at every full hour
// after since and once more now, oldest first, reaching back at most 100 days.
// Hours the metrics don't cover are left out.
func (c *Client) WorkspaceStandings(ctx context.Context, since time.Time) ([]WorkspaceStanding, error) {
	counts, err := c.workspaceCounts(ctx, since, "standing")

	if err != nil {
		return nil, err
	}

	result := make([]WorkspaceStanding, 0, len(counts))

	for _, count := range counts {
		result = append(result, WorkspaceStanding{
			At:      count.at,
			Good:    count.values["good"],
			Billing: count.values["billing"],
			Abuse:   count.values["abuse"],
		})
	}

	return result, nil
}

// WorkspacePlans returns the workspaces per plan at every full hour after
// since and once more now, oldest first, reaching back at most 100 days.
// Hours the metrics don't cover are left out.
func (c *Client) WorkspacePlans(ctx context.Context, since time.Time) ([]WorkspacePlan, error) {
	counts, err := c.workspaceCounts(ctx, since, "plan")

	if err != nil {
		return nil, err
	}

	result := make([]WorkspacePlan, 0, len(counts))

	for _, count := range counts {
		result = append(result, WorkspacePlan{
			At:    count.at,
			None:  count.values["none"],
			Trial: count.values["trial"],
			Hobby: count.values["hobby"],
			Pro:   count.values["pro"],
		})
	}

	return result, nil
}

func (c *Client) workspaceCounts(ctx context.Context, since time.Time, label string) ([]workspaceCount, error) {
	start, current, now := hours(since)
	query := fmt.Sprintf("sum by (%s) (lucity_workspaces)", label)
	byTime := map[time.Time]map[string]int{}

	add := func(at time.Time, value string, count float64) {
		if byTime[at] == nil {
			byTime[at] = map[string]int{}
		}

		byTime[at][value] += int(math.Round(count))
	}

	if start.Before(current) {
		series, err := c.metrics.QueryRange(ctx, query, start.Add(time.Hour), current, time.Hour)

		if err != nil {
			return nil, fmt.Errorf("query workspaces by %s: %w", label, err)
		}

		for _, s := range series {
			for _, point := range s.Points {
				add(point.Time, s.Labels[label], point.Value)
			}
		}
	}

	samples, err := c.metrics.QueryVector(ctx, query, now)

	if err != nil {
		return nil, fmt.Errorf("query workspaces by %s now: %w", label, err)
	}

	for _, sample := range samples {
		add(now.UTC(), sample.Labels[label], sample.Value)
	}

	result := make([]workspaceCount, 0, len(byTime))

	for at, values := range byTime {
		result = append(result, workspaceCount{at: at, values: values})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].at.Before(result[j].at)
	})

	return result, nil
}
