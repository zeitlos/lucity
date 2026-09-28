package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/zeitlos/lucity/cli/internal/api"
	"github.com/zeitlos/lucity/cli/internal/authflow"
	"github.com/zeitlos/lucity/cli/internal/mcpserver"
	"github.com/zeitlos/lucity/cli/internal/session"
	"github.com/zeitlos/lucity/pkg/oidc"
)

func decodeJWTClaims(token string) map[string]any {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	return claims
}

var version = "dev"

const usage = `lucity — deploy software on your Lucity platform

Usage:
  lucity login [--api <url>]   Sign in through your browser
  lucity logout                Discard the stored session
  lucity account               Show identity and workspace memberships
  lucity workspace [<id>]      Show or switch the active workspace
  lucity deploy <service>      Build and roll out a service (--ref, --wait)
  lucity db <cmd> <args>       Manage databases (create, list, credentials, expose, unexpose, delete)
  lucity vars <cmd> <args>     Manage service variables (list, available, set)
  lucity status <service>      Show the latest rollout status of a service
  lucity token [--account]     Print a valid bearer token for scripting
  lucity mcp                   Serve the Lucity MCP server on stdio
  lucity version               Print the CLI version

Environment:
  LUCITY_API_URL    Override the platform URL (default: stored, then ` + api.DefaultBaseURL + `)
  LUCITY_WORKSPACE  Override the active workspace
  LUCITY_API_TOKEN  Authenticate with a workspace API token (for CI and automation)
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
	case "db":
		err = cmdDB(ctx, os.Args[2:])
	case "vars":
		err = cmdVars(ctx, os.Args[2:])
	case "status":
		err = cmdStatus(ctx, os.Args[2:])
	case "token":
		err = cmdToken(ctx, os.Args[2:])
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

	httpClient := &http.Client{Timeout: 30 * time.Second}
	authCfg, err := api.Config(ctx, httpClient, base)
	if err != nil {
		return fmt.Errorf("fetch auth config from %s: %w", base, err)
	}
	if authCfg.CliClientID == "" {
		return errors.New("this platform has no CLI login configured — the maintainer must register a native CLI client and set OIDC_CLI_CLIENT_ID")
	}

	provider := &oidc.Provider{
		Endpoint:     authCfg.Endpoint,
		ClientID:     authCfg.CliClientID,
		Audience:     authCfg.Audience,
		DirectSignIn: session.DirectSignIn,
		Scopes:       session.LoginScopes,
		HTTP:         httpClient,
	}

	refreshToken, err := authflow.Login(ctx, provider)
	if err != nil {
		return err
	}
	if err := manager.SetLogin(base, refreshToken); err != nil {
		return err
	}

	identity, err := manager.Identity(ctx)
	if err != nil {
		return fmt.Errorf("signed in, but fetching the account failed: %w", err)
	}

	if len(identity.Workspaces) == 0 {
		if err := manager.BootstrapWorkspaces(ctx); err == nil {
			if refreshed, err := manager.Identity(ctx); err == nil {
				identity = refreshed
			}
		}
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

	token, err := manager.Token(ctx)
	if err != nil {
		if errors.Is(err, session.ErrLoggedOut) {
			fmt.Println("Not signed in. Run `lucity login`, or set LUCITY_API_TOKEN for automation.")
			return nil
		}
		return err
	}

	claims := decodeJWTClaims(token)
	subject, _ := claims["sub"].(string)
	clientID, _ := claims["client_id"].(string)
	email, _ := claims["email"].(string)
	scope, _ := claims["scope"].(string)

	machine := subject != "" && subject == clientID
	sessionKind := "personal"
	if machine {
		sessionKind = "API token"
	}

	if !machine {
		if identity, idErr := manager.Identity(ctx); idErr == nil {
			fmt.Printf("%s <%s>\n", identity.Name, identity.Email)
			fmt.Printf("Platform: %s\n", manager.APIURL())
			fmt.Printf("Session:  %s\n\n", sessionKind)
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
	}

	fmt.Printf("Platform:  %s\n", manager.APIURL())
	fmt.Printf("Session:   %s\n", sessionKind)
	fmt.Printf("Workspace: %s\n", manager.Workspace())
	if email != "" {
		fmt.Printf("Email:     %s\n", email)
	}
	if scope != "" {
		fmt.Printf("Roles:     %s\n", scope)
	}
	return nil
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
	identity, err := manager.Identity(ctx)
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

func cmdToken(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("token", flag.ExitOnError)
	account := flags.Bool("account", false, "print the account token used for GitHub-backed calls")
	if err := flags.Parse(args); err != nil {
		return err
	}

	manager, err := session.Load()
	if err != nil {
		return err
	}

	if *account {
		token, err := manager.AccountToken(ctx)
		if err != nil {
			return err
		}
		if token == "" {
			return errors.New("no account token for this session — sign in with `lucity login`")
		}
		fmt.Println(token)
		return nil
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
