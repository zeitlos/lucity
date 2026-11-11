package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/zeitlos/lucity/cli/internal/ids"
)

const restartUsage = `lucity restart — replace the instances of a service with fresh ones

Usage:
  lucity restart <service>

Arguments:
  <service>   Full service id (workspace/project/environment/service), or the
              project/environment/service form relative to the active workspace.

The service keeps the image and settings it already runs; nothing is built.
Follow the rollout with 'lucity status <service>'.

Examples:
  lucity restart site/production/web
`

func cmdRestart(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("restart", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, restartUsage) }

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

	const mutation = `mutation($service: ServiceID!) { restartService(service: $service) { id } }`
	if err := client.GraphQL(ctx, mutation, map[string]any{"service": serviceID}, nil); err != nil {
		return deployError("restart", err)
	}
	fmt.Fprintf(os.Stderr, "restarting %s; follow it with 'lucity status %s'\n", serviceID, positionals[0])
	return nil
}
