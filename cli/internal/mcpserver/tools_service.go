package mcpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serviceSummaryFields = `id name status replicas { desired ready } port command resources { cpu memory } user endpoints { host protocol type tls }`

func (s *server) registerService(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "add_service",
		Description: "Add a service to an environment. repository = owner/repo or a full https URL for source builds (mutually exclusive with image, which deploys a prebuilt image). variables set initial env vars; build-time pins like RAILPACK_PYTHON_VERSION belong here so the first build already sees them. cpu and memory (Kubernetes quantities, e.g. '500m'/'512Mi') size the service at creation; pass both or omit both for platform defaults. user is the run-as user id for an image-based service (999 for mysql/postgres/redis, 1000 for node-based images like ghost): it makes stock images whose entrypoint would otherwise chown a data dir as root start cleanly under the hardened default, and the same id owns mounted volumes.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.addService)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "configure_service",
		Description: "Change service settings. Only the fields you provide are applied: start_command, port, replicas, cpu, memory, user. cpu and memory are Kubernetes quantities (e.g. cpu '500m', memory '512Mi'). user sets the run-as user id for image-based services only (rejected for source builds); the same id owns mounted volumes.",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true, DestructiveHint: ptr(false)},
	}, s.configureService)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "set_variables",
		Description: "Set environment variables for a service or shared across an environment (exactly one of service/environment). set entries apply by key; unset removes keys; other keys are preserved. Each entry is exactly one of: a literal value; a ref to a resource variable from list_variables' available list (service scope only, e.g. workspace/proj/env/lucity-app-pg-maindb-app/fqdn-uri wires a database's connection URI into the service); or generate, which stores a random secret you never see (for app-owned secrets like SECRET_KEY_BASE or a JWT secret; a key that already has a value keeps it). Values are never echoed back: the result lists keys and kinds only.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.setVariables)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "list_variables",
		Description: "List a service's environment variables as keys, each a literal or a ref (with its ref id), plus the variables available to reference in that environment (database/kv-store/bucket/shared sources). Values are never returned; confirm one with check_variables. HOST and PORT are injected by the platform for services with a port and are not listed.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.listVariables)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "check_variables",
		Description: "Check whether variables hold the values you expect without revealing them, e.g. that DEBUG is \"true\". Exactly one of service/environment (shared variables). Each check is an exact, case-sensitive comparison returning match, mismatch, missing (key not set), or ref (the key references a resource or shared variable: its ref id is returned and it is not compared, so check a shared variable with environment scope). A mismatch carries detail empty when the stored value is empty, or whitespace when the two differ only in surrounding whitespace. Each key may appear once per call. HOST and PORT are injected by the platform and are not variables.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.checkVariables)
}

type variableEntry struct {
	Key   string `json:"key" jsonschema:"variable name"`
	Value string `json:"value,omitempty" jsonschema:"literal value; mutually exclusive with ref"`
	Ref   string `json:"ref,omitempty" jsonschema:"reference to an availableVariables id from list_variables; mutually exclusive with value"`
}

type addServiceInput struct {
	Environment string          `json:"environment" jsonschema:"environment id (workspace/project/environment)"`
	Name        string          `json:"name,omitempty" jsonschema:"service name (2-16 chars, lowercase alphanumeric and hyphens); derived from the repository if omitted"`
	Repository  string          `json:"repository,omitempty" jsonschema:"owner/repo or full https URL for a source build; mutually exclusive with image"`
	ContextPath string          `json:"context_path,omitempty" jsonschema:"subdirectory within the repository to build from"`
	Image       string          `json:"image,omitempty" jsonschema:"prebuilt image reference (e.g. docker.io/library/nginx:latest); mutually exclusive with repository"`
	Variables   []variableEntry `json:"variables,omitempty" jsonschema:"initial environment variables (literal values only)"`
	CPU         string          `json:"cpu,omitempty" jsonschema:"CPU limit as a Kubernetes quantity (e.g. 500m); provide together with memory, or omit both for platform defaults"`
	Memory      string          `json:"memory,omitempty" jsonschema:"memory limit as a Kubernetes quantity (e.g. 512Mi); provide together with cpu, or omit both for platform defaults"`
	User        *int            `json:"user,omitempty" jsonschema:"run-as user id (0-65535) for image-based services, e.g. 999 for mysql/postgres/redis or 1000 for node-based images; the same id owns mounted volumes; rejected for source builds"`
}

