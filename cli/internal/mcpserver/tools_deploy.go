package mcpserver

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *server) registerDeploy(m *mcp.Server) {
	mcp.AddTool(m, &mcp.Tool{
		Name:        "deploy",
		Description: "Build and roll out a service, creating a new release. git_ref optionally pins the branch, tag, or commit to build. Deploys are asynchronous: poll get_deploy_status.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.deploy)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_deploy_status",
		Description: "Get the status of a service and one of its releases (build, deploy, security scan, and rollout phases) plus its endpoints. Without release_id the newest release is used. Read get_logs for a failed phase.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getDeployStatus)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_logs",
		Description: "Fetch logs for a build, deploy, security scan, or a running service. kind=runtime streams live service logs and never completes, so it is sampled for a few seconds. Use the build/deploy/scan id from get_deploy_status, or the service id for runtime.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getLogs)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "rollback",
		Description: "Roll a service back to a previous deployment. Pass the deployment id of a previous ACTIVE or SUPERSEDED deployment (from get_deploy_status).",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.rollback)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "add_domain",
		Description: "Add a domain to a service. Without hostname a platform subdomain is generated. With hostname a custom domain is added; the response includes the DNS records to create and the TLS provisioning status.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.addDomain)
}

type deployInput struct {
	Service string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	GitRef  string `json:"git_ref,omitempty" jsonschema:"branch, tag, or commit to build; defaults to the service's default branch"`
}

func (s *server) deploy(ctx context.Context, _ *mcp.CallToolRequest, input deployInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($service: ServiceID!, $gitRef: String) {
  deploy(service: $service, gitRef: $gitRef) { id status build { id status } createdAt }
}`
	variables := map[string]any{"service": serviceID}
	if input.GitRef != "" {
		variables["gitRef"] = input.GitRef
	}

	var out struct {
		Deploy struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			CreatedAt string `json:"createdAt"`
			Build     *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"build"`
		} `json:"deploy"`
	}
	if err := s.query(ctx, "deploy", mutation, variables, &out); err != nil {
		return nil, nil, err
	}

	result := map[string]any{
		"releaseId": out.Deploy.ID,
		"status":    out.Deploy.Status,
		"createdAt": out.Deploy.CreatedAt,
		"note":      fmt.Sprintf("poll get_deploy_status with service=%q release_id=%q", serviceID, out.Deploy.ID),
	}
	if out.Deploy.Build != nil {
		result["buildId"] = out.Deploy.Build.ID
		result["buildStatus"] = out.Deploy.Build.Status
	}
	return jsonResult(result)
}

type getDeployStatusInput struct {
	Service   string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	ReleaseID string `json:"release_id,omitempty" jsonschema:"release id to inspect; defaults to the newest release"`
}

func (s *server) getDeployStatus(ctx context.Context, _ *mcp.CallToolRequest, input getDeployStatusInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	const query = `query($id: ServiceID!) {
  service(id: $id) {
    id status
    endpoints { host protocol type tls }
    releases {
      id status createdAt
      trigger { kind }
      source { ref commit { sha message } }
      build { id status startedAt finishedAt }
      deploy { id status }
      scan { id status findingsCount }
      deployment { id status replicas { desired ready } rollout { status reason message restarts } }
    }
  }
}`

	var out struct {
		Service struct {
			ID        string           `json:"id"`
			Status    string           `json:"status"`
			Endpoints []any            `json:"endpoints"`
			Releases  []map[string]any `json:"releases"`
		} `json:"service"`
	}
	if err := s.query(ctx, "get_deploy_status", query, map[string]any{"id": serviceID}, &out); err != nil {
		return nil, nil, err
	}

	release := pickRelease(out.Service.Releases, input.ReleaseID)
	if release == nil {
		return nil, nil, fmt.Errorf("no release found for service %q%s — call deploy first", serviceID, releaseIDSuffix(input.ReleaseID))
	}

	result := map[string]any{
		"serviceStatus": out.Service.Status,
		"endpoints":     out.Service.Endpoints,
		"release":       release,
	}
	if hints := actionHints(release); len(hints) > 0 {
		result["actions"] = hints
	}
	return jsonResult(result)
}

func releaseIDSuffix(releaseID string) string {
	if releaseID == "" {
		return ""
	}
	return fmt.Sprintf(" with id %q", releaseID)
}

func pickRelease(releases []map[string]any, releaseID string) map[string]any {
	if len(releases) == 0 {
		return nil
	}
	if releaseID != "" {
		for _, r := range releases {
			if id, _ := r["id"].(string); id == releaseID {
				return r
			}
		}
		return nil
	}
	sorted := make([]map[string]any, len(releases))
	copy(sorted, releases)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, _ := sorted[i]["createdAt"].(string)
		b, _ := sorted[j]["createdAt"].(string)
		return a > b
	})
	return sorted[0]
}

