package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/ids"
)

const varsUsage = `lucity vars — manage service variables

Usage:
  lucity vars list <service> [--json]
  lucity vars available <env> [--json]
  lucity vars set <service> KEY=VALUE... [--ref KEY=<variableID>]... [--json]

Arguments:
  <service>  Service id (workspace/project/environment/service) or its relative form.
  <env>      Environment id (workspace/project/environment) or its relative form.

Flags:
  --ref KEY=<variableID>   Bind KEY to an available variable (list them with 'lucity vars available')
  --json                   Emit the result as JSON on stdout

'vars set' is non-destructive: it merges your changes into the existing variables
and re-sends the complete set, so variables you do not name are preserved.

Requires a signed-in member session (run 'lucity login'); a CI deploy token is
not sufficient for variable operations.
`

const serviceVariablesQuery = `query($service: ServiceID!) {
  serviceVariables(service: $service) { key value ref }
}`

type serviceVariableView struct {
	Key   string  `json:"key"`
	Value *string `json:"value"`
	Ref   *string `json:"ref"`
}

type availableVariableView struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Source struct {
		Typename string `json:"__typename"`
		Name     string `json:"name"`
	} `json:"source"`
}

func cmdVars(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, varsUsage)
		return errors.New("missing vars subcommand")
	}
	rest := args[1:]
	switch args[0] {
	case "list":
		return varsList(ctx, rest)
	case "available":
		return varsAvailable(ctx, rest)
	case "set":
		return varsSet(ctx, rest)
	case "help", "--help", "-h":
		fmt.Print(varsUsage)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown vars subcommand %q\n\n%s", args[0], varsUsage)
		return errors.New("unknown vars subcommand")
	}
}

func varsList(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("vars list", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, varsUsage) }
	asJSON := flags.Bool("json", false, "emit the variables as JSON")

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

	variables, err := readServiceVariables(ctx, client, serviceID)
	if err != nil {
		return deployError("vars list", err)
	}

	if *asJSON {
		return printJSON(variables)
	}
	return printServiceVariables(variables)
}

