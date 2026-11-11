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
		Description: "Build and roll out a repository service, creating a new release. add_service already starts the first build, so call this for later source or RAILPACK_* changes; an image service has nothing to build. git_ref optionally pins the branch, tag, or commit to build. Deploys are asynchronous: poll get_deploy_status.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.deploy)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_deploy_status",
		Description: "Get the status of a service and one of its releases (build, deploy, security scan, and rollout phases) plus its endpoints, and whether a push to its branch deploys on its own (autoDeploy; branch null means the repository's default branch). When the scan found committed credentials, secretScan lists where. Without release_id the newest release is used, which can be newer than the one serving traffic: liveRelease then names the one that serves. Pass wait to block until the release settles instead of polling. Read get_logs for a failed phase.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getDeployStatus)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "get_logs",
		Description: "Fetch logs for a build, deploy, security scan, or a service at runtime. kind=runtime reads the service's logs of the past hour (or since), including instances that crashed or were replaced, and can narrow them by search text or instance; credentials, and variable values that look secret (by key name, a random-looking value, or a URL with a password), are replaced in runtime logs with [redacted:<source>]. Pass id (the build, deploy or scan id from get_deploy_status) for those kinds, and service for runtime. Example: {\"kind\": \"runtime\", \"service\": \"acme/shop/development/web\", \"since\": \"15m\", \"search\": \"error\"}.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true, DestructiveHint: ptr(false)},
	}, s.getLogs)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "cancel_build",
		Description: "Cancel a queued or running build, e.g. one started for a bad commit. Its release ends CANCELLED and never rolls out; the service keeps serving its current release. Pass the build id from get_deploy_status or deploy.",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.cancelBuild)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "rollback",
		Description: "Roll a service back to a previous deployment. Pass the deployment id of a previous ACTIVE or SUPERSEDED deployment (from get_deploy_status).",
		Annotations: &mcp.ToolAnnotations{DestructiveHint: ptr(false)},
	}, s.rollback)

	mcp.AddTool(m, &mcp.Tool{
		Name:        "add_domain",
		Description: "Add a domain to a service. Without hostname a platform subdomain is generated. With hostname a custom domain is added; the response includes the DNS records to create and the TLS provisioning status. An apex domain (example.com) either redirects to a subdomain via redirect_to (usually www.example.com, which must be added first) or is served directly, which needs an ALIAS record and therefore a DNS provider with ALIAS, ANAME or CNAME flattening support.",
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
	Wait      int    `json:"wait,omitempty" jsonschema:"optional: seconds (1-300) to wait until the release is live, failed, cancelled or superseded, or its rollout reports a failure, instead of polling; 120 suits most builds"`
}

func (s *server) getDeployStatus(ctx context.Context, _ *mcp.CallToolRequest, input getDeployStatusInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}
	if err := validateWait(input.Wait); err != nil {
		return nil, nil, err
	}

	result, settled, err := pollUntil(ctx, input.Wait, func() (map[string]any, bool, error) {
		return s.deployStatus(ctx, serviceID, input.ReleaseID)
	})
	if err != nil {
		return nil, nil, err
	}
	release, _ := result["release"].(map[string]any)
	if note := waitNote(input.Wait, settled, release["status"]); note != "" {
		result["note"] = note
	}
	return jsonResult(result)
}

func releaseSettled(release map[string]any) bool {
	switch status, _ := release["status"].(string); status {
	case "FAILED", "CANCELLED", "SUPERSEDED":
		return true
	case "LIVE":
		rollout := releaseRollout(release)
		return rollout == nil || rollout["status"] != "PROGRESSING"
	}
	rollout := releaseRollout(release)
	return rollout != nil && (rollout["status"] == "DEGRADED" || rollout["status"] == "FAILED")
}

func releaseRollout(release map[string]any) map[string]any {
	deployment, _ := release["deployment"].(map[string]any)
	rollout, _ := deployment["rollout"].(map[string]any)
	return rollout
}