func actionHints(release map[string]any) []string {
	var hints []string

	if build, ok := release["build"].(map[string]any); ok {
		if status, _ := build["status"].(string); status == "FAILED" {
			if id, _ := build["id"].(string); id != "" {
				hints = append(hints, fmt.Sprintf("build FAILED — call get_logs with kind=\"build\" id=%q", id))
			}
		}
	}

	if deploy, ok := release["deploy"].(map[string]any); ok {
		if status, _ := deploy["status"].(string); status == "FAILED" {
			if id, _ := deploy["id"].(string); id != "" {
				hints = append(hints, fmt.Sprintf("deploy FAILED — call get_logs with kind=\"deploy\" id=%q", id))
			}
		}
	}

	if deployment, ok := release["deployment"].(map[string]any); ok {
		serviceID := deploymentServiceID(deployment)
		if rollout, ok := deployment["rollout"].(map[string]any); ok {
			status, _ := rollout["status"].(string)
			reason, _ := rollout["reason"].(string)
			if status == "DEGRADED" || status == "FAILED" {
				hints = append(hints, rolloutHint(reason, serviceID)...)
			}
		}
	}

	return hints
}

func deploymentServiceID(deployment map[string]any) string {
	id, _ := deployment["id"].(string)
	parts := strings.Split(id, "/")
	if len(parts) < 5 {
		return ""
	}
	return strings.Join(parts[:4], "/")
}

func rolloutHint(reason, serviceID string) []string {
	switch reason {
	case "OOM_KILLED":
		return []string{"rollout DEGRADED: the container was killed for exceeding its memory limit — raise memory via configure_service, which rolls the service out automatically with its current image (no rebuild); then poll get_deploy_status"}
	case "CRASH_LOOP":
		hint := "rollout DEGRADED: the container keeps crashing on startup"
		if serviceID != "" {
			hint += fmt.Sprintf(" — call get_logs with kind=\"runtime\" id=%q to see why", serviceID)
		}
		return []string{hint}
	case "IMAGE_PULL_FAILED":
		return []string{"rollout DEGRADED: the image could not be pulled — verify the image reference (add_service image) or that the build succeeded"}
	case "CONFIG_ERROR":
		return []string{"rollout DEGRADED: invalid configuration (e.g. a bad env var or reference) — review set_variables and configure_service"}
	case "UNSCHEDULABLE":
		return []string{"rollout DEGRADED: no node can schedule the pod — reduce requested resources via configure_service or the environment is at capacity"}
	case "QUOTA_EXCEEDED":
		return []string{"rollout DEGRADED: the environment resource quota is exhausted — raise the environment quota in the dashboard or reduce service resources via configure_service"}
	case "DEADLINE_EXCEEDED":
		return []string{"rollout DEGRADED: the rollout timed out becoming ready — check get_logs (kind=runtime) and readiness of the service"}
	case "NOT_READY":
		return []string{"rollout not ready yet — poll get_deploy_status again shortly"}
	default:
		return []string{"rollout DEGRADED/FAILED — inspect get_logs (kind=runtime) for the service"}
	}
}

type rollbackInput struct {
	Deployment string `json:"deployment" jsonschema:"deployment id (workspace/project/environment/service/hash) of a previous ACTIVE or SUPERSEDED deployment"`
}

func (s *server) rollback(ctx context.Context, _ *mcp.CallToolRequest, input rollbackInput) (*mcp.CallToolResult, any, error) {
	deploymentID, err := s.scopedID(input.Deployment, 5)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($deployment: DeploymentID!) { rollback(deployment: $deployment) }`
	var out struct {
		Rollback bool `json:"rollback"`
	}
	if err := s.query(ctx, "rollback", mutation, map[string]any{"deployment": deploymentID}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"rolledBack": out.Rollback,
		"note":       "poll get_deploy_status to watch the rolled-back deployment become active",
	})
}

type addDomainInput struct {
	Service  string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	Hostname string `json:"hostname,omitempty" jsonschema:"custom hostname (4-253 chars); omit to generate a platform subdomain"`
}

func (s *server) addDomain(ctx context.Context, _ *mcp.CallToolRequest, input addDomainInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	const endpointFields = `endpoints { host protocol type tls dns { status requiredRecords { type host value } } }`

	if input.Hostname == "" {
		mutation := `mutation($service: ServiceID!) { generateDomain(service: $service) { id ` + endpointFields + ` } }`
		var out struct {
			GenerateDomain any `json:"generateDomain"`
		}
		if err := s.query(ctx, "add_domain", mutation, map[string]any{"service": serviceID}, &out); err != nil {
			return nil, nil, err
		}
		return jsonResult(map[string]any{"service": out.GenerateDomain})
	}

	mutation := `mutation($service: ServiceID!, $hostname: String!) { addCustomDomain(service: $service, hostname: $hostname) { id ` + endpointFields + ` } }`
	var out struct {
		AddCustomDomain any `json:"addCustomDomain"`
	}
	if err := s.query(ctx, "add_domain", mutation, map[string]any{"service": serviceID, "hostname": input.Hostname}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"service": out.AddCustomDomain,
		"note":    "create the DNS records shown under dns.requiredRecords, then poll get_deploy_status for TLS to become ACTIVE",
	})
}
