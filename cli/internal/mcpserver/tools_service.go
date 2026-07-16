package mcpserver

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serviceSummaryFields = `id name status replicas { desired ready } port command resources { cpu memory } endpoints { host protocol type tls }`

func (s *server) registerService(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "add_service",
		Description: "Add a service to an environment. repository = owner/repo or a full https URL for source builds (mutually exclusive with image, which deploys a prebuilt image). variables set initial env vars; build-time pins like RAILPACK_PYTHON_VERSION belong here so the first build already sees them. cpu and memory (Kubernetes quantities, e.g. '500m'/'512Mi') size the service at creation; pass both or omit both for platform defaults.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.addService)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "configure_service",
		Description: "Change service settings. Only the fields you provide are applied: start_command, port, replicas, cpu, memory. cpu and memory are Kubernetes quantities (e.g. cpu '500m', memory '512Mi').",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true, DestructiveHint: ptr(false)},
	}, s.configureService)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "set_variables",
		Description: "Set environment variables for a service or shared across an environment (exactly one of service/environment). set entries apply by key; unset removes keys; other keys are preserved. A service variable is either a literal value or a ref to a resource variable (from list_variables' available list, e.g. workspace/proj/env/maindb-app/DATABASE_URL wires a database's DATABASE_URL into the service). Shared (environment) variables are literal values only.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.setVariables)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "list_variables",
		Description: "List a service's environment variables (literal values or refs) plus the variables available to reference in that environment (database/kv-store/bucket/shared sources).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.listVariables)
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
	Service      string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	StartCommand string `json:"start_command,omitempty" jsonschema:"custom start command; overrides the detected default"`
	Port         *int   `json:"port,omitempty" jsonschema:"container port (1-65535)"`
	Replicas     *int   `json:"replicas,omitempty" jsonschema:"desired replica count (1-20)"`
	CPU          string `json:"cpu,omitempty" jsonschema:"CPU request as a Kubernetes quantity (e.g. 500m)"`
	Memory       string `json:"memory,omitempty" jsonschema:"memory request as a Kubernetes quantity (e.g. 512Mi)"`
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

type setVariableEntry struct {
	Key   string `json:"key" jsonschema:"variable name"`
	Value string `json:"value,omitempty" jsonschema:"literal value; mutually exclusive with ref"`
	Ref   string `json:"ref,omitempty" jsonschema:"reference id from list_variables' available list; service scope only; mutually exclusive with value"`
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

	const readQuery = `query($environment: EnvironmentID!) { sharedVariables(environment: $environment) { key value } }`
	var current struct {
		SharedVariables []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"sharedVariables"`
	}
	if err := s.query(ctx, "set_variables (read shared)", readQuery, map[string]any{"environment": environmentID}, &current); err != nil {
		return nil, nil, err
	}

	merged := map[string]string{}
	for _, v := range current.SharedVariables {
		merged[v.Key] = v.Value
	}
	for _, entry := range input.Set {
		merged[entry.Key] = entry.Value
	}
	for _, key := range input.Unset {
		delete(merged, key)
	}

	variables := make([]map[string]any, 0, len(merged))
	for key, value := range merged {
		variables = append(variables, map[string]any{"key": key, "value": value})
	}

	const mutation = `mutation($environment: EnvironmentID!, $variables: [VariableInput!]!) { setSharedVariables(environment: $environment, variables: $variables) }`
	if err := s.query(ctx, "set_variables (shared)", mutation, map[string]any{"environment": environmentID, "variables": variables}, nil); err != nil {
		return nil, nil, err
	}

	var result struct {
		SharedVariables any `json:"sharedVariables"`
	}
	if err := s.query(ctx, "set_variables (read shared)", readQuery, map[string]any{"environment": environmentID}, &result); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"scope":     "environment",
		"variables": result.SharedVariables,
		"note":      "applied immediately: services in this environment roll out with their current images (no rebuild). Do NOT call deploy unless you changed source code.",
	})
}

func (s *server) setServiceVariables(ctx context.Context, input setVariablesInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range input.Set {
		if entry.Value != "" && entry.Ref != "" {
			return nil, nil, fmt.Errorf("variable %q: set either value or ref, not both", entry.Key)
		}
	}

	const readQuery = `query($service: ServiceID!) { serviceVariables(service: $service) { key value ref } }`
	var current struct {
		ServiceVariables []struct {
			Key   string  `json:"key"`
			Value *string `json:"value"`
			Ref   *string `json:"ref"`
		} `json:"serviceVariables"`
	}
	if err := s.query(ctx, "set_variables (read service)", readQuery, map[string]any{"service": serviceID}, &current); err != nil {
		return nil, nil, err
	}

	type entry struct {
		value *string
		ref   *string
	}
	merged := map[string]entry{}
	order := []string{}
	seen := map[string]bool{}
	appendKey := func(key string) {
		if !seen[key] {
			seen[key] = true
			order = append(order, key)
		}
	}
	for _, v := range current.ServiceVariables {
		merged[v.Key] = entry{value: v.Value, ref: v.Ref}
		appendKey(v.Key)
	}
	usedRef := false
	for _, e := range input.Set {
		if e.Ref != "" {
			merged[e.Key] = entry{ref: ptr(e.Ref)}
			usedRef = true
		} else {
			merged[e.Key] = entry{value: ptr(e.Value)}
		}
		appendKey(e.Key)
	}
	for _, key := range input.Unset {
		delete(merged, key)
	}

	variables := make([]map[string]any, 0, len(merged))
	for _, key := range order {
		e, ok := merged[key]
		if !ok {
			continue
		}
		item := map[string]any{"key": key}
		if e.ref != nil {
			item["ref"] = *e.ref
		} else if e.value != nil {
			item["value"] = *e.value
		} else {
			item["value"] = ""
		}
		variables = append(variables, item)
	}

	const mutation = `mutation($service: ServiceID!, $variables: [ServiceVariableInput!]!) { setServiceVariables(service: $service, variables: $variables) }`
	if err := s.query(ctx, "set_variables (service)", mutation, map[string]any{"service": serviceID, "variables": variables}, nil); err != nil {
		if usedRef {
			return nil, nil, fmt.Errorf("%w — a ref must be an id from list_variables' available list for this environment; call list_variables to see valid refs", err)
		}
		return nil, nil, err
	}

	var result struct {
		ServiceVariables any `json:"serviceVariables"`
	}
	if err := s.query(ctx, "set_variables (read service)", readQuery, map[string]any{"service": serviceID}, &result); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"scope":     "service",
		"variables": result.ServiceVariables,
		"note":      "applied immediately: the service rolls out with its current image (no rebuild). Poll get_deploy_status to watch the rollout. Do NOT call deploy unless you changed source code — deploy rebuilds from scratch.",
	})
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
  serviceVariables(service: $service) { key value ref }
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
		ServiceVariables   any `json:"serviceVariables"`
		AvailableVariables any `json:"availableVariables"`
	}
	if err := s.query(ctx, "list_variables", query, map[string]any{"service": serviceID, "environment": environmentID}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"variables": out.ServiceVariables,
		"available": out.AvailableVariables,
		"note":      "wire an available entry into the service with set_variables using its id as a ref",
	})
}
