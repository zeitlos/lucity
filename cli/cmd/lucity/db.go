package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/ids"
)

const dbUsage = `lucity db — manage databases

Usage:
  lucity db create <env> <name> [--id <id>] [--size <size>] [--cpu <cpu>] [--memory <mem>] [--json]
  lucity db list <env> [--json]
  lucity db credentials <db> [--json]
  lucity db query <db> [<sql>] [--file <path>] [--json]
  lucity db expose <db> [--json]
  lucity db unexpose <db> [--json]
  lucity db delete <db> [--yes] [--json]

Commands:
  create        Create a PostgreSQL database in an environment
  list          List the databases in an environment
  credentials   Print the connection details, password included
  query         Run one SQL statement and print its rows, at most 1,000
  expose        Give the database a public hostname, reachable over TLS from
                this machine's address only
  unexpose      Remove the public hostname and its allowed addresses again
  delete        Delete the database and all of its data

Arguments:
  <env>    Environment id (workspace/project/environment) or its relative form.
  <name>   What people read, any text. Quote it when it has spaces.
  <db>     Database id (workspace/project/environment/database) or its relative form.

Flags:
  --id <id>       Id for a new database, 2-16 lowercase letters, digits and
                  hyphens. Derived from <name> when left out.
  --size <size>   Storage size for a new database (e.g. 32Gi)
  --cpu <cpu>     CPU limit for a new database, e.g. 500m (with --memory)
  --memory <mem>  Memory limit for a new database, e.g. 512Mi (with --cpu)
  --file <path>   Read the SQL for query from a file, or from stdin with -
  --yes           Skip the confirmation prompt on delete
  --json          Emit the result as JSON on stdout

'lucity db query' runs the statement as the database's own user and commits it
right away. Values come back as text, NULL as null in --json. For more than
1,000 rows, page with LIMIT and OFFSET, or expose the database and use pg_dump.

Examples:
  lucity db query shop/production/main "select id, email from users limit 10"
  lucity db query shop/production/main --file report.sql --json | jq '.rows'

Needs a signed-in session or a workspace API token. A keyless GitHub Actions
session cannot manage databases.
`

type databaseView struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	Size      string         `json:"size"`
	Public    bool           `json:"public"`
	Resources *resourcesView `json:"resources,omitempty"`
}