func (s *server) addService(ctx context.Context, _ *mcp.CallToolRequest, input addServiceInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}
	if input.Repository == "" && input.Image == "" {
		return nil, nil, fmt.Errorf("provide either repository (for a source build) or image (for a prebuilt image)")
	}
	if input.Repository != "" && input.Image != "" {
		return nil, nil, fmt.Errorf("repository and image are mutually exclusive")
	}
	if (input.CPU == "") != (input.Memory == "") {
		return nil, nil, fmt.Errorf("cpu and memory must be provided together")
	}

	serviceInput := map[string]any{}
	if input.Name != "" {
		serviceInput["name"] = input.Name
	}
	if input.Repository != "" {
		serviceInput["repository"] = input.Repository
	}
	if input.ContextPath != "" {
		serviceInput["contextPath"] = input.ContextPath
	}
	if input.Image != "" {
		serviceInput["image"] = input.Image
	}
	if len(input.Variables) > 0 {
		vars := make([]map[string]any, 0, len(input.Variables))
		for _, v := range input.Variables {
			vars = append(vars, map[string]any{"key": v.Key, "value": v.Value})
		}
		serviceInput["variables"] = vars
	}
	if input.CPU != "" {
		serviceInput["resources"] = map[string]any{"cpu": input.CPU, "memory": input.Memory}
	}
	if input.User != nil {
		serviceInput["user"] = *input.User
	}

	mutation := `mutation($environment: EnvironmentID!, $input: AddServiceInput!) {
  addService(environment: $environment, input: $input) { ` + serviceSummaryFields + ` }
}`

	var out struct {
		AddService any `json:"addService"`
	}
	if err := s.query(ctx, "add_service", mutation, map[string]any{"environment": environmentID, "input": serviceInput}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"service": out.AddService,
		"note":    "service created but not deployed yet. 'status: FAILED' and a zero port are EXPECTED here (no successful deploy yet, NOT an error). Call deploy once to build and roll it out — the first deploy always builds; after that, config changes (set_variables, configure_service) roll out automatically with no rebuild. Poll get_deploy_status after deploy.",
	})
}

