package mcpserver

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getLogsInput struct {
	Kind      string `json:"kind" jsonschema:"one of build, deploy, runtime, scan"`
	Service   string `json:"service,omitempty" jsonschema:"service id for kind=runtime, e.g. acme/shop/development/web"`
	ID        string `json:"id,omitempty" jsonschema:"build, deploy or scan id from get_deploy_status for those kinds, e.g. acme/build-x7k2pq"`
	TailLines int    `json:"tail_lines,omitempty" jsonschema:"maximum number of trailing lines to return (default 200); the result also stops at about 40,000 characters, keeping the newest lines"`
	Since     string `json:"since,omitempty" jsonschema:"runtime only: how far back to read, as a duration such as 15m or 6h (default 1h, at most 336h)"`
	Search    string `json:"search,omitempty" jsonschema:"runtime only: only lines containing this text, ignoring case"`
	Instance  string `json:"instance,omitempty" jsonschema:"runtime only: only entries of this instance, the name in brackets in earlier output"`
}

const maxRuntimeLookback = 14 * 24 * time.Hour

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
		service := input.Service
		if service == "" {
			service = input.ID
		}
		id, err := s.serviceID(service)
		if err != nil {
			return nil, nil, err
		}
		return s.runtimeLogs(ctx, id, input, tail)

	default:
		return nil, nil, fmt.Errorf("invalid kind %q — use one of build, deploy, runtime, scan", input.Kind)
	}

	lines, complete, err := s.subscribeLogs(ctx, query, variables, selector, tail)
	if err != nil {
		return nil, nil, wrapGraphQL("get_logs", err)
	}

	lines, left, shortened := fitLogLines(lines)
	result := map[string]any{
		"lines":    lines,
		"complete": complete,
	}
	var notes []string
	if !complete {
		notes = append(notes, "stream sampled until an idle pause; call again to continue")
	}
	if note := fitNote(left, shortened, "pass a smaller tail_lines to see fewer"); note != "" {
		notes = append(notes, note)
	}
	if len(notes) > 0 {
		result["note"] = strings.Join(notes, ". ")
	}
	return jsonResult(result)
}

const (
	maxLogCharacters     = 40000
	maxLogLineCharacters = 2000
)

func fitLogLines(lines []string) ([]string, int, int) {
	kept := make([]string, 0, len(lines))
	total := 0
	shortened := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if runes := []rune(line); len(runes) > maxLogLineCharacters {
			line = string(runes[:maxLogLineCharacters]) + fmt.Sprintf(" … [%d characters cut]", len(runes)-maxLogLineCharacters)
			shortened++
		}
		if total+len(line) > maxLogCharacters && len(kept) > 0 {
			slices.Reverse(kept)
			return kept, i + 1, shortened
		}
		total += len(line)
		kept = append(kept, line)
	}
	slices.Reverse(kept)
	return kept, 0, shortened
}

func fitNote(left, shortened int, narrow string) string {
	var parts []string
	if left > 0 {
		parts = append(parts, fmt.Sprintf("%d older lines left out to keep this result small; %s", left, narrow))
	}
	if shortened > 0 {
		parts = append(parts, fmt.Sprintf("%d long lines cut to %d characters", shortened, maxLogLineCharacters))
	}
	return strings.Join(parts, ". ")
}

func (s *server) runtimeLogs(ctx context.Context, service string, input getLogsInput, tail int) (*mcp.CallToolResult, any, error) {
	since := time.Hour
	if input.Since != "" {
		parsed, err := time.ParseDuration(input.Since)
		if err != nil || parsed <= 0 {
			return nil, nil, fmt.Errorf("invalid since %q — use a duration such as 15m or 6h", input.Since)
		}
		since = min(parsed, maxRuntimeLookback)
	}

	redactor, err := s.logRedactor(ctx, service)
	if err != nil {
		return nil, nil, fmt.Errorf("runtime logs withheld because the values to redact could not be loaded: %w", err)
	}

	filter := map[string]any{}
	if input.Search != "" {
		filter["search"] = input.Search
	}
	if input.Instance != "" {
		filter["pod"] = input.Instance
	}

	const query = `query($service: ServiceID!, $from: Time!, $filter: ServiceLogFilter, $limit: Int) {
  serviceLogs(service: $service, from: $from, filter: $filter, limit: $limit) { time line pod event }
}`
	var out struct {
		ServiceLogs []struct {
			Time  string  `json:"time"`
			Line  string  `json:"line"`
			Pod   string  `json:"pod"`
			Event *string `json:"event"`
		} `json:"serviceLogs"`
	}
	variables := map[string]any{
		"service": service,
		"from":    time.Now().Add(-since).UTC().Format(time.RFC3339),
		"filter":  filter,
		"limit":   min(tail, 5000),
	}
	if err := s.query(ctx, "get_logs", query, variables, &out); err != nil {
		return nil, nil, err
	}

	lines := make([]string, 0, len(out.ServiceLogs))
	for _, entry := range out.ServiceLogs {
		text := entry.Line
		if entry.Event != nil {
			text = "platform: " + strings.ToLower(strings.ReplaceAll(*entry.Event, "_", " "))
			if entry.Line != "" {
				text += ": " + entry.Line
			}
		}
		lines = append(lines, entry.Time+" ["+entry.Pod+"] "+redactor.Replace(text))
	}

	lines, left, shortened := fitLogLines(lines)
	note := fmt.Sprintf("at most %d entries from the past %s, oldest first, including instances that crashed or were replaced. Entries starting with \"platform:\" are what the platform did with an instance. Credentials and secret-looking variable values appear as [redacted:<source>]", tail, since)
	if fitted := fitNote(left, shortened, "narrow with since, search or instance, or a smaller tail_lines"); fitted != "" {
		note += ". " + fitted
	}
	return jsonResult(map[string]any{
		"lines":    lines,
		"complete": true,
		"note":     note,
	})
}

func scalarLine(field string) func(map[string]any) []string {
	return func(payload map[string]any) []string {
		value, ok := payload[field]
		if !ok {
			return nil
		}
		if text, ok := value.(string); ok {
			return []string{lastRedraw(text)}
		}
		return nil
	}
}

func lastRedraw(line string) string {
	line = strings.TrimRight(line, "\r")
	if i := strings.LastIndex(line, "\r"); i >= 0 {
		return line[i+1:]
	}
	return line
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
		if secretLiteral(v.Key, v.literal()) {
			add(v.Key, v.literal())
		}
	}
	for _, v := range sources.SharedVariables {
		if secretLiteral(v.Key, v.literal()) {
			add(v.Key, v.literal())
		}
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

var secretKeyWords = []string{"SECRET", "PASS", "TOKEN", "KEY", "PRIVATE", "CREDENTIAL", "AUTH", "SALT", "SIGN", "COOKIE", "SESSION", "DSN"}

func secretLiteral(key, value string) bool {
	upper := strings.ToUpper(key)
	for _, word := range secretKeyWords {
		if strings.Contains(upper, word) {
			return true
		}
	}
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		return err != nil || parsed.User != nil
	}
	return looksRandom(value)
}

func looksRandom(value string) bool {
	if len(value) < 16 || strings.ContainsAny(value, " \t\n") {
		return false
	}
	letter := strings.ContainsFunc(value, unicode.IsLetter)
	digit := strings.ContainsFunc(value, unicode.IsDigit)
	return letter && digit
}

func provisioning(err error) bool {
	return strings.Contains(err.Error(), "is provisioning")
}