type resourcesView struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
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
	case "query":
		return dbQuery(ctx, rest)
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
	id := flags.String("id", "", "database id (derived from the name when left out)")
	size := flags.String("size", "", "storage size (e.g. 32Gi)")
	cpu := flags.String("cpu", "", "CPU limit (e.g. 500m)")
	memory := flags.String("memory", "", "memory limit (e.g. 512Mi)")
	asJSON := flags.Bool("json", false, "emit the database as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) < 2 {
		flags.Usage()
		return errors.New("usage: lucity db create <env> <name> [--id <id>] [--size <size>] [--cpu <cpu>] [--memory <mem>]")
	}
	if (*cpu == "") != (*memory == "") {
		return errors.New("--cpu and --memory must be provided together")
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
	if *id != "" {
		input["id"] = *id
	}
	if *size != "" {
		input["size"] = *size
	}
	if *cpu != "" {
		input["resources"] = map[string]any{"cpu": *cpu, "memory": *memory}
	}
	const mutation = `mutation($input: CreateDatabaseInput!) {
  createDatabase(input: $input) { id name status size public resources { cpu memory } }
}`
	var out struct {
		CreateDatabase databaseView `json:"createDatabase"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"input": input}, &out); err != nil {
		return deployError("db create", err)
	}
	database := out.CreateDatabase

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

const maxQueryRows = 1000

type queryResultView struct {
	Columns      []string    `json:"columns"`
	Rows         [][]*string `json:"rows"`
	AffectedRows int         `json:"affectedRows"`
}

func dbQuery(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("db query", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, dbUsage) }
	file := flags.String("file", "", "read the SQL from a file, or stdin with -")
	asJSON := flags.Bool("json", false, "emit the result as JSON")

	positionals, err := positionalArgs(flags, args)
	if err != nil {
		return err
	}
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("missing <db> argument")
	}
	statement, err := queryStatement(positionals[1:], *file)
	if err != nil {
		return err
	}

	client, workspace, err := resolveClient(ctx)
	if err != nil {
		return err
	}
	databaseID, err := ids.Resource(workspace, positionals[0])
	if err != nil {
		return err
	}

	const mutation = `mutation($database: DatabaseID!, $query: String!) {
  executeQuery(database: $database, query: $query) { columns rows affectedRows }
}`
	var out struct {
		ExecuteQuery queryResultView `json:"executeQuery"`
	}
	if err := client.GraphQL(ctx, mutation, map[string]any{"database": databaseID, "query": statement}, &out); err != nil {
		return deployError("db query", err)
	}
	result := out.ExecuteQuery

	if len(result.Rows) >= maxQueryRows {
		fmt.Fprintf(os.Stderr, "Showing the first %d rows; page with LIMIT and OFFSET for more.\n", maxQueryRows)
	}
	if *asJSON {
		return printJSON(result)
	}
	if len(result.Columns) == 0 {
		fmt.Printf("%d rows affected\n", result.AffectedRows)
		return nil
	}
	writer := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, strings.Join(result.Columns, "\t"))
	for _, row := range result.Rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			if cell == nil {
				cells[i] = "NULL"
				continue
			}
			cells[i] = strings.NewReplacer("\n", `\n`, "\t", `\t`).Replace(*cell)
		}
		fmt.Fprintln(writer, strings.Join(cells, "\t"))
	}
	return writer.Flush()
}

func queryStatement(arguments []string, file string) (string, error) {
	switch {
	case file != "" && len(arguments) > 0:
		return "", errors.New("pass the SQL either as an argument or with --file, not both")
	case file == "-":
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(content), nil
	case file != "":
		content, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(content), nil
	case len(arguments) == 1:
		return arguments[0], nil
	case len(arguments) > 1:
		return "", errors.New("pass the SQL as one quoted argument")
	default:
		return "", errors.New("missing SQL: pass it as an argument or with --file")
	}
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
	result := exposureView{ID: out[field].ID, Public: out[field].Public}
	if result.ID == "" {
		result.ID = databaseID
	}

	if public {
		result.Allowed, err = allowCaller(ctx, client, databaseID)
		if err != nil {
			return deployError(operation, err)
		}
	}

	if *asJSON {
		return printJSON(result)
	}
	state := "private"
	if result.Public {
		state = "public"
	}
	fmt.Printf("Database %s is now %s.\n", result.ID, state)
	if public && result.Allowed != "" {
		fmt.Printf("Only this machine (%s) can connect. Allow more addresses in the dashboard.\n", result.Allowed)
	} else if public {
		fmt.Println("Nobody can connect yet: this machine's address is unknown. Allow addresses in the dashboard.")
	}
	return nil
}

type exposureView struct {
	ID      string `json:"id"`
	Public  bool   `json:"public"`
	Allowed string `json:"allowed,omitempty"`
}

func allowCaller(ctx context.Context, client *api.Client, databaseID string) (string, error) {
	var out struct {
		ClientAddress *string `json:"clientAddress"`
	}
	if err := client.GraphQL(ctx, `query { clientAddress }`, nil, &out); err != nil {
		return "", err
	}
	if out.ClientAddress == nil {
		return "", nil
	}

	const mutation = `mutation($database: DatabaseID!, $rule: DatabaseAllowRuleInput!) {
  addDatabaseAllowRule(database: $database, rule: $rule) { id }
}`
	rule := map[string]any{"range": *out.ClientAddress, "description": "Added by lucity db expose"}
	if err := client.GraphQL(ctx, mutation, map[string]any{"database": databaseID, "rule": rule}, nil); err != nil {
		return "", err
	}
	return *out.ClientAddress, nil
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
	if database.Resources != nil {
		fmt.Printf("  cpu:    %s\n", database.Resources.CPU)
		fmt.Printf("  memory: %s\n", database.Resources.Memory)
	}
}