func (s *server) deployStatus(ctx context.Context, serviceID, releaseID string) (map[string]any, bool, error) {
	const query = `query($id: ServiceID!) {
  service(id: $id) {
    id status port branch autoDeploy
    endpoints { host protocol type tls redirectTo }
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
			ID         string           `json:"id"`
			Status     string           `json:"status"`
			Port       int              `json:"port"`
			Branch     *string          `json:"branch"`
			AutoDeploy bool             `json:"autoDeploy"`
			Endpoints  []any            `json:"endpoints"`
			Releases   []map[string]any `json:"releases"`
		} `json:"service"`
	}
	if err := s.query(ctx, "get_deploy_status", query, map[string]any{"id": serviceID}, &out); err != nil {
		return nil, false, err
	}

	release := pickRelease(out.Service.Releases, releaseID)
	if release == nil {
		return nil, false, fmt.Errorf("no release found for service %q%s — call deploy first", serviceID, releaseIDSuffix(releaseID))
	}

	result := map[string]any{
		"serviceStatus": out.Service.Status,
		"autoDeploy":    out.Service.AutoDeploy,
		"branch":        out.Service.Branch,
		"endpoints":     out.Service.Endpoints,
		"release":       release,
	}
	hints := actionHints(release, out.Service.Port)
	if scan, _ := release["scan"].(map[string]any); scan["status"] == "FINDINGS" {
		report, err := s.secretScanReport(ctx, serviceID)
		if err != nil {
			return nil, false, err
		}
		if report != nil {
			result["secretScan"] = report
			hints = append(hints, "the secret scan found credentials committed to the repository, listed under secretScan.findings. Verified ones are live: rotate them, then remove them from the code")
		}
	}
	if live := liveRelease(out.Service.Releases); live != nil && live["id"] != release["id"] {
		result["liveRelease"] = map[string]any{"id": live["id"], "status": live["status"], "source": live["source"]}
		if status, _ := release["status"].(string); status == "FAILED" || status == "CANCELLED" {
			hints = append(hints, fmt.Sprintf("this release did not go live; the service still serves release %v, so serviceStatus describes that one", live["id"]))
		}
	}
	if len(hints) > 0 {
		result["actions"] = hints
	}
	return result, releaseSettled(release), nil
}

func (s *server) secretScanReport(ctx context.Context, serviceID string) (any, error) {
	const query = `query($id: ServiceID!) {
  service(id: $id) {
    secretScanReport { commit scannedAt findings { rule file line commit verified url } }
  }
}`
	var out struct {
		Service struct {
			SecretScanReport any `json:"secretScanReport"`
		} `json:"service"`
	}
	if err := s.query(ctx, "get_deploy_status (secret scan)", query, map[string]any{"id": serviceID}, &out); err != nil {
		return nil, err
	}
	return out.Service.SecretScanReport, nil
}

func liveRelease(releases []map[string]any) map[string]any {
	for _, r := range releases {
		if status, _ := r["status"].(string); status == "LIVE" {
			return r
		}
	}
	return nil
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

func actionHints(release map[string]any, port int) []string {
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
				hints = append(hints, rolloutHint(reason, serviceID, port)...)
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

func rolloutHint(reason, serviceID string, port int) []string {
	switch reason {
	case "OOM_KILLED":
		return []string{"rollout DEGRADED: the container was killed for exceeding its memory limit — raise memory via configure_service, which rolls the service out automatically with its current image (no rebuild); then poll get_deploy_status"}
	case "CRASH_LOOP":
		hint := "rollout DEGRADED: the container keeps crashing on startup. rollout.message has the exit code and, when the container could not even start (a missing shell or binary, a permission error), the runtime's reason"
		if serviceID != "" {
			hint += fmt.Sprintf("; get_logs with kind=\"runtime\" service=%q shows what the app printed before it exited", serviceID)
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
		hint := fmt.Sprintf("rollout FAILED: the container runs but never became ready. The platform checks port %d and passes it to the app as PORT; an app listening on another port never gets ready, so make it listen on $PORT or set its port with configure_service", port)
		if serviceID != "" {
			hint += fmt.Sprintf(". get_logs with kind=\"runtime\" service=%q shows what it listens on", serviceID)
		}
		return []string{hint}
	default:
		return []string{"rollout DEGRADED/FAILED — inspect get_logs (kind=runtime) for the service"}
	}
}

type cancelBuildInput struct {
	Build string `json:"build" jsonschema:"build id (workspace/build-...), from get_deploy_status or deploy"`
}

func (s *server) cancelBuild(ctx context.Context, _ *mcp.CallToolRequest, input cancelBuildInput) (*mcp.CallToolResult, any, error) {
	buildID, err := s.scopedID(input.Build, 2)
	if err != nil {
		return nil, nil, err
	}

	const mutation = `mutation($build: BuildID!) { cancelBuild(build: $build) { id status } }`
	var out struct {
		CancelBuild any `json:"cancelBuild"`
	}
	if err := s.query(ctx, "cancel_build", mutation, map[string]any{"build": buildID}, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"build": out.CancelBuild,
		"note":  "a running build stops within seconds; poll get_deploy_status to see its release end CANCELLED",
	})
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
	Service    string `json:"service" jsonschema:"service id (workspace/project/environment/service)"`
	Hostname   string `json:"hostname,omitempty" jsonschema:"custom hostname (4-253 chars); omit to generate a platform subdomain"`
	RedirectTo string `json:"redirect_to,omitempty" jsonschema:"another domain already added to the same service that hostname should permanently redirect to, e.g. www.example.com for the apex example.com (add the target first). Re-adding an existing hostname without it clears the redirect"`
}

func (s *server) addDomain(ctx context.Context, _ *mcp.CallToolRequest, input addDomainInput) (*mcp.CallToolResult, any, error) {
	serviceID, err := s.serviceID(input.Service)
	if err != nil {
		return nil, nil, err
	}

	const endpointFields = `endpoints { host protocol type tls redirectTo dns { status requiredRecords { type host value } } }`

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

	mutation := `mutation($service: ServiceID!, $hostname: String!, $redirectTo: String) { addCustomDomain(service: $service, hostname: $hostname, redirectTo: $redirectTo) { id ` + endpointFields + ` } }`
	variables := map[string]any{"service": serviceID, "hostname": input.Hostname}
	if input.RedirectTo != "" {
		variables["redirectTo"] = input.RedirectTo
	}
	var out struct {
		AddCustomDomain any `json:"addCustomDomain"`
	}
	if err := s.query(ctx, "add_domain", mutation, variables, &out); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{
		"service": out.AddCustomDomain,
		"note":    "create the DNS records shown under dns.requiredRecords for every endpoint (ALIAS needs a DNS provider with ALIAS, ANAME or CNAME flattening support), then poll get_deploy_status for TLS to become ACTIVE. When redirectTo is set, report the redirect target as the live URL",
	})
}
