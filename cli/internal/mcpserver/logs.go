package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getLogsInput struct {
	Kind      string `json:"kind" jsonschema:"one of build, deploy, runtime, scan"`
	ID        string `json:"id" jsonschema:"build/deploy/scan job id for those kinds, or a service id for runtime"`
	TailLines int    `json:"tail_lines,omitempty" jsonschema:"maximum number of trailing lines to return (default 200)"`
}

func (s *server) getLogs(ctx context.Context, _ *mcp.CallToolRequest, input getLogsInput) (*mcp.CallToolResult, any, error) {
	if _, err := s.requireWorkspace(); err != nil {
		return nil, nil, err
	}

	tail := input.TailLines
	if tail <= 0 {
		tail = 200
	}

	var (
		query     string
		variables map[string]any
		selector  func(payload map[string]any) []string
		note      string
	)

	switch input.Kind {
	case "build":
		id, err := s.scopedID(input.ID, 2)
		if err != nil {
			return nil, nil, err
		}
		query = `subscription($id: BuildID!) { buildLogs(id: $id) }`
		variables = map[string]any{"id": id}
		selector = scalarLine("buildLogs")

	case "deploy":
		id, err := s.scopedID(input.ID, 2)
		if err != nil {
			return nil, nil, err
		}
		query = `subscription($id: DeployID!) { deployLogs(id: $id) }`
		variables = map[string]any{"id": id}
		selector = scalarLine("deployLogs")

	case "scan":
		id, err := s.scopedID(input.ID, 2)
		if err != nil {
			return nil, nil, err
		}
		query = `subscription($id: ScanID!) { scanLogs(id: $id) }`
		variables = map[string]any{"id": id}
		selector = scalarLine("scanLogs")

	case "runtime":
		id, err := s.serviceID(input.ID)
		if err != nil {
			return nil, nil, err
		}
		query = `subscription($service: ServiceID!, $tail: Int) { serviceLogs(service: $service, tailLines: $tail) { line pod } }`
		variables = map[string]any{"service": id, "tail": tail}
		selector = serviceLogLine
		note = "live stream sampled for a few seconds — runtime logs never complete, so this returns after an idle pause"

	default:
		return nil, nil, fmt.Errorf("invalid kind %q — use one of build, deploy, runtime, scan", input.Kind)
	}

	lines, complete, err := s.subscribeLogs(ctx, query, variables, selector, tail)
	if err != nil {
		return nil, nil, wrapGraphQL("get_logs", err)
	}

	result := map[string]any{
		"lines":    lines,
		"complete": complete,
	}
	if note != "" {
		result["note"] = note
	} else if !complete {
		result["note"] = "stream sampled until an idle pause; call again to continue"
	}
	return jsonResult(result)
}

func scalarLine(field string) func(map[string]any) []string {
	return func(payload map[string]any) []string {
		value, ok := payload[field]
		if !ok {
			return nil
		}
		if text, ok := value.(string); ok {
			return []string{text}
		}
		return nil
	}
}

func serviceLogLine(payload map[string]any) []string {
	entry, ok := payload["serviceLogs"].(map[string]any)
	if !ok {
		return nil
	}
	line, _ := entry["line"].(string)
	pod, _ := entry["pod"].(string)
	if pod != "" {
		return []string{"[" + pod + "] " + line}
	}
	return []string{line}
}
