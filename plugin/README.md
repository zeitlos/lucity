# Lucity plugin for Claude Code

Deploy software to Lucity with one prompt. Point Claude at a local project, say "deploy
this," and the plugin builds it, provisions the databases and buckets it needs, wires the credentials,
ships it, self-heals a failed rollout, and hands you back a live URL.

## What's in the box

- **A `deploy` skill** that turns a local codebase into a running deployment. It discovers your
  application topology (single app, monorepo, frontend + API, web + worker, static site, or a prebuilt
  image), reconciles zero-config build detection with what's actually in your repo, provisions data
  dependencies, and runs a bounded self-heal loop on failures.
- **Bundled Railpack references**, one file per supported technology (Node, Python, Go, PHP, Java,
  Ruby, Rust, Elixir, Deno, Bun, .NET, Gleam, C/C++, shell scripts, static sites) plus the platform
  contracts, so Claude configures builds correctly without guessing.
- **The Lucity MCP server**, wired to your local `lucity` CLI, giving Claude the tools to create
  projects, add services, provision PostgreSQL / Redis / S3-compatible storage, deploy, read logs, and
  add domains.

## Prerequisites

1. **A Lucity workspace.** You need a [Lucity](https://lucity.cloud) account with an active workspace.
2. **The `lucity` CLI**, on your `PATH` as `lucity`. Install it with Go (ensure `$(go env GOPATH)/bin` is on your `PATH`), then authenticate:

   ```sh
   go install github.com/zeitlos/lucity/cli/cmd/lucity@latest
   lucity login
   lucity workspace   # confirm an active workspace
   ```

   The MCP server runs as `lucity mcp`; the plugin starts it for you.
3. **Your code on GitHub.** Lucity builds from GitHub, and the workspace's GitHub app must be able to
   see the repository. The platform only ever reads your repo. It never commits, pushes, or writes hooks.

## Install

Marketplace-style, from the Lucity repository:

```sh
claude plugin marketplace add zeitlos/lucity
claude plugin install lucity
```

Then start a Claude Code session in your project and ask it to deploy.

## License

AGPL-3.0. Part of [Lucity](https://github.com/zeitlos/lucity).
