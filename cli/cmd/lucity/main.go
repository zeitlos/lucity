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
  lucity <command> [arguments]

Commands:
  login [--api <url>]      Sign in through your browser
  logout                   Discard the stored session
  account                  Show who you are signed in as, and your workspaces
  workspace [<workspace>]  Show or switch the active workspace
  deploy <service>         Build and roll out a service
  status <service>         Show the latest rollout status of a service
  vars <command>           Manage service variables (list, available, set)
  db <command>             Manage databases (create, list, credentials, expose, unexpose, delete)
  token [--account]        Print a bearer token for scripting
  mcp                      Serve the Lucity MCP server on stdio
  version                  Print the CLI version

Run 'lucity help <command>' for the details of a command.

Environment:
  LUCITY_API_URL     Platform URL, in place of the one you signed in to
  LUCITY_WORKSPACE   Workspace to use, in place of the active one
  LUCITY_API_TOKEN   Workspace API token, used in place of the stored session
  LUCITY_CONFIG_DIR  Directory holding the session (default: '$XDG_CONFIG_HOME/lucity'
                     or '~/.config/lucity')
`

const loginUsage = `lucity login — sign in through your browser

Usage:
  lucity login [--api <url>]

Flags:
  --api <url>   Platform to sign in to. Defaults to LUCITY_API_URL, then the
                platform you last signed in to, then ` + api.DefaultBaseURL + `.

Opens your browser to sign in with GitHub and stores the session in the config
directory. Only a refresh token is written to disk, and access tokens are
fetched as commands need them.

The browser hands the session back to a listener on 127.0.0.1, on port 8765,
8766 or 8767, so sign in on the machine your browser runs on. To sign in on a
remote machine, forward the port with 'ssh -L 8765:127.0.0.1:8765 <host>', or
use an API token instead.

Examples:
  lucity login
  lucity login --api https://paas.example.com
`

const logoutUsage = `lucity logout — discard the stored session

Usage:
  lucity logout

Deletes the refresh token from the config directory. The platform URL and the
active workspace stay, ready for the next 'lucity login'.
`

const accountUsage = `lucity account — show who you are signed in as

Usage:
  lucity account

Prints your name and email, the platform, and each workspace you belong to with
your role in it, marking the active one with '*'. With LUCITY_API_TOKEN set, it
describes the token instead.
`

const workspaceUsage = `lucity workspace — show or switch the active workspace

Usage:
  lucity workspace
  lucity workspace <workspace>

Arguments:
  <workspace>   Workspace to switch to. 'lucity account' lists the ones you belong to.

Ids that leave out the workspace, such as 'shop/production/web', resolve against
the active workspace. LUCITY_WORKSPACE overrides it without changing the stored
one.
`

const tokenUsage = `lucity token — print a bearer token for scripting

Usage:
  lucity token [--account]

Flags:
  --account   Print the account token instead. API calls that read from GitHub on
              your behalf send it in the 'X-Lucity-Account-Token' header.

Prints a short-lived access token for the active workspace, to call the API
with directly.

Examples:
  curl https://lucity.cloud/graphql \
    -H "Authorization: Bearer $(lucity token)" \
    -H "Content-Type: application/json" \
    -d '{"query": "{ projects { id } }"}'
`

const mcpUsage = `lucity mcp — serve the Lucity MCP server on stdio

Usage:
  lucity mcp

Speaks the Model Context Protocol on stdin and stdout, so an AI agent can create
projects, deploy, provision databases and read logs with your session. MCP
clients start the server themselves, so configure 'lucity mcp' as its command
rather than running it by hand.
`

const versionUsage = `lucity version — print the CLI version

Usage:
  lucity version
  lucity --version
`

var commandUsage = map[string]string{
	"login":     loginUsage,
	"logout":    logoutUsage,
	"account":   accountUsage,
	"workspace": workspaceUsage,
	"deploy":    deployUsage,
	"status":    statusUsage,
	"vars":      varsUsage,
	"db":        dbUsage,
	"token":     tokenUsage,
	"mcp":       mcpUsage,
	"version":   versionUsage,
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	command, args := os.Args[1], os.Args[2:]
	if text, ok := commandUsage[command]; ok && len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Print(text)
		return
	}

	var err error
	switch command {
	case "login":
		err = cmdLogin(ctx, args)
	case "logout":
		err = cmdLogout()
	case "account":
		err = cmdAccount(ctx)
	case "workspace":
		err = cmdWorkspace(ctx, args)
	case "deploy":
		err = cmdDeploy(ctx, args)
	case "db":
		err = cmdDB(ctx, args)
	case "vars":
		err = cmdVars(ctx, args)
	case "status":
		err = cmdStatus(ctx, args)
	case "token":
		err = cmdToken(ctx, args)
	case "mcp":
		err = cmdMCP(ctx)
	case "version", "--version", "-v":
		fmt.Println("lucity " + version)
	case "help", "--help", "-h":
		err = cmdHelp(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", command, usage)
		os.Exit(2)
	}
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdHelp(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}
	text, ok := commandUsage[args[0]]
	if !ok {
		return fmt.Errorf("unknown command %q — run 'lucity help' for the list", args[0])
	}
	fmt.Print(text)
	return nil
}

func cmdLogin(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("login", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, loginUsage) }
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
	flags := flag.NewFlagSet("token", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.Usage = func() { fmt.Fprint(os.Stderr, tokenUsage) }
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
