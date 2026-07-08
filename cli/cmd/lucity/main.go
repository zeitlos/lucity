package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/authflow"
	"github.com/zeitlos/lucity/cli/internal/mcpserver"
	"github.com/zeitlos/lucity/cli/internal/session"
)

var version = "dev"

const usage = `lucity — deploy software on your Lucity platform

Usage:
  lucity login [--api <url>]   Sign in through your browser
  lucity logout                Discard the stored session
  lucity account               Show identity and workspace memberships
  lucity workspace [<id>]      Show or switch the active workspace
  lucity deploy <service>      Build and roll out a service (--ref, --wait)
  lucity token                 Print a valid bearer token for scripting
  lucity mcp                   Serve the Lucity MCP server on stdio
  lucity version               Print the CLI version

Environment:
  LUCITY_API_URL    Override the platform URL (default: stored, then ` + api.DefaultBaseURL + `)
  LUCITY_WORKSPACE  Override the active workspace
  LUCITY_TOKEN      Use a fixed bearer token instead of the stored session
`

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "login":
		err = cmdLogin(ctx, os.Args[2:])
	case "logout":
		err = cmdLogout()
	case "account":
		err = cmdAccount(ctx)
	case "workspace":
		err = cmdWorkspace(ctx, os.Args[2:])
	case "deploy":
		err = cmdDeploy(ctx, os.Args[2:])
	case "token":
		err = cmdToken(ctx)
	case "mcp":
		err = cmdMCP(ctx)
	case "version", "--version", "-v":
		fmt.Println("lucity " + version)
	case "help", "--help", "-h":
		fmt.Print(usage)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdLogin(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("login", flag.ExitOnError)
	apiURL := flags.String("api", "", "platform URL (e.g. https://lucity.cloud)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	manager, err := session.Load()
	if err != nil {
		return err
	}
	base := manager.APIURL()
	if *apiURL != "" {
		base = *apiURL
	}

	newSession, err := authflow.Login(ctx, base)
	if err != nil {
		return err
	}
	if err := manager.SetSession(base, newSession); err != nil {
		return err
	}

	identity, err := manager.Client().Me(ctx)
	if err != nil {
		return fmt.Errorf("signed in, but fetching the account failed: %w", err)
	}

	if manager.Workspace() == "" && len(identity.Workspaces) > 0 {
		if err := manager.SetWorkspace(identity.Workspaces[0].Workspace); err != nil {
			return err
		}
	}

	fmt.Printf("Signed in as %s (%s) on %s\n", identity.Name, identity.Email, base)
	fmt.Printf("Active workspace: %s\n", manager.Workspace())
	return nil
}

func cmdLogout() error {
	manager, err := session.Load()
	if err != nil {
		return err
	}
	if err := manager.Clear(); err != nil {
		return err
	}
	fmt.Println("Signed out.")
	return nil
}

func cmdAccount(ctx context.Context) error {
	manager, err := session.Load()
	if err != nil {
		return err
	}
	identity, err := manager.Client().Me(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%s <%s>\n", identity.Name, identity.Email)
	fmt.Printf("Platform: %s\n\n", manager.APIURL())
	writer := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	fmt.Fprintln(writer, "WORKSPACE\tROLE\tACTIVE")
	for _, membership := range identity.Workspaces {
		active := ""
		if membership.Workspace == manager.Workspace() {
			active = "*"
		}
		fmt.Fprintf(writer, "%s\t%s\t%s\n", membership.Workspace, membership.Role, active)
	}
	return writer.Flush()
}

func cmdWorkspace(ctx context.Context, args []string) error {
	manager, err := session.Load()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		if manager.Workspace() == "" {
			fmt.Println("No active workspace. Set one with `lucity workspace <id>`.")
			return nil
		}
		fmt.Println(manager.Workspace())
		return nil
	}

	target := args[0]
	identity, err := manager.Client().Me(ctx)
	if err != nil {
		return err
	}
	for _, membership := range identity.Workspaces {
		if membership.Workspace == target {
			if err := manager.SetWorkspace(target); err != nil {
				return err
			}
			fmt.Printf("Active workspace: %s\n", target)
			return nil
		}
	}
	return fmt.Errorf("you are not a member of workspace %q — run `lucity account` to list memberships", target)
}

func cmdToken(ctx context.Context) error {
	manager, err := session.Load()
	if err != nil {
		return err
	}
	token, err := manager.Token(ctx)
	if err != nil {
		return err
	}
	fmt.Println(token)
	return nil
}

func cmdMCP(ctx context.Context) error {
	manager, err := session.Load()
	if err != nil {
		return err
	}
	return mcpserver.Serve(ctx, manager, version)
}