func varsAvailable(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("vars available", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, varsUsage) }
	asJSON := flags.Bool("json", false, "emit the variables as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <env> argument")
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	environmentID, err := ids.Environment(workspace, positionals[0])
	if err != nil {
		return err
	}

	const query = `query($environment: EnvironmentID!) {
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
		AvailableVariables []availableVariableView `json:"availableVariables"`
	}
	if err := client.GraphQL(ctx, query, map[string]any{"environment": environmentID}, &out); err != nil {
		return deployError("vars available", err)
	}
	variables := out.AvailableVariables

	if *asJSON {
		return printJSON(variables)
	}
	if len(variables) == 0 {
		fmt.Fprintln(os.Stderr, "No available variables in "+environmentID)
		return nil
	}
	writer := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "KEY\tSOURCE\tID")
	for _, variable := range variables {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", variable.Key, sourceLabel(variable.Source.Typename, variable.Source.Name), variable.ID)
	}
	return writer.Flush()
}

func varsSet(ctx context.Context, args []string) error {
	rest := args
	var serviceArg string
	if len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		serviceArg = rest[0]
		rest = rest[1:]
	}

	asJSON := false
	literals := map[string]string{}
	refs := map[string]string{}
	var order []string
	seen := map[string]bool{}
	appendKey := func(key string) {
		if !seen[key] {
			seen[key] = true
			order = append(order, key)
		}
	}

	for index := 0; index < len(rest); index++ {
		token := rest[index]
		switch {
		case token == "--json" || token == "-json":
			asJSON = true
		case token == "--ref" || token == "-ref":
			if index+1 >= len(rest) {
				return errors.New("--ref requires KEY=<variableID>")
			}
			index++
			key, value, err := splitAssignment(rest[index], "--ref")
			if err != nil {
				return err
			}
			refs[key] = value
			appendKey(key)
		case strings.HasPrefix(token, "--ref="):
			key, value, err := splitAssignment(strings.TrimPrefix(token, "--ref="), "--ref")
			if err != nil {
				return err
			}
			refs[key] = value
			appendKey(key)
		case strings.HasPrefix(token, "-"):
			fmt.Fprint(os.Stderr, varsUsage)
			return fmt.Errorf("unknown flag %q", token)
		default:
			key, value, err := splitAssignment(token, "variable")
			if err != nil {
				return err
			}
			literals[key] = value
			appendKey(key)
		}
	}

	if serviceArg == "" {
		fmt.Fprint(os.Stderr, varsUsage)
		return errors.New("missing <service> argument")
	}
	if len(literals) == 0 && len(refs) == 0 {
		fmt.Fprint(os.Stderr, varsUsage)
		return errors.New("no changes given — pass at least one KEY=VALUE or --ref KEY=<variableID>")
	}
	for key := range literals {
		if _, clash := refs[key]; clash {
			return fmt.Errorf("variable %q: set either value or ref, not both", key)
		}
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	serviceID, err := ids.Service(workspace, serviceArg)
	if err != nil {
		return err
	}

	current, err := readServiceVariables(ctx, client, serviceID)
	if err != nil {
		return deployError("vars set (read current)", err)
	}

	type mergedEntry struct {
		value *string
		ref   *string
	}
	merged := map[string]mergedEntry{}
	var mergedOrder []string
	mergedSeen := map[string]bool{}
	appendMerged := func(key string) {
		if !mergedSeen[key] {
			mergedSeen[key] = true
			mergedOrder = append(mergedOrder, key)
		}
	}
	for _, variable := range current {
		merged[variable.Key] = mergedEntry{value: variable.Value, ref: variable.Ref}
		appendMerged(variable.Key)
	}
	usedRef := false
	for _, key := range order {
		if reference, ok := refs[key]; ok {
			value := reference
			merged[key] = mergedEntry{ref: &value}
			usedRef = true
		} else {
			value := literals[key]
			merged[key] = mergedEntry{value: &value}
		}
		appendMerged(key)
	}

	variables := make([]map[string]any, 0, len(mergedOrder))
	for _, key := range mergedOrder {
		entry := merged[key]
		item := map[string]any{"key": key}
		switch {
		case entry.ref != nil:
			item["ref"] = *entry.ref
		case entry.value != nil:
			item["value"] = *entry.value
		default:
			item["value"] = ""
		}
		variables = append(variables, item)
	}

	const mutation = `mutation($service: ServiceID!, $variables: [ServiceVariableInput!]!) {
  setServiceVariables(service: $service, variables: $variables)
}`
	var writeOut struct {
		SetServiceVariables bool `json:"setServiceVariables"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"service": serviceID, "variables": variables}, &writeOut); err != nil {
		if usedRef {
			return fmt.Errorf("%w — a ref must be an id from `lucity vars available`", deployError("vars set", err))
		}
		return deployError("vars set", err)
	}

	updated, err := readServiceVariables(ctx, client, serviceID)
	if err != nil {
		return deployError("vars set (read back)", err)
	}

	fmt.Fprintf(os.Stderr, "Updated variables for %s\n", serviceID)
	if asJSON {
		return printJSON(updated)
	}
	return printServiceVariables(updated)
}

func readServiceVariables(ctx context.Context, client *api.Client, serviceID string) ([]serviceVariableView, error) {
	var out struct {
		ServiceVariables []serviceVariableView `json:"serviceVariables"`
	}
	if err := client.GraphQL(ctx, serviceVariablesQuery, map[string]any{"service": serviceID}, &out); err != nil {
		return nil, err
	}
	return out.ServiceVariables, nil
}

func printServiceVariables(variables []serviceVariableView) error {
	if len(variables) == 0 {
		fmt.Fprintln(os.Stderr, "No variables set.")
		return nil
	}
	writer := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "KEY\tVALUE")
	for _, variable := range variables {
		fmt.Fprintf(writer, "%s\t%s\n", variable.Key, variableDisplay(variable))
	}
	return writer.Flush()
}

func variableDisplay(variable serviceVariableView) string {
	if variable.Ref != nil {
		return "→ " + *variable.Ref
	}
	if variable.Value != nil {
		return *variable.Value
	}
	return ""
}

func sourceLabel(typename, name string) string {
	kind := strings.TrimSuffix(typename, "Source")
	if name == "" {
		return kind
	}
	return kind + " " + name
}

func splitAssignment(token, context string) (string, string, error) {
	parts := strings.SplitN(token, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", fmt.Errorf("%s must be KEY=VALUE, got %q", context, token)
	}
	return parts[0], parts[1], nil
}
