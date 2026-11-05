// Package servicelogs reads the runtime logs of services, and what the
// platform did with their instances, from VictoriaLogs.
package servicelogs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/platform"
	"github.com/zeitlos/lucity/pkg/usererror"
)

type Stream string

const (
	Stdout Stream = "stdout"
	Stderr Stream = "stderr"
)

// Event is something the platform did with an instance.
type Event string

const (
	Started           Event = "started"
	Restarted         Event = "restarted"
	Stopping          Event = "stopping"
	HealthCheckFailed Event = "health_check_failed"
	RestartDelayed    Event = "restart_delayed"
)

// Entry is a line an instance wrote, or an event when Event is set.
type Entry struct {
	Time     time.Time
	Instance string
	Stream   Stream
	Event    Event
	Line     string
}

// Filter narrows entries down. Search leaves out events.
type Filter struct {
	Search   string
	Instance string
}

// Instance is a pod that wrote lines in a range. Restarts are the times it
// started again within the range.
type Instance struct {
	Name      string
	FirstSeen time.Time
	LastSeen  time.Time
	Restarts  []time.Time
}

// Volume counts the lines written in one slot of a range.
type Volume struct {
	Time   time.Time
	Stdout int
	Stderr int
}

const (
	deploymentPrefix = "lucity-app-"
	maxLineSize      = 1 << 20
	tailOverlap      = 10 * time.Second
	entryFields      = "_time, _msg, k8s.pod.name, log.iostream, k8s.object.name, k8s.event.reason, k8s.event.count"

	// missingMessage is what VictoriaLogs stores in place of an empty line.
	missingMessage = "missing _msg field; see https://docs.victoriametrics.com/victorialogs/keyconcepts/#message-field"
)

var errDisabled = usererror.New("Runtime logs aren't available on this platform.")

type Client struct {
	url    string
	http   *http.Client
	stream *http.Client
}

// New returns a client for the VictoriaLogs at url. An empty url disables
// runtime logs.
func New(url string) *Client {
	return &Client{
		url:    strings.TrimSuffix(url, "/"),
		http:   &http.Client{Timeout: time.Minute},
		stream: &http.Client{},
	}
}

