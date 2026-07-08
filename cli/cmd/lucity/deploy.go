package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/ids"
	"github.com/zeitlos/lucity/cli/internal/session"
)

const deployUsage = `lucity deploy — build and roll out a service

Usage:
  lucity deploy <service> [flags]

Arguments:
  <service>   Full service id (workspace/project/environment/service), or the
              project/environment/service form relative to the active workspace.

Flags:
  --ref <ref>       Branch, tag, or commit to build (default: the service's branch)
  --wait            Block until the release is live or failed; exit non-zero on failure
  --timeout <dur>   Give up waiting after this long (default 20m; only with --wait)
  --interval <dur>  Poll interval while waiting (default 4s)
  --json            Emit the final release as JSON on stdout

Examples:
  lucity deploy site/production/web --ref "$GITHUB_SHA" --wait
  lucity deploy acme/site/production/web --ref v1.4.2 --wait --timeout 30m

Authentication:
  In GitHub Actions, deploys are keyless — no stored secret. Grant the job
  "permissions: id-token: write" and enable CI deploys for the service in its
  Lucity settings; the CLI exchanges the job's OIDC token for a short-lived,
  deploy-only session automatically. The workspace is inferred, so
  LUCITY_WORKSPACE is optional.

  For other CI systems, set LUCITY_TOKEN (and LUCITY_WORKSPACE to pin the
  workspace).
`

const deployStatusQuery = `query($id: ServiceID!) {
  service(id: $id) {
    status
    endpoints { host protocol type tls }
    releases {
      id status
      source { ref commit { sha message } }
      build { id status }
      deploy { id status }
      deployment { id status replicas { desired ready } rollout { status reason message restarts } }
    }
  }
}`

