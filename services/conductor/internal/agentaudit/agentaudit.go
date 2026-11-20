// Package agentaudit reads the calls agents made through the agent gateway
// from the gateway's access log in VictoriaLogs.
package agentaudit

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/zeitlos/lucity/pkg/platform"
)

type Kind string

const (
	Model Kind = "model"
	Tool  Kind = "tool"
)

type Result string

const (
	OK     Result = "ok"
	Denied Result = "denied"
	Failed Result = "failed"
)

type Entry struct {
	Time          time.Time
	Kind          Kind
	ModelProvider string
	Connector     string
	Target        string
	InputTokens   *int
	OutputTokens  *int
	Result        Result
}

const (
	Limit       = 500
	maxBody     = 16 << 20
	maxLineSize = 1 << 20
	maxTarget   = 256
)

var errDisabled = errors.New("agent audit is not configured")

type Client struct {
	url       string
	namespace string
	http      *http.Client
}

// New returns a client for the VictoriaLogs at url, reading the log of the
// agent gateway in namespace. An empty url disables the audit.
func New(url, namespace string) *Client {
	return &Client{
		url:       strings.TrimSuffix(url, "/"),
		namespace: namespace,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

type line struct {
	Time         string `json:"_time"`
	Route        string `json:"route"`
	Status       string `json:"http.status"`
	Reason       string `json:"reason"`
	Model        string `json:"gen_ai.request.model"`
	InputTokens  string `json:"gen_ai.usage.input_tokens"`
	OutputTokens string `json:"gen_ai.usage.output_tokens"`
	Method       string `json:"mcp.method.name"`
	Tool         string `json:"gen_ai.tool.name"`
}

// Entries returns an agent's model and tool calls since since, newest first.
// A tool call is a call of a connector's tool, or any refused connector
// request. Without kinds, every kind is returned.
func (c *Client) Entries(ctx context.Context, agent platform.AgentID, kinds []Kind, since time.Time) ([]Entry, error) {
	if c.url == "" {
		return nil, errDisabled
	}

	routes := []string{}

	if len(kinds) == 0 || containsKind(kinds, Model) {
		routes = append(routes, `route:~"\\.model\\."`)
	}

	if len(kinds) == 0 || containsKind(kinds, Tool) {
		routes = append(routes, `(route:~"\\.connector\\." and (mcp.method.name:="tools/call" or http.status:>=400))`)
	}

	query := fmt.Sprintf(
		`_time:>=%s {k8s.namespace.name=%s} "lucity.agent" | unpack_json from _msg fields (route, http.status, reason, lucity.workspace, lucity.agent, gen_ai.request.model, gen_ai.usage.input_tokens, gen_ai.usage.output_tokens, mcp.method.name, gen_ai.tool.name) | filter lucity.workspace:=%s lucity.agent:=%s | filter %s | sort by (_time desc) | limit %d | fields _time, route, http.status, reason, gen_ai.request.model, gen_ai.usage.input_tokens, gen_ai.usage.output_tokens, mcp.method.name, gen_ai.tool.name`,
		since.UTC().Format(time.RFC3339),
		strconv.Quote(c.namespace),
		strconv.Quote(agent.Workspace),
		strconv.Quote(agent.String()),
		strings.Join(routes, " or "),
		Limit,
	)

	form := url.Values{"query": {query}, "limit": {strconv.Itoa(Limit)}}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/select/logsql/query", strings.NewReader(form.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)

	if err != nil {
		return nil, fmt.Errorf("query agent audit: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

		return nil, fmt.Errorf("query agent audit: HTTP %d: %s", resp.StatusCode, body)
	}

	scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxBody))
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	var entries []Entry

	for scanner.Scan() && len(entries) < Limit {
		var raw line

		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		if entry, ok := convert(raw, agent.Workspace); ok {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read agent audit: %w", err)
	}

	return entries, nil
}

func convert(raw line, workspace string) (Entry, bool) {
	at, err := time.Parse(time.RFC3339Nano, raw.Time)

	if err != nil {
		return Entry{}, false
	}

	_, route, _ := strings.Cut(raw.Route, "/")
	prefix, name, ok := strings.Cut(route, ".")

	if !ok || prefix != workspace {
		return Entry{}, false
	}

	entry := Entry{Time: at, Result: result(raw.Status, raw.Reason)}

	if provider, ok := strings.CutPrefix(name, "model."); ok {
		entry.Kind = Model
		entry.ModelProvider = provider
		entry.Target = clip(raw.Model)
		entry.InputTokens = tokens(raw.InputTokens)
		entry.OutputTokens = tokens(raw.OutputTokens)
	} else if connector, ok := strings.CutPrefix(name, "connector."); ok {
		entry.Kind = Tool
		entry.Connector = connector
		entry.Target = clip(raw.Tool)

		if entry.Target == "" {
			entry.Target = clip(raw.Method)
		}
	} else {
		return Entry{}, false
	}

	if entry.Target == "" {
		entry.Target = "unknown"
	}

	return entry, true
}

func result(status, reason string) Result {
	code, _ := strconv.Atoi(status)

	switch {
	case reason == "Authorization" || reason == "Authentication" || code == http.StatusUnauthorized || code == http.StatusForbidden:
		return Denied
	case code >= 200 && code < 400:
		return OK
	default:
		return Failed
	}
}

func tokens(raw string) *int {
	value, err := strconv.Atoi(raw)

	if err != nil || value <= 0 {
		return nil
	}

	return &value
}

func clip(s string) string {
	if len(s) > maxTarget {
		return s[:maxTarget]
	}

	return s
}

func containsKind(kinds []Kind, kind Kind) bool {
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}

	return false
}

// Disabled reports whether err means the audit is not configured.
func Disabled(err error) bool {
	return errors.Is(err, errDisabled)
}
