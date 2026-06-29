package victoria

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	url  string
	http *http.Client
}

func New(vmURL string) (*Client, error) {
	if vmURL == "" {
		return nil, fmt.Errorf("VictoriaMetrics URL is empty")
	}

	return &Client{
		url:  strings.TrimRight(vmURL, "/"),
		http: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.QueryVector(ctx, "vm_app_uptime_seconds", time.Now())
	return err
}

type Sample struct {
	Labels map[string]string
	Value  float64
}

type Point struct {
	Time  time.Time
	Value float64
}

type Series struct {
	Labels map[string]string
	Points []Point
}

func (c *Client) QueryVector(ctx context.Context, query string, at time.Time) ([]Sample, error) {
	form := url.Values{}
	form.Set("query", query)
	form.Set("time", strconv.FormatInt(at.Unix(), 10))

	body, err := c.post(ctx, "/api/v1/query", form)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]any            `json:"value"`
			} `json:"result"`
		} `json:"data"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics error: %s: %s", parsed.ErrorType, parsed.Error)
	}

	out := make([]Sample, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		s, ok := r.Value[1].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected value type %T in response", r.Value[1])
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %q: %w", s, err)
		}
		out = append(out, Sample{Labels: r.Metric, Value: v})
	}
	return out, nil
}

func (c *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]Series, error) {
	form := url.Values{}
	form.Set("query", query)
	form.Set("start", strconv.FormatInt(start.Unix(), 10))
	form.Set("end", strconv.FormatInt(end.Unix(), 10))
	form.Set("step", strconv.FormatInt(int64(step.Seconds()), 10)+"s")

	body, err := c.post(ctx, "/api/v1/query_range", form)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Values [][2]any          `json:"values"`
			} `json:"result"`
		} `json:"data"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics error: %s: %s", parsed.ErrorType, parsed.Error)
	}

	out := make([]Series, 0, len(parsed.Data.Result))
	for _, r := range parsed.Data.Result {
		points := make([]Point, 0, len(r.Values))
		for _, v := range r.Values {
			ts, ok := v[0].(float64)
			if !ok {
				continue
			}
			raw, ok := v[1].(string)
			if !ok {
				continue
			}
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				continue
			}
			points = append(points, Point{Time: time.Unix(int64(ts), 0).UTC(), Value: f})
		}
		out = append(out, Series{Labels: r.Metric, Points: points})
	}
	return out, nil
}

func (c *Client) post(ctx context.Context, path string, form url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("VictoriaMetrics returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}
