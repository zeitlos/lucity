package cockpit

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// LogLevel is a slog level the platform logs are counted by.
type LogLevel string

const (
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
)

// PlatformLogs counts the errors and warnings of the platform services.
type PlatformLogs struct {
	// Hours holds the last 24 hours, oldest first. The last hour is the
	// current one, so far.
	Hours       []LogHour
	Errors24h   int
	Errors7d    int
	Warnings24h int
	Warnings7d  int
	Messages    []LogMessage
}

type LogHour struct {
	Start    time.Time
	Errors   int
	Warnings int
}

// LogMessage is one message of one service at one level. FirstSeenAt reaches
// back at most 30 days.
type LogMessage struct {
	Level       LogLevel
	Message     string
	Service     string
	Count24h    int
	Count7d     int
	FirstSeenAt time.Time
}

const (
	logMessages = 20

	platformLogStream = `_stream:{k8s.namespace.name="lucity-system",k8s.deployment.name=~"lucity-(conductor|stagehand|cashier|exporter)"}`
)

// PlatformLogs returns the errors and warnings of the platform services, with
// the messages seen over the last 7 days, most frequent in the last 24 hours
// first.
func (c *Client) PlatformLogs(ctx context.Context) (PlatformLogs, error) {
	if c.logs == nil {
		return PlatformLogs{}, errors.New("the logs store wasn't reachable when the cockpit connected")
	}

	now := time.Now()
	first := now.Truncate(time.Hour).Add(-23 * time.Hour)
	result := PlatformLogs{Hours: make([]LogHour, 24)}

	for i := range result.Hours {
		result.Hours[i].Start = first.Add(time.Duration(i) * time.Hour)
	}

	var rows []map[string]string

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		query := platformLogStream + ` | unpack_json from _msg fields (level) | filter level:in(ERROR, WARN) | stats by (level) count() hits`
		series, err := c.logs.StatsRange(ctx, query, first, now, time.Hour)

		if err != nil {
			return fmt.Errorf("query log levels per hour: %w", err)
		}

		for _, s := range series {
			for _, point := range s.Points {
				hour := int(point.Time.Sub(first) / time.Hour)

				if hour < 0 || hour > 23 {
					continue
				}

				if LogLevel(s.Labels["level"]) == LogLevelError {
					result.Hours[hour].Errors += int(math.Round(point.Value))
				} else {
					result.Hours[hour].Warnings += int(math.Round(point.Value))
				}
			}
		}

		return nil
	})

	group.Go(func() error {
		query := `_time:30d ` + platformLogStream + ` | unpack_json from _msg fields (level, msg) | filter level:in(ERROR, WARN) | stats by (k8s.deployment.name, level, msg) count() if (_time:24h) day, count() if (_time:7d) week, min(_time) first | filter week:>0`

		var err error
		rows, err = c.logs.Query(ctx, query)

		return wrap("log messages", err)
	})

	if err := group.Wait(); err != nil {
		return PlatformLogs{}, err
	}

	result.Messages = []LogMessage{}

	for _, row := range rows {
		day, _ := strconv.Atoi(row["day"])
		week, _ := strconv.Atoi(row["week"])
		firstSeen, _ := time.Parse(time.RFC3339Nano, row["first"])
		level := LogLevel(row["level"])

		if level == LogLevelError {
			result.Errors24h += day
			result.Errors7d += week
		} else {
			result.Warnings24h += day
			result.Warnings7d += week
		}

		result.Messages = append(result.Messages, LogMessage{
			Level:       level,
			Message:     row["msg"],
			Service:     strings.TrimPrefix(row["k8s.deployment.name"], "lucity-"),
			Count24h:    day,
			Count7d:     week,
			FirstSeenAt: firstSeen,
		})
	}

	sort.Slice(result.Messages, func(i, j int) bool {
		left := result.Messages[i]
		right := result.Messages[j]

		if left.Count24h != right.Count24h {
			return left.Count24h > right.Count24h
		}

		return left.Count7d > right.Count7d
	})

	result.Messages = result.Messages[:min(len(result.Messages), logMessages)]

	return result, nil
}
