package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zeitlos/lucity/cli/internal/ids"
)

const dbUsage = `lucity db — manage databases

Usage:
  lucity db create <env> <name> [--size <size>] [--json]
  lucity db list <env> [--json]
  lucity db credentials <db> [--json]
  lucity db expose <db> [--json]
  lucity db unexpose <db> [--json]
  lucity db delete <db> [--yes] [--json]

Arguments:
  <env>    Environment id (workspace/project/environment) or its relative form.
  <name>   Database name (2-16 chars).
  <db>     Database id (workspace/project/environment/name) or its relative form.

Flags:
  --size <size>   Storage size for a new database (e.g. 32Gi)
  --yes           Skip the confirmation prompt on delete
  --json          Emit the result as JSON on stdout

Requires a signed-in member session (run 'lucity login'); a CI deploy token is
not sufficient for database operations.
`

type databaseView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Size   string `json:"size"`
	Public bool   `json:"public"`
}

type credentialsView struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Dbname   string `json:"dbname"`
	User     string `json:"user"`
	Password string `json:"password"`
	URI      string `json:"uri"`
}

func cmdDB(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, dbUsage)
		return errors.New("missing db subcommand")
	}
	rest := args[1:]
	switch args[0] {
	case "create":
		return dbCreate(ctx, rest)
	case "list":
		return dbList(ctx, rest)
	case "credentials":
		return dbCredentials(ctx, rest)
	case "expose":
		return setDatabaseExposure(ctx, rest, true)
	case "unexpose":
		return setDatabaseExposure(ctx, rest, false)
	case "delete":
		return dbDelete(ctx, rest)
	case "help", "--help", "-h":
		fmt.Print(dbUsage)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown db subcommand %q\n\n%s", args[0], dbUsage)
		return errors.New("unknown db subcommand")
	}
}

func dbCreate(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("db create", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	size := flags.String("size", "", "storage size (e.g. 32Gi)")
	asJSON := flags.Bool("json", false, "emit the database as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) < 2 {
		flags.Usage()
		return errors.New("usage: lucity db create <env> <name> [--size <size>]")
	}
	envArg, nameArg := positionals[0], positionals[1]

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	environmentID, err := ids.Environment(workspace, envArg)
	if err != nil {
		return err
	}

	input := map[string]any{"environment": environmentID, "name": nameArg}
	if *size != "" {
		input["size"] = *size
	}
	const mutation = `mutation($input: CreateDatabaseInput!) {
  createDatabase(input: $input) { id name status size public }
}`
	var out struct {
		CreateDatabase databaseView `json:"createDatabase"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"input": input}, &out); err != nil {
		return deployError("db create", err)
	}
	database := out.CreateDatabase
	if strings.Trim(database.ID, "/") == "" {
		database.ID = environmentID + "/" + nameArg
	}

	if *asJSON {
		return printJSON(database)
	}
	fmt.Printf("Created database %s\n", database.ID)
	printDatabase(database)
	return nil
}

func dbList(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("db list", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	asJSON := flags.Bool("json", false, "emit the databases as JSON")

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
  environment(environment: $environment) { databases { id name status size public } }
}`
	var out struct {
		Environment struct {
			Databases []databaseView `json:"databases"`
		} `json:"environment"`
	}
	if err := client.GraphQL(ctx, query, map[string]any{"environment": environmentID}, &out); err != nil {
		return deployError("db list", err)
	}
	databases := out.Environment.Databases

	if *asJSON {
		return printJSON(databases)
	}
	if len(databases) == 0 {
		fmt.Fprintln(os.Stderr, "No databases in "+environmentID)
		return nil
	}
	writer := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tSTATUS\tSIZE\tPUBLIC\tID")
	for _, database := range databases {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%t\t%s\n", database.Name, database.Status, database.Size, database.Public, database.ID)
	}
	return writer.Flush()
}

func dbCredentials(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("db credentials", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	asJSON := flags.Bool("json", false, "emit the credentials as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <db> argument")
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	databaseID, err := ids.Resource(workspace, positionals[0])
	if err != nil {
		return err
	}

	const query = `query($database: DatabaseID!) {
  databaseCredentials(database: $database) { type host port dbname user password uri }
}`
	var out struct {
		DatabaseCredentials []credentialsView `json:"databaseCredentials"`
	}
	if err := client.GraphQL(ctx, query, map[string]any{"database": databaseID}, &out); err != nil {
		return deployError("db credentials", err)
	}

	if *asJSON {
		return printJSON(out.DatabaseCredentials)
	}
	for index, credentials := range out.DatabaseCredentials {
		if index > 0 {
			fmt.Println()
		}
		fmt.Println(credentials.Type)
		fmt.Printf("  host:     %s\n", credentials.Host)
		fmt.Printf("  port:     %s\n", credentials.Port)
		fmt.Printf("  dbname:   %s\n", credentials.Dbname)
		fmt.Printf("  user:     %s\n", credentials.User)
		fmt.Printf("  password: %s\n", credentials.Password)
		fmt.Printf("  uri:      %s\n", credentials.URI)
	}
	return nil
}

func setDatabaseExposure(ctx context.Context, args []string, public bool) error {
	operation := "db unexpose"
	field := "unexposeDatabase"
	mutation := `mutation($database: DatabaseID!) { unexposeDatabase(database: $database) { id public } }`
	if public {
		operation = "db expose"
		field = "exposeDatabase"
		mutation = `mutation($database: DatabaseID!) { exposeDatabase(database: $database) { id public } }`
	}

	flags := flag.NewFlagSet(operation, flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	asJSON := flags.Bool("json", false, "emit the database as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <db> argument")
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	databaseID, err := ids.Resource(workspace, positionals[0])
	if err != nil {
		return err
	}

	var out map[string]struct {
		ID     string `json:"id"`
		Public bool   `json:"public"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"database": databaseID}, &out); err != nil {
		return deployError(operation, err)
	}
	result := out[field]
	if result.ID == "" {
		result.ID = databaseID
	}

	if *asJSON {
		return printJSON(result)
	}
	state := "private"
	if result.Public {
		state = "public"
	}
	fmt.Printf("Database %s is now %s.\n", result.ID, state)
	return nil
}

func dbDelete(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("db delete", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	yes := flags.Bool("yes", false, "skip the confirmation prompt")
	asJSON := flags.Bool("json", false, "emit the result as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <db> argument")
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	databaseID, err := ids.Resource(workspace, positionals[0])
	if err != nil {
		return err
	}

	if !*yes {
		fmt.Fprintf(os.Stderr, "Delete database %s? This cannot be undone. [y/N]: ", databaseID)
		answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "y", "yes":
		default:
			return errors.New("aborted")
		}
	}

	const mutation = `mutation($database: DatabaseID!) { deleteDatabase(database: $database) }`
	var out struct {
		DeleteDatabase bool `json:"deleteDatabase"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"database": databaseID}, &out); err != nil {
		return deployError("db delete", err)
	}
	if !out.DeleteDatabase {
		return fmt.Errorf("db delete failed: the platform reported that database %s was not deleted", databaseID)
	}

	if *asJSON {
		return printJSON(map[string]any{"id": databaseID, "deleted": true})
	}
	fmt.Printf("Deleted database %s\n", databaseID)
	return nil
}

func printDatabase(database databaseView) {
	fmt.Printf("  name:   %s\n", database.Name)
	fmt.Printf("  status: %s\n", database.Status)
	fmt.Printf("  size:   %s\n", database.Size)
	fmt.Printf("  public: %t\n", database.Public)
}
