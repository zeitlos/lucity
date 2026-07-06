# Lucity plugin for Claude Code

Deploy software to your Lucity platform with one prompt. Point Claude at a local project, say "deploy
this," and the plugin builds it, provisions the databases and buckets it needs, wires the credentials,
ships it, self-heals a failed rollout, and hands you back a live URL.

Everything it deploys stays yours. Lucity runs your workloads as standard Kubernetes and Helm, so there
is no lock-in: you can eject to plain infrastructure-as-code at any time and take your cluster with you.

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

## How it builds

Builds are zero-config, powered by [Railpack](https://railpack.com). Railpack analyzes your source,
picks the right toolchain, and produces an image (no Dockerfile required). When a default is wrong, the
skill overrides it through Railpack's standard knobs (version pins, custom build/start commands, extra
packages).

Lucity stands on a lot of great open source: [Kubernetes](https://kubernetes.io) and
[Helm](https://helm.sh) for orchestration and releases, [Railpack](https://railpack.com) for builds,
[CloudNativePG](https://cloudnative-pg.io) for managed PostgreSQL, and [Zot](https://zotregistry.dev)
for the container registry.

## Prerequisites

1. **A Lucity workspace.** You need an account on a Lucity instance and an active workspace.
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

Or from a local checkout (handy while developing the plugin):

```sh
claude plugin install ./plugin
```

Then start a Claude Code session in your project and ask it to deploy.

## Quickstart

A first deploy usually reads like this:

```
You:    deploy this to lucity

Claude: Checking your Lucity account and workspace... authenticated, workspace "acme" is active.
        Found this repo on GitHub as acme/store-api, visible to the workspace app.
        Detected a single Node.js service (Node 22, start: "node dist/server.js").
        It imports "pg" and reads DATABASE_URL, and there's a migrations/ directory.

        Plan:
          - Create project "store-api" (development environment)
          - Add one service, pinned to Node 22
          - Provision a PostgreSQL database, wire DATABASE_URL by reference
          - Run the migrations, then deploy

        Provisioning PostgreSQL is a paid resource. Go ahead?

You:    yes

Claude: Created the project and service. Database ready, DATABASE_URL wired.
        Ran 4 migrations. Deploying...
        Build succeeded. First rollout hit OOM_KILLED at 512Mi, bumped to 1Gi, rolled out clean.
        Added a domain and curled it: 200 OK.

        Live: https://store-api-development.<your-lucity-domain>
        Services: store-api (Node 22, "node dist/server.js", 1Gi memory)
        Resources: 1 PostgreSQL database (DATABASE_URL wired by reference)
        Decisions: pinned Node 22 from engines.node; raised memory to 1Gi after an OOM.
        No secrets were found committed in the repo.
```

Claude decides anything your repo already answers and tells you why. It only asks the business questions:
which repos form your product, whether it should be public, whether it may import a data dump, and
whether it may create paid resources.

## License

AGPL-3.0. Part of [Lucity](https://github.com/zeitlos/lucity).
