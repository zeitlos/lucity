package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/ids"
	"github.com/zeitlos/lucity/cli/internal/session"
)

const statusUsage = `lucity status — show the latest rollout status of a service

Usage:
  lucity status <service> [--json]

Arguments:
  <service>   Full service id (workspace/project/environment/service), or the
              project/environment/service form relative to the active workspace.

Flags:
  --json   Emit the service and its active deployment as JSON on stdout

Exit status is non-zero when the active rollout is FAILED or DEGRADED, so CI can
gate on it.

Examples:
  lucity status site/production/web
  lucity status acme/site/production/web --json
`

const statusQuery = `query($id: ServiceID!) {
  service(id: $id) {
    id status
    activeDeployment {
      id status image commit ref
      replicas { desired ready }
      rollout { status reason message restarts startedAt }
    }
  }
}`

type statusView struct {
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	ActiveDeployment *deploymentView `json:"activeDeployment"`
}

type deploymentView struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Image    string `json:"image"`
	Commit   string `json:"commit"`
	Ref      string `json:"ref"`
	Replicas struct {
		Desired int `json:"desired"`
		Ready   int `json:"ready"`
	} `json:"replicas"`
	Rollout *struct {
		Status    string `json:"status"`
		Reason    string `json:"reason"`
		Message   string `json:"message"`
		Restarts  int    `json:"restarts"`
		StartedAt string `json:"startedAt"`
	} `json:"rollout"`
}

func resolveClient(ctx context.Context) (*api.Client, string, error) {
	manager, err := session.Load()
	if err != nil {
		return nil, "", err
	}
	if err := manager.Prepare(ctx); err != nil {
		return nil, "", err
	}
	workspace := manager.Workspace()
	if workspace == "" {
		return nil, "", errors.New("no active workspace — run `lucity workspace <id>` (list with `lucity account`)")
	}
	return manager.Client(), workspace, nil
}

func positionalArgs(flags *flag.FlagSet, args []string) ([]string, error) {
	var positionals []string
	rest := args
	for len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		positionals = append(positionals, rest[0])
		rest = rest[1:]
	}
	if err := flags.Parse(rest); err != nil {
		return nil, err
	}
	return append(positionals, flags.Args()...), nil
}

func cmdStatus(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("status", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, statusUsage) }
	asJSON := flags.Bool("json", false, "emit the service as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <service> argument")
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	serviceID, err := ids.Service(workspace, positionals[0])
	if err != nil {
		return err
	}

	var out struct {
		Service statusView `json:"service"`
	}
	if err := client.GraphQL(ctx, statusQuery, map[string]any{"id": serviceID}, &out); err != nil {
		return deployError("status", err)
	}
	service := out.Service

	if *asJSON {
		if err := printJSON(service); err != nil {
			return err
		}
	} else {
		printStatus(service)
	}

	if deployment := service.ActiveDeployment; deployment != nil && deployment.Rollout != nil {
		switch deployment.Rollout.Status {
		case "FAILED", "DEGRADED":
			return fmt.Errorf("rollout %s: %s", strings.ToLower(deployment.Rollout.Status), rolloutReasonHint(deployment.Rollout.Reason))
		}
	}
	return nil
}

func printStatus(service statusView) {
	fmt.Printf("Service %s: %s\n", service.ID, service.Status)
	deployment := service.ActiveDeployment
	if deployment == nil {
		fmt.Println("No active deployment yet.")
		return
	}
	fmt.Printf("Deployment %s: %s\n", deployment.ID, deployment.Status)
	fmt.Printf("Replicas:   %d/%d ready\n", deployment.Replicas.Ready, deployment.Replicas.Desired)
	if source := strings.TrimSpace(deployment.Ref + " " + shortSHA(deployment.Commit)); source != "" {
		fmt.Printf("Source:     %s\n", source)
	}
	rollout := deployment.Rollout
	if rollout == nil {
		return
	}
	fmt.Printf("Rollout:    %s", rollout.Status)
	if rollout.Restarts > 0 {
		fmt.Printf(" (restarts: %d)", rollout.Restarts)
	}
	fmt.Println()
	if rollout.Status == "DEGRADED" || rollout.Status == "FAILED" {
		fmt.Printf("Reason:     %s\n", rolloutReasonHint(rollout.Reason))
	}
	if rollout.Message != "" {
		fmt.Printf("Message:    %s\n", rollout.Message)
	}
}