type releaseView struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Source *struct {
		Ref    string `json:"ref"`
		Commit struct {
			SHA     string `json:"sha"`
			Message string `json:"message"`
		} `json:"commit"`
	} `json:"source"`
	Build *struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"build"`
	Deploy *struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	} `json:"deploy"`
	Deployment *struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Rollout *struct {
			Status   string `json:"status"`
			Reason   string `json:"reason"`
			Message  string `json:"message"`
			Restarts int    `json:"restarts"`
		} `json:"rollout"`
	} `json:"deployment"`
}

type endpointView struct {
	Host     string `json:"host"`
	Protocol string `json:"protocol"`
	Type     string `json:"type"`
	TLS      string `json:"tls"`
}

type serviceView struct {
	Status    string         `json:"status"`
	Endpoints []endpointView `json:"endpoints"`
	Releases  []releaseView  `json:"releases"`
}

func cmdDeploy(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("deploy", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, deployUsage) }
	ref := flags.String("ref", "", "branch, tag, or commit to build")
	wait := flags.Bool("wait", false, "block until the release is live or failed")
	timeout := flags.Duration("timeout", 20*time.Minute, "give up waiting after this long")
	interval := flags.Duration("interval", 4*time.Second, "poll interval while waiting")
	asJSON := flags.Bool("json", false, "emit the final release as JSON")

	rest := args
	var serviceArg string
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		serviceArg = rest[0]
		rest = rest[1:]
	}
	if err := flags.Parse(rest); err != nil {
		return err
	}
	if serviceArg == "" {
		serviceArg = flags.Arg(0)
	}
	if serviceArg == "" {
		flags.Usage()
		return errors.New("missing <service> argument")
	}

	manager, err := session.Load()
	if err != nil {
		return err
	}
	if err := manager.Prepare(ctx); err != nil {
		return err
	}
	workspace := manager.Workspace()
	if workspace == "" {
		return errors.New("no active workspace — run `lucity workspace <id>` (list with `lucity account`)")
	}
	serviceID, err := ids.Service(workspace, serviceArg)
	if err != nil {
		return err
	}

	client := manager.Client()

	started, err := startDeploy(ctx, client, serviceID, *ref)
	if err != nil {
		return deployError("deploy", err)
	}

	fmt.Fprintf(os.Stderr, "Release %s queued for %s", started.ID, serviceID)
	if source := releaseRef(started); source != "" {
		fmt.Fprintf(os.Stderr, " (%s)", source)
	}
	fmt.Fprintln(os.Stderr)

	if !*wait {
		if *asJSON {
			return printJSON(started)
		}
		fmt.Println(started.ID)
		fmt.Fprintln(os.Stderr, "Deploy started. Re-run with --wait to block on the rollout, or watch it in the dashboard.")
		return nil
	}

	final, endpoints, err := waitForRelease(ctx, client, serviceID, started.ID, *timeout, *interval)
	if final != nil && *asJSON {
		if jsonErr := printJSON(*final); jsonErr != nil {
			return jsonErr
		}
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "✓ Release %s is live\n", final.ID)
	printEndpoints(endpoints)
	return nil
}

func startDeploy(ctx context.Context, client *api.Client, serviceID, ref string) (releaseView, error) {
	const mutation = `mutation($service: ServiceID!, $gitRef: String) {
  deploy(service: $service, gitRef: $gitRef) {
    id status
    source { ref commit { sha message } }
    build { id status }
  }
}`
	variables := map[string]any{"service": serviceID}
	if ref != "" {
		variables["gitRef"] = ref
	}

	var out struct {
		Deploy releaseView `json:"deploy"`
	}
	if err := client.GraphQL(ctx, mutation, variables, &out); err != nil {
		return releaseView{}, err
	}
	return out.Deploy, nil
}

func waitForRelease(ctx context.Context, client *api.Client, serviceID, releaseID string, timeout, interval time.Duration) (*releaseView, []string, error) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	lastStatus := ""
	lastRollout := ""
	pollErrors := 0

	for {
		service, err := fetchService(ctx, client, serviceID)
		if err != nil {
			pollErrors++
			if pollErrors >= 5 {
				return nil, nil, deployError("status", err)
			}
		} else {
			pollErrors = 0
			release := findRelease(service.Releases, releaseID)
			if release == nil {
				return nil, nil, fmt.Errorf("release %s vanished from service %s", releaseID, serviceID)
			}

			if release.Status != lastStatus {
				fmt.Fprintf(os.Stderr, "→ %s\n", strings.ToLower(release.Status))
				lastStatus = release.Status
			}
			if rollout := rolloutProgress(*release); rollout != "" && rollout != lastRollout {
				fmt.Fprintf(os.Stderr, "  %s\n", rollout)
				lastRollout = rollout
			}

			switch release.Status {
			case "LIVE":
				return release, endpointHosts(service.Endpoints), nil
			case "FAILED", "CANCELLED":
				return release, nil, fmt.Errorf("deploy failed: %s", failureReason(*release))
			case "SUPERSEDED":
				return release, nil, fmt.Errorf("release %s was superseded by a newer release before it went live", releaseID)
			}
		}

		if time.Now().After(deadline) {
			return nil, nil, fmt.Errorf("timed out after %s waiting for release %s to go live (last status: %s)", timeout, releaseID, strings.ToLower(orUnknown(lastStatus)))
		}

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func fetchService(ctx context.Context, client *api.Client, serviceID string) (*serviceView, error) {
	var out struct {
		Service serviceView `json:"service"`
	}
	if err := client.GraphQL(ctx, deployStatusQuery, map[string]any{"id": serviceID}, &out); err != nil {
		return nil, err
	}
	return &out.Service, nil
}

func findRelease(releases []releaseView, releaseID string) *releaseView {
	for i := range releases {
		if releases[i].ID == releaseID {
			return &releases[i]
		}
	}
	return nil
}

func rolloutProgress(release releaseView) string {
	if release.Deployment == nil || release.Deployment.Rollout == nil {
		return ""
	}
	rollout := release.Deployment.Rollout
	switch rollout.Status {
	case "PROGRESSING":
		if rollout.Restarts > 0 {
			return fmt.Sprintf("rollout progressing (restarts: %d)", rollout.Restarts)
		}
		return "rollout progressing"
	case "DEGRADED":
		return "rollout degraded: " + rolloutReasonHint(rollout.Reason)
	}
	return ""
}

func failureReason(release releaseView) string {
	if release.Build != nil && release.Build.Status == "FAILED" {
		return fmt.Sprintf("build failed (build %s) — inspect its logs in the dashboard", release.Build.ID)
	}
	if release.Deploy != nil && release.Deploy.Status == "FAILED" {
		return fmt.Sprintf("deploy failed (deploy %s) — inspect its logs in the dashboard", release.Deploy.ID)
	}
	if release.Deployment != nil && release.Deployment.Rollout != nil {
		rollout := release.Deployment.Rollout
		if rollout.Status == "DEGRADED" || rollout.Status == "FAILED" {
			return "rollout " + strings.ToLower(rollout.Status) + ": " + rolloutReasonHint(rollout.Reason)
		}
	}
	if release.Status == "CANCELLED" {
		return "the release was cancelled"
	}
	return "see the release in the dashboard for details"
}

func rolloutReasonHint(reason string) string {
	switch reason {
	case "OOM_KILLED":
		return "the container exceeded its memory limit — raise memory for the service"
	case "CRASH_LOOP":
		return "the container keeps crashing on startup — check the service's runtime logs"
	case "IMAGE_PULL_FAILED":
		return "the image could not be pulled — verify the image reference or that the build succeeded"
	case "CONFIG_ERROR":
		return "invalid configuration (e.g. a bad env var or reference) — review the service's variables"
	case "UNSCHEDULABLE":
		return "no node can schedule the pod — reduce requested resources or the environment is at capacity"
	case "QUOTA_EXCEEDED":
		return "the environment resource quota is exhausted — raise the quota or reduce service resources"
	case "DEADLINE_EXCEEDED":
		return "the rollout timed out becoming ready — check the service's runtime logs and readiness"
	case "NOT_READY":
		return "not ready yet"
	default:
		return "check the service's runtime logs in the dashboard"
	}
}

func releaseRef(release releaseView) string {
	if release.Source == nil {
		return ""
	}
	ref := release.Source.Ref
	sha := release.Source.Commit.SHA
	switch {
	case ref != "" && sha != "":
		return fmt.Sprintf("%s @ %s", ref, shortSHA(sha))
	case ref != "":
		return ref
	case sha != "":
		return shortSHA(sha)
	default:
		return ""
	}
}

func endpointHosts(endpoints []endpointView) []string {
	var hosts []string
	for _, endpoint := range endpoints {
		scheme := "http"
		if endpoint.TLS != "" && endpoint.TLS != "NONE" {
			scheme = "https"
		}
		hosts = append(hosts, scheme+"://"+endpoint.Host)
	}
	return hosts
}

func printEndpoints(hosts []string) {
	for _, host := range hosts {
		fmt.Fprintf(os.Stderr, "  %s\n", host)
	}
}

func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func deployError(operation string, err error) error {
	var requestErr *api.RequestError
	if errors.As(err, &requestErr) {
		lower := strings.ToLower(requestErr.Error())
		if strings.Contains(lower, "logto access token") || strings.Contains(lower, "github token") || strings.Contains(lower, "invalid_grant") {
			return fmt.Errorf("%s failed: the platform could not access the source repository — run `lucity login`, or in CI set LUCITY_LOGTO_TOKEN", operation)
		}
		if strings.Contains(lower, "not found") {
			return fmt.Errorf("%s failed: %s — check the service id; its first segment must be your active workspace", operation, requestErr.Error())
		}
		return fmt.Errorf("%s failed: %s", operation, requestErr.Error())
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}

func shortSHA(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}

func orUnknown(status string) string {
	if status == "" {
		return "unknown"
	}
	return status
}
