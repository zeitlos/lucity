package mcpserver

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

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
		redactor  *strings.Replacer
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
		redactor, err = s.logRedactor(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("runtime logs withheld because the values to redact could not be loaded: %w", err)
		}
		query = `subscription($service: ServiceID!, $tail: Int) { serviceLogs(service: $service, tailLines: $tail) { line pod } }`
		variables = map[string]any{"service": id, "tail": tail}
		selector = serviceLogLine
		note = "live stream sampled for a few seconds — runtime logs never complete, so this returns after an idle pause. Variable values and credentials in it appear as [redacted:<source>]"

	default:
		return nil, nil, fmt.Errorf("invalid kind %q — use one of build, deploy, runtime, scan", input.Kind)
	}

	lines, complete, err := s.subscribeLogs(ctx, query, variables, selector, tail)
	if err != nil {
		return nil, nil, wrapGraphQL("get_logs", err)
	}
	if redactor != nil {
		for i, line := range lines {
			lines[i] = redactor.Replace(line)
		}
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

const minimumRedactedLength = 8

type environmentResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type endpointCredentials struct {
	Password string `json:"password"`
	URI      string `json:"uri"`
}

func (s *server) logRedactor(ctx context.Context, serviceID string) (*strings.Replacer, error) {
	const operation = "get_logs (redaction)"

	const sourcesQuery = `query($service: ServiceID!, $environment: EnvironmentID!) {
  serviceVariables(service: $service) { key value }
  sharedVariables(environment: $environment) { key value }
  environment(environment: $environment) {
    databases { id name }
    keyValueStores { id name }
    buckets { id name }
  }
}`
	var sources struct {
		ServiceVariables []variable `json:"serviceVariables"`
		SharedVariables  []variable `json:"sharedVariables"`
		Environment      struct {
			Databases      []environmentResource `json:"databases"`
			KeyValueStores []environmentResource `json:"keyValueStores"`
			Buckets        []environmentResource `json:"buckets"`
		} `json:"environment"`
	}
	if err := s.query(ctx, operation, sourcesQuery, map[string]any{"service": serviceID, "environment": environmentOfService(serviceID)}, &sources); err != nil {
		return nil, err
	}

	secrets := map[string]string{}
	add := func(label, value string) {
		candidates := []string{value}
		if strings.Contains(value, "\n") {
			candidates = append(candidates, strings.Split(value, "\n")...)
		}
		for _, candidate := range candidates {
			candidate = strings.TrimSpace(candidate)
			if len(candidate) >= minimumRedactedLength {
				secrets[candidate] = label
			}
		}
	}

	for _, v := range sources.ServiceVariables {
		add(v.Key, v.literal())
	}
	for _, v := range sources.SharedVariables {
		add(v.Key, v.literal())
	}

	for _, database := range sources.Environment.Databases {
		const query = `query($database: DatabaseID!) { databaseCredentials(database: $database) { password uri } }`
		var out struct {
			DatabaseCredentials []endpointCredentials `json:"databaseCredentials"`
		}
		if err := s.query(ctx, operation, query, map[string]any{"database": database.ID}, &out); err != nil {
			if provisioning(err) {
				continue
			}
			return nil, err
		}
		for _, endpoint := range out.DatabaseCredentials {
			add(database.Name+" password", endpoint.Password)
			add(database.Name+" uri", endpoint.URI)
		}
	}

	for _, store := range sources.Environment.KeyValueStores {
		const query = `query($keyValueStore: KeyValueStoreID!) { keyValueStoreCredentials(keyValueStore: $keyValueStore) { password uri } }`
		var out struct {
			KeyValueStoreCredentials []endpointCredentials `json:"keyValueStoreCredentials"`
		}
		if err := s.query(ctx, operation, query, map[string]any{"keyValueStore": store.ID}, &out); err != nil {
			if provisioning(err) {
				continue
			}
			return nil, err
		}
		for _, endpoint := range out.KeyValueStoreCredentials {
			add(store.Name+" password", endpoint.Password)
			add(store.Name+" uri", endpoint.URI)
		}
	}

	for _, bucket := range sources.Environment.Buckets {
		const query = `query($bucket: BucketID!) { bucketCredentials(bucket: $bucket) { accessKeyId secretAccessKey } }`
		var out struct {
			BucketCredentials struct {
				AccessKeyID     string `json:"accessKeyId"`
				SecretAccessKey string `json:"secretAccessKey"`
			} `json:"bucketCredentials"`
		}
		if err := s.query(ctx, operation, query, map[string]any{"bucket": bucket.ID}, &out); err != nil {
			return nil, err
		}
		add(bucket.Name+" access key id", out.BucketCredentials.AccessKeyID)
		add(bucket.Name+" secret access key", out.BucketCredentials.SecretAccessKey)
	}

	values := slices.SortedFunc(maps.Keys(secrets), func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	replacements := make([]string, 0, 2*len(values))
	for _, value := range values {
		replacements = append(replacements, value, "[redacted:"+secrets[value]+"]")
	}
	return strings.NewReplacer(replacements...), nil
}

func provisioning(err error) bool {
	return strings.Contains(err.Error(), "is provisioning")
}
