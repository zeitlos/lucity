package victorialogs

import (
	"bufio"
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

func New(baseURL string) *Client {
	return &Client{
		url:  strings.TrimRight(baseURL, "/"),
		http: &http.Client{Timeout: time.Minute},
	}
}

type Point struct {
	Time  time.Time
	Value float64
}

type Series struct {
	Labels map[string]string
	Points []Point
}

// Query runs a LogsQL query and returns every result row as its fields.
func (c *Client) Query(ctx context.Context, query string) ([]map[string]string, error) {
	form := url.Values{}
	form.Set("query", query)

	body, err := c.post(ctx, "/select/logsql/query", form)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	var rows []map[string]string

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}

		row := map[string]string{}

		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}

		rows = append(rows, row)
	}

	return rows, scanner.Err()
}

// StatsRange runs a LogsQL stats query over buckets of step between start
// and end.
func (c *Client) StatsRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]Series, error) {
	form := url.Values{}
	form.Set("query", query)
	form.Set("start", strconv.FormatInt(start.Unix(), 10))
	form.Set("end", strconv.FormatInt(end.Unix(), 10))
	form.Set("step", strconv.FormatInt(int64(step.Seconds()), 10)+"s")

	body, err := c.post(ctx, "/select/logsql/stats_query_range", form)

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
		return nil, fmt.Errorf("parse stats: %w", err)
	}

	result := make([]Series, 0, len(parsed.Data.Result))

	for _, r := range parsed.Data.Result {
		points := make([]Point, 0, len(r.Values))

		for _, v := range r.Values {
			at, ok := v[0].(float64)

			if !ok {
				continue
			}

			raw, ok := v[1].(string)

			if !ok {
				continue
			}

			value, err := strconv.ParseFloat(raw, 64)

			if err != nil {
				continue
			}

			points = append(points, Point{Time: time.Unix(int64(at), 0), Value: value})
		}

		result = append(result, Series{Labels: r.Metric, Points: points})
	}

	return result, nil
}

func (c *Client) post(ctx context.Context, path string, form url.Values) (io.ReadCloser, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+path, strings.NewReader(form.Encode()))

	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.http.Do(request)

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