type configureServiceInput struct {
	Service      string  `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	StartCommand string  `json:"start_command,omitempty" jsonschema:"custom start command; overrides the detected default"`
	Port         *int    `json:"port,omitempty" jsonschema:"container port (1-65535)"`
	Replicas     *int    `json:"replicas,omitempty" jsonschema:"desired replica count (1-20)"`
	CPU          string  `json:"cpu,omitempty" jsonschema:"CPU request as a Kubernetes quantity (e.g. 500m)"`
	Memory       string  `json:"memory,omitempty" jsonschema:"memory request as a Kubernetes quantity (e.g. 512Mi)"`
	User         *int    `json:"user,omitempty" jsonschema:"run-as user id (0-65535) for image-based services; the same id owns mounted volumes; -1 clears it back to the image default. Rejected for source builds"`
}

func (s *server) configureService(ctx context.Context, _ *mcp.CallToolRequest, input configureServiceInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	if input.StartCommand != "" {
		const mutation = `mutation($service: ServiceID!, $command: String!) { setCustomStartCommand(service: $service, command: $command) { id } }`
		if err := s.query(ctx, "configure_service (start command)", mutation, map[string]any{"service": serviceID, "command": input.StartCommand}, nil); err != nil {
			return nil, nil, err
		}
	}

	if input.Port != nil {
		const mutation = `mutation($service: ServiceID!, $port: Int) { setServicePort(service: $service, port: $port) { id } }`
		if err := s.query(ctx, "configure_service (port)", mutation, map[string]any{"service": serviceID, "port": *input.Port}, nil); err != nil {
			return nil, nil, err
		}
	}

	if input.Replicas != nil {
		const mutation = `mutation($input: SetServiceScalingInput!) { setServiceScaling(input: $input) { id } }`
		scaling := map[string]any{"service": serviceID, "replicas": *input.Replicas}
		if err := s.query(ctx, "configure_service (scaling)", mutation, map[string]any{"input": scaling}, nil); err != nil {
			return nil, nil, err
		}
	}

	if input.CPU != "" || input.Memory != "" {
		cpu, memory := input.CPU, input.Memory
		if cpu == "" || memory == "" {
			current, err := s.serviceResources(ctx, serviceID)
			if err != nil {
				return nil, nil, err
			}
			if cpu == "" {
				cpu = current.CPU
			}
			if memory == "" {
				memory = current.Memory
			}
		}
		const mutation = `mutation($service: ServiceID!, $resources: ResourcesInput!) { setServiceResources(service: $service, resources: $resources) { id } }`
		resources := map[string]any{"cpu": cpu, "memory": memory}
		if err := s.query(ctx, "configure_service (resources)", mutation, map[string]any{"service": serviceID, "resources": resources}, nil); err != nil {
			return nil, nil, err
		}
	}

	if input.User != nil {
		const mutation = `mutation($service: ServiceID!, $user: Int) { setServiceUser(service: $service, user: $user) { id } }`
		vars := map[string]any{"service": serviceID}
		if *input.User >= 0 {
			vars["user"] = *input.User
		}
		if err := s.query(ctx, "configure_service (user)", mutation, vars, nil); err != nil {
			return nil, nil, err
		}
	}

	summary, err := s.serviceSummary(ctx, serviceID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"service": summary,
		"note":    "applied immediately: the service rolls out with its current image (no rebuild). Poll get_deploy_status to watch the rollout. Do NOT call deploy unless you changed source code — deploy rebuilds from scratch.",
	})
}

type resources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

func (s *server) serviceResources(ctx context.Context, serviceID string) (resources, error) {
	const query = `query($id: ServiceID!) { service(id: $id) { resources { cpu memory } } }`
	var out struct {
		Service struct {
			Resources resources `json:"resources"`
		} `json:"service"`
	}
	if err := s.query(ctx, "configure_service (read resources)", query, map[string]any{"id": serviceID}, &out); err != nil {
		return resources{}, err
	}
	return out.Service.Resources, nil
}

func (s *server) serviceSummary(ctx context.Context, serviceID string) (any, error) {
	query := `query($id: ServiceID!) { service(id: $id) { ` + serviceSummaryFields + ` } }`
	var out struct {
		Service any `json:"service"`
	}
	if err := s.query(ctx, "read service", query, map[string]any{"id": serviceID}, &out); err != nil {
		return nil, err
	}
	return out.Service, nil
}

const (
	minimumGeneratedLength = 16
	maximumGeneratedLength = 256
)

type variable struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
	Ref   *string `json:"ref"`
}

func (v variable) literal() string {
	if v.Value == nil {
		return ""
	}
	return *v.Value
}

func (v variable) view() variableView {
	if v.Ref != nil {
		return variableView{Key: v.Key, Kind: "ref", Ref: *v.Ref}
	}
	return variableView{Key: v.Key, Kind: "literal"}
}

type variableView struct {
	Key  string `json:"key"`
	Kind string `json:"kind"`
	Ref  string `json:"ref,omitempty"`
}

func (s *server) serviceVariables(ctx context.Context, operation, serviceID string) ([]variable, error) {
	const query = `query($service: ServiceID!) { serviceVariables(service: $service) { key value ref } }`
	var out struct {
		ServiceVariables []variable `json:"serviceVariables"`
	}
	if err := s.query(ctx, operation, query, map[string]any{"service": serviceID}, &out); err != nil {
		return nil, err
	}
	return out.ServiceVariables, nil
}

func (s *server) sharedVariables(ctx context.Context, operation, environmentID string) ([]variable, error) {
	const query = `query($environment: EnvironmentID!) { sharedVariables(environment: $environment) { key value } }`
	var out struct {
		SharedVariables []variable `json:"sharedVariables"`
	}
	if err := s.query(ctx, operation, query, map[string]any{"environment": environmentID}, &out); err != nil {
		return nil, err
	}
	return out.SharedVariables, nil
}

func randomSecret(length int) string {
	buffer := make([]byte, (length+1)/2)
	rand.Read(buffer)
	return hex.EncodeToString(buffer)[:length]
}

type setVariableEntry struct {
	Key      string  `json:"key" jsonschema:"variable name"`
	Value    *string `json:"value,omitempty" jsonschema:"literal value, may be empty; exactly one of value, ref, generate"`
	Ref      string  `json:"ref,omitempty" jsonschema:"reference id from list_variables' available list; service scope only; exactly one of value, ref, generate"`
	Generate int     `json:"generate,omitempty" jsonschema:"length (16-256) of a random hex secret to store instead of a value; it is never returned, and a key that already has a value keeps it (unset it in a separate call first to rotate). 64 suits SECRET_KEY_BASE, Django SECRET_KEY, AUTH_SECRET and JWT secrets; a Laravel APP_KEY needs exactly 32. Only for secrets the app owns, never for values that must match something elsewhere such as RAILS_MASTER_KEY or third-party API keys; exactly one of value, ref, generate"`
}

func validateVariableEntry(entry setVariableEntry, unset map[string]bool) error {
	if unset[entry.Key] {
		return fmt.Errorf("variable %q is in both set and unset", entry.Key)
	}
	given := 0
	for _, present := range []bool{entry.Value != nil, entry.Ref != "", entry.Generate != 0} {
		if present {
			given++
		}
	}
	if given != 1 {
		return fmt.Errorf("variable %q: set exactly one of value, ref, or generate", entry.Key)
	}
	if entry.Generate != 0 && (entry.Generate < minimumGeneratedLength || entry.Generate > maximumGeneratedLength) {
		return fmt.Errorf("variable %q: generate must be between %d and %d characters", entry.Key, minimumGeneratedLength, maximumGeneratedLength)
	}
	return nil
}

type setVariablesInput struct {
	Service     string             `json:"service,omitempty" jsonschema:"service id to set service-scoped variables; exactly one of service/environment"`
	Environment string             `json:"environment,omitempty" jsonschema:"environment id to set shared variables; exactly one of service/environment"`
	Set         []setVariableEntry `json:"set,omitempty" jsonschema:"variables to set or update by key"`
	Unset       []string           `json:"unset,omitempty" jsonschema:"variable keys to remove"`
}

func (s *server) setVariables(ctx context.Context, _ *mcp.CallToolRequest, input setVariablesInput) (*mcp.CallToolResult, any, error) {
	if (input.Service == "") == (input.Environment == "") {
		return nil, nil, fmt.Errorf("provide exactly one of service or environment")
	}
	unset := make(map[string]bool, len(input.Unset))
	for _, key := range input.Unset {
		unset[key] = true
	}
	for _, entry := range input.Set {
		if err := validateVariableEntry(entry, unset); err != nil {
			return nil, nil, err
		}
	}
	if input.Environment != "" {
		return s.setSharedVariables(ctx, input)
	}
	return s.setServiceVariables(ctx, input)
}

func (s *server) setSharedVariables(ctx context.Context, input setVariablesInput) (*mcp.CallToolResult, any, error) {
	environmentID, err := s.environmentID(input.Environment)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range input.Set {
		if entry.Ref != "" {
			return nil, nil, fmt.Errorf("shared (environment) variables are literal values only; refs are not allowed — set the ref on a service with service scope instead")
		}
	}

	current, err := s.sharedVariables(ctx, "set_variables (read shared)", environmentID)
	if err != nil {
		return nil, nil, err
	}

	merged := map[string]string{}
	for _, v := range current {
		merged[v.Key] = v.literal()
	}
	var generated, kept []string
	for _, entry := range input.Set {
		switch {
		case entry.Generate != 0 && merged[entry.Key] != "":
			kept = append(kept, entry.Key)
		case entry.Generate != 0:
			merged[entry.Key] = randomSecret(entry.Generate)
			generated = append(generated, entry.Key)
		default:
			merged[entry.Key] = *entry.Value
		}
	}
	for _, key := range input.Unset {
		delete(merged, key)
	}

	keys := slices.Sorted(maps.Keys(merged))
	variables := make([]map[string]any, 0, len(keys))
	views := make([]variableView, 0, len(keys))
	for _, key := range keys {
		variables = append(variables, map[string]any{"key": key, "value": merged[key]})
		views = append(views, variableView{Key: key, Kind: "literal"})
	}

	const mutation = `mutation($environment: EnvironmentID!, $variables: [VariableInput!]!) { setSharedVariables(environment: $environment, variables: $variables) }`
	if err := s.query(ctx, "set_variables (shared)", mutation, map[string]any{"environment": environmentID, "variables": variables}, nil); err != nil {
		return nil, nil, err
	}

	return variablesResult("environment", views, generated, kept, "applied immediately: services in this environment roll out with their current images (no rebuild). Do NOT call deploy unless you changed source code.")
}

func variablesResult(scope string, views []variableView, generated, kept []string, note string) (*mcp.CallToolResult, any, error) {
	result := map[string]any{
		"scope":     scope,
		"variables": views,
		"note":      note,
	}
	if len(generated) > 0 {
		result["generated"] = generated
	}
	if len(kept) > 0 {
		result["kept"] = kept
		result["note"] = note + " Keys under kept already had a value, so generate left them unchanged; to rotate one, unset it in one call and generate it in the next."
	}
	return jsonResult(result)
}

func (s *server) setServiceVariables(ctx context.Context, input setVariablesInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	current, err := s.serviceVariables(ctx, "set_variables (read service)", serviceID)
	if err != nil {
		return nil, nil, err
	}

	merged := map[string]variable{}
	order := []string{}
	for _, v := range current {
		merged[v.Key] = v
		order = append(order, v.Key)
	}
	usedRef := false
	var generated, kept []string
	for _, entry := range input.Set {
		existing, exists := merged[entry.Key]
		switch {
		case entry.Generate != 0 && exists && (existing.Ref != nil || existing.literal() != ""):
			kept = append(kept, entry.Key)
			continue
		case entry.Generate != 0:
			merged[entry.Key] = variable{Key: entry.Key, Value: ptr(randomSecret(entry.Generate))}
			generated = append(generated, entry.Key)
		case entry.Ref != "":
			merged[entry.Key] = variable{Key: entry.Key, Ref: ptr(entry.Ref)}
			usedRef = true
		default:
			merged[entry.Key] = variable{Key: entry.Key, Value: entry.Value}
		}
		if !exists {
			order = append(order, entry.Key)
		}
	}
	for _, key := range input.Unset {
		delete(merged, key)
	}

	variables := make([]map[string]any, 0, len(merged))
	views := make([]variableView, 0, len(merged))
	for _, key := range order {
		v, ok := merged[key]
		if !ok {
			continue
		}
		item := map[string]any{"key": key}
		if v.Ref != nil {
			item["ref"] = *v.Ref
		} else {
			item["value"] = v.literal()
		}
		variables = append(variables, item)
		views = append(views, v.view())
	}

	const mutation = `mutation($service: ServiceID!, $variables: [ServiceVariableInput!]!) { setServiceVariables(service: $service, variables: $variables) }`
	if err := s.query(ctx, "set_variables (service)", mutation, map[string]any{"service": serviceID, "variables": variables}, nil); err != nil {
		if usedRef {
			return nil, nil, fmt.Errorf("%w — a ref must be an id from list_variables' available list for this environment; call list_variables to see valid refs", err)
		}
		return nil, nil, err
	}

	return variablesResult("service", views, generated, kept, "applied immediately: the service rolls out with its current image (no rebuild). Poll get_deploy_status to watch the rollout. Do NOT call deploy unless you changed source code — deploy rebuilds from scratch.")
}

type listVariablesInput struct {
	Service string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
}

func (s *server) listVariables(ctx context.Context, _ *mcp.CallToolRequest, input listVariablesInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}
	environmentID := environmentOfService(serviceID)

	const query = `query($service: ServiceID!, $environment: EnvironmentID!) {
  serviceVariables(service: $service) { key ref }
  availableVariables(environment: $environment) {
    id key
    source {
      __typename
      ... on DatabaseSource { name }
      ... on KeyValueStoreSource { name }
      ... on BucketSource { name }
      ... on SharedSource { name }
    }
  }
}`

	var out struct {
		ServiceVariables   []variable `json:"serviceVariables"`
		AvailableVariables any        `json:"availableVariables"`
	}
	if err := s.query(ctx, "list_variables", query, map[string]any{"service": serviceID, "environment": environmentID}, &out); err != nil {
		return nil, nil, err
	}

	views := make([]variableView, 0, len(out.ServiceVariables))
	for _, v := range out.ServiceVariables {
		views = append(views, v.view())
	}
	return jsonResult(map[string]any{
		"variables": views,
		"available": out.AvailableVariables,
		"note":      "values are never returned: confirm one with check_variables, and wire an available entry into the service with set_variables using its id as a ref",
	})
}

type variableCheck struct {
	Key      string `json:"key" jsonschema:"variable name"`
	Expected string `json:"expected" jsonschema:"the exact value you expect, compared case-sensitively; may be empty"`
}

type checkVariablesInput struct {
	Service     string          `json:"service,omitempty" jsonschema:"service id to check service-scoped variables; exactly one of service/environment"`
	Environment string          `json:"environment,omitempty" jsonschema:"environment id to check shared variables; exactly one of service/environment"`
	Checks      []variableCheck `json:"checks" jsonschema:"variables to compare; each key at most once per call"`
}

type variableCheckResult struct {
	Key    string `json:"key"`
	Result string `json:"result"`
	Ref    string `json:"ref,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func (s *server) checkVariables(ctx context.Context, _ *mcp.CallToolRequest, input checkVariablesInput) (*mcp.CallToolResult, any, error) {
	if (input.Service == "") == (input.Environment == "") {
		return nil, nil, fmt.Errorf("provide exactly one of service or environment")
	}
	if len(input.Checks) == 0 {
		return nil, nil, fmt.Errorf("provide at least one check")
	}
	seen := make(map[string]bool, len(input.Checks))
	for _, check := range input.Checks {
		if seen[check.Key] {
			return nil, nil, fmt.Errorf("variable %q appears more than once; check each key at most once per call", check.Key)
		}
		seen[check.Key] = true
	}

	var variables []variable
	if input.Environment != "" {
		environmentID, err := s.environmentID(input.Environment)
		if err != nil {
			return nil, nil, err
		}
		if variables, err = s.sharedVariables(ctx, "check_variables", environmentID); err != nil {
			return nil, nil, err
		}
	} else {
		serviceID, err := s.serviceID(input.Service)
		if err != nil {
			return nil, nil, err
		}
		if variables, err = s.serviceVariables(ctx, "check_variables", serviceID); err != nil {
			return nil, nil, err
		}
	}

	byKey := make(map[string]variable, len(variables))
	for _, v := range variables {
		byKey[v.Key] = v
	}
	results := make([]variableCheckResult, 0, len(input.Checks))
	for _, check := range input.Checks {
		results = append(results, compareVariable(byKey[check.Key], check))
	}
	return jsonResult(map[string]any{"results": results})
}

func compareVariable(current variable, check variableCheck) variableCheckResult {
	result := variableCheckResult{Key: check.Key}
	switch {
	case current.Key == "":
		result.Result = "missing"
	case current.Ref != nil:
		result.Result = "ref"
		result.Ref = *current.Ref
	case current.literal() == check.Expected:
		result.Result = "match"
	default:
		result.Result = "mismatch"
		if current.literal() == "" {
			result.Detail = "empty"
		} else if strings.TrimSpace(current.literal()) == strings.TrimSpace(check.Expected) {
			result.Detail = "whitespace"
		}
	}
	return result
}