// Entries returns the entries between from and to, oldest first. When there
// are more than limit, it returns the newest. With surrounding above zero, every
// line comes with that many lines of the same instance around it.
func (c *Client) Entries(ctx context.Context, service platform.ServiceID, from, to time.Time, filter Filter, surrounding, limit int) ([]Entry, error) {
	query := entrySelector(service, filter)

	if surrounding > 0 {
		query += fmt.Sprintf(" | stream_context before %d after %d", surrounding, surrounding)
	}

	query += fmt.Sprintf(" | sort by (_time desc) | limit %d | fields %s", limit, entryFields)

	rows, err := c.query(ctx, query, from, to)

	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(rows))

	for i := len(rows) - 1; i >= 0; i-- {
		if entry, ok := toEntry(rows[i]); ok {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Instances returns the instances that wrote lines between from and to,
// the one that started last first.
func (c *Client) Instances(ctx context.Context, service platform.ServiceID, from, to time.Time) ([]Instance, error) {
	query := lineSelector(service, Filter{}) + " | stats by (k8s.pod.name, k8s.container.restart_count) min(_time) first, max(_time) last"

	rows, err := c.query(ctx, query, from, to)

	if err != nil {
		return nil, err
	}

	type run struct {
		restart int
		first   time.Time
		last    time.Time
	}

	runs := map[string][]run{}

	for _, row := range rows {
		restart, _ := strconv.Atoi(row["k8s.container.restart_count"])
		first, err := time.Parse(time.RFC3339Nano, row["first"])

		if err != nil {
			continue
		}

		last, err := time.Parse(time.RFC3339Nano, row["last"])

		if err != nil {
			continue
		}

		name := row["k8s.pod.name"]
		runs[name] = append(runs[name], run{restart: restart, first: first, last: last})
	}

	instances := make([]Instance, 0, len(runs))

	for name, pod := range runs {
		slices.SortFunc(pod, func(a, b run) int { return a.restart - b.restart })

		instance := Instance{Name: name, FirstSeen: pod[0].first, LastSeen: pod[0].last, Restarts: []time.Time{}}

		for i, r := range pod {
			if r.first.Before(instance.FirstSeen) {
				instance.FirstSeen = r.first
			}

			if r.last.After(instance.LastSeen) {
				instance.LastSeen = r.last
			}

			if i > 0 {
				instance.Restarts = append(instance.Restarts, r.first)
			}
		}

		instances = append(instances, instance)
	}

	slices.SortFunc(instances, func(a, b Instance) int { return b.FirstSeen.Compare(a.FirstSeen) })

	return instances, nil
}

// Volume counts the lines between from and to in slots of equal length.
func (c *Client) Volume(ctx context.Context, service platform.ServiceID, from, to time.Time, filter Filter, slots int) ([]Volume, error) {
	step := to.Sub(from) / time.Duration(slots)

	if step < time.Second {
		step = time.Second
	}

	query := lineSelector(service, filter) + " | stats by (log.iostream) count() lines"

	series, err := c.statsRange(ctx, query, from, to, step)

	if err != nil {
		return nil, err
	}

	volume := make([]Volume, slots)

	for i := range volume {
		volume[i].Time = from.Add(time.Duration(i) * step)
	}

	for _, s := range series {
		for _, p := range s.points {
			slot := min(max(int(math.Floor(float64(p.time.Sub(from))/float64(step))), 0), slots-1)

			if Stream(s.labels["log.iostream"]) == Stderr {
				volume[slot].Stderr += p.value
			} else {
				volume[slot].Stdout += p.value
			}
		}
	}

	return volume, nil
}

// Tail sends the last lines entries written since from, then every new entry
// until ctx is done or the connection ends.
func (c *Client) Tail(ctx context.Context, service platform.ServiceID, from time.Time, filter Filter, lines int) (<-chan Entry, error) {
	started := time.Now()

	initial, err := c.Entries(ctx, service, from, started, filter, 0, lines)

	if err != nil {
		return nil, err
	}

	since := started.Add(-tailOverlap)
	var last time.Time

	if len(initial) > 0 {
		last = initial[len(initial)-1].Time

		if last.After(since) {
			since = last
		}
	}

	out := make(chan Entry, 128)

	go func() {
		defer close(out)

		for _, entry := range initial {
			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}

		form := url.Values{}
		form.Set("query", entrySelector(service, filter))
		form.Set("start_offset", fmt.Sprintf("%ds", int(math.Ceil(time.Since(since).Seconds()))))

		body, err := c.post(ctx, c.stream, "/select/logsql/tail", form)

		if err != nil {
			if ctx.Err() == nil {
				slog.WarnContext(ctx, "failed to tail service logs", "service", service.String(), "error", err)
			}

			return
		}

		defer body.Close()

		scanner := bufio.NewScanner(body)
		scanner.Buffer(make([]byte, 64*1024), maxLineSize)

		for scanner.Scan() {
			if len(scanner.Bytes()) == 0 {
				continue
			}

			row := map[string]string{}

			if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
				continue
			}

			entry, ok := toEntry(row)

			if !ok || !entry.Time.After(last) {
				continue
			}

			select {
			case out <- entry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func entrySelector(service platform.ServiceID, filter Filter) string {
	lines := lineSelector(service, filter)

	if filter.Search != "" {
		return lines
	}

	return "(" + lines + ") or (" + eventSelector(service, filter) + ")"
}

func lineSelector(service platform.ServiceID, filter Filter) string {
	streams := []string{
		"k8s.namespace.name=" + strconv.Quote(service.Namespace()),
		"k8s.deployment.name=" + strconv.Quote(deploymentPrefix+service.Name),
		"k8s.container.name=" + strconv.Quote(service.Name),
	}

	if filter.Instance != "" {
		streams = append(streams, "k8s.pod.name="+strconv.Quote(filter.Instance))
	}

	selector := "{" + strings.Join(streams, ", ") + "}"

	if filter.Search != "" {
		selector += " ~" + strconv.Quote("(?i)"+regexp.QuoteMeta(filter.Search))
	}

	return selector
}

func eventSelector(service platform.ServiceID, filter Filter) string {
	pods := "k8s.object.name=~" + strconv.Quote("^"+regexp.QuoteMeta(deploymentPrefix+service.Name)+"-[a-z0-9]+-[a-z0-9]+$")

	if filter.Instance != "" {
		pods = "k8s.object.name=" + strconv.Quote(filter.Instance)
	}

	streams := []string{
		`k8s.object.kind="Pod"`,
		"k8s.object.fieldpath=" + strconv.Quote("spec.containers{"+service.Name+"}"),
		pods,
	}

	return "{" + strings.Join(streams, ", ") + "} k8s.namespace.name:=" + strconv.Quote(service.Namespace()) + " k8s.event.reason:in(Started, Killing, Unhealthy, BackOff)"
}

func toEntry(row map[string]string) (Entry, bool) {
	at, err := time.Parse(time.RFC3339Nano, row["_time"])

	if err != nil {
		return Entry{}, false
	}

	reason := row["k8s.event.reason"]

	if reason == "" {
		instance := row["k8s.pod.name"]
		line := row["_msg"]

		if line == missingMessage {
			line = ""
		}

		return Entry{Time: at, Instance: instance, Stream: Stream(row["log.iostream"]), Line: line}, instance != ""
	}

	entry := Entry{Time: at, Instance: row["k8s.object.name"]}

	switch reason {
	case "Started":
		count, _ := strconv.Atoi(row["k8s.event.count"])
		entry.Event = Started

		if count > 1 {
			entry.Event = Restarted
		}
	case "Killing":
		entry.Event = Stopping
	case "Unhealthy":
		entry.Event = HealthCheckFailed
		entry.Line = probeDetail(row["_msg"])
	case "BackOff":
		entry.Event = RestartDelayed
	default:
		return Entry{}, false
	}

	return entry, true
}

func probeDetail(message string) string {
	for _, marker := range []string{" probe failed: ", " probe errored: "} {
		if _, detail, ok := strings.Cut(message, marker); ok {
			return detail
		}
	}

	return message
}

func (c *Client) query(ctx context.Context, query string, from, to time.Time) ([]map[string]string, error) {
	form := url.Values{}
	form.Set("query", query)
	form.Set("start", from.UTC().Format(time.RFC3339Nano))
	form.Set("end", to.UTC().Format(time.RFC3339Nano))

	body, err := c.post(ctx, c.http, "/select/logsql/query", form)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	var rows []map[string]string

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}

		row := map[string]string{}

		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("parse log row: %w", err)
		}

		rows = append(rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read logs: %w", err)
	}

	return rows, nil
}

type point struct {
	time  time.Time
	value int
}

type series struct {
	labels map[string]string
	points []point
}

func (c *Client) statsRange(ctx context.Context, query string, from, to time.Time, step time.Duration) ([]series, error) {
	form := url.Values{}
	form.Set("query", query)
	form.Set("start", from.UTC().Format(time.RFC3339Nano))
	form.Set("end", to.UTC().Format(time.RFC3339Nano))
	form.Set("step", fmt.Sprintf("%ds", int64(step.Seconds())))

	body, err := c.post(ctx, c.http, "/select/logsql/stats_query_range", form)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	var parsed struct {
		Data struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][2]any          `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("parse log stats: %w", err)
	}

	result := make([]series, 0, len(parsed.Data.Result))

	for _, r := range parsed.Data.Result {
		points := make([]point, 0, len(r.Values))

		for _, v := range r.Values {
			at, ok := v[0].(float64)

			if !ok {
				continue
			}

			raw, ok := v[1].(string)

			if !ok {
				continue
			}

			value, err := strconv.Atoi(raw)

			if err != nil {
				continue
			}

			points = append(points, point{time: time.Unix(int64(at), 0), value: value})
		}

		result = append(result, series{labels: r.Metric, points: points})
	}

	return result, nil
}

func (c *Client) post(ctx context.Context, client *http.Client, path string, form url.Values) (io.ReadCloser, error) {
	if c.url == "" {
		return nil, errDisabled
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+path, strings.NewReader(form.Encode()))

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("query logs: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()

		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))

		return nil, fmt.Errorf("logs returned %d: %s", response.StatusCode, strings.TrimSpace(string(message)))
	}

	return response.Body, nil
}
