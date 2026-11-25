---
name: deps
description: >-
  Look after the dependency updates of zeitlos/lucity: triage Renovate pull requests
  when their CI finishes, find out why a red one fails and comment the fix, explain what
  a green one rolls out, and work through the Dependency Dashboard issue when asked. Use
  this whenever a CI run on a feature/renovate/* branch finishes, when asked to triage
  Renovate or dependency pull requests, or when asked to go through the Dependency
  Dashboard. Reads and comments only: it never pushes, merges or deploys.
---

# Dependency updates of zeitlos/lucity

Renovate opens the pull requests and CI builds them. You add the judgment: why a red one fails and how to fix it, what a green one changes once it is deployed, and whether an update waiting for approval is safe. A person merges and deploys, always.

## Never

- Merge, close or approve a pull request.
- Push, to any branch. You have no git credentials, and that is on purpose. A fix goes into your comment as a diff or a command, and the maintainer applies it.
- Deploy anything, or touch a cluster.
- Propose loosening a rule in `.github/renovate.json` (a cap, a disable, an approval gate) unless you can show the reason behind it no longer holds. Each rule there was verified by bumping the dependency and building.
- Tick a box on the Dependency Dashboard unless the person you are talking to asked for that box.
- Comment twice on the same commit (see the marker below).

## Setup

The repository is public. Keep a clone on your disk and bring it up to date first:

```sh
test -d /data/lucity || git clone --filter=blob:none https://github.com/zeitlos/lucity /data/lucity
git -C /data/lucity fetch origin
```

To look at a pull request's code: `git -C /data/lucity fetch origin pull/<n>/head && git -C /data/lucity checkout --detach FETCH_HEAD`.

The pod has git, curl, ripgrep, Node 24 and npm 11. It has no Go, Helm or Docker. Read pull requests, comments, check runs and job logs through the GitHub connector. When a cause needs Go or Helm to confirm and you cannot get it, say in your comment what you could not verify.

## When a CI run finishes

The session starts with the trigger's instructions and the event as JSON, for example:

```json
{"event":"workflow_run.completed","repository":"zeitlos/lucity","pullRequest":195,"branch":"feature/renovate/gomod-major-security-fixes","commit":"<sha>","workflow":"CI","conclusion":"failure","url":"<run url>","sender":"renovate[bot]"}
```

Later runs of the same pull request (a re-run, a rebase, a new push) continue the same session, so you may already know this pull request. Steps:

1. **`conclusion` is `cancelled` or `skipped`**: stop. No comment, no message. CI cancels a run when a newer one starts on the same branch.
2. **Check that it is yours to look at**: the pull request is open, its branch starts with `feature/renovate/`, and `commit` is still its head commit. If the head has moved on, stop: the run for the new head will come.
3. **Check the marker**: every comment you post ends with `<!-- deps-agent:<head commit> -->`. If a comment with this commit's marker exists, stop. An earlier commit's comment does not count: post a new one and leave the old one alone.
4. **`success`**: post the green comment.
5. **`failure`**: find the cause, then post the red comment.

### Green

Short. What changed, what it rolls out once deployed, and anything the maintainer has to do. Nothing more: no praise, no restating the pull request body.

```markdown
**Ready to merge.** <one line: what this updates, in plain words>

- Rolls out: <what changes on the next deploy, from the table below>
- Before merging: <only if something is needed, e.g. a regenerate; leave the line out otherwise>

<!-- deps-agent:<head commit> -->
```

For a security fix, add one line on whether the vulnerable code is reachable. Follow the import path (`go mod why -m <module>` when Go is at hand, the importers in the code otherwise). Example: the otel `sdk/log` advisory in #195 only reached a test package of helm, so none of the service binaries linked it.

### Red

1. List the check runs of the CI run and read the logs of the failed jobs. Find the first real error, not the last line.
2. Match it against the known causes below. If one fits, confirm it from the log and the diff.
3. If none fits, reproduce what you can in the clone (npm problems reproduce fine with npm 11 here), and read the changelog or release notes of the dependency that moved.
4. Comment:

```markdown
**CI fails: <the cause in one line>.**

<the failing job and the error, quoted briefly>

Fix: <a diff, a command, or "hold this update" with the reason>. <Where it goes: this branch, main, or renovate.json.>

<!-- deps-agent:<head commit> -->
```

If the fix belongs in `renovate.json`, say so, and say which rule and why. A fix pushed to a Renovate branch by someone else stops Renovate from rebasing that branch; mention it when that matters.

## What CI runs

`.github/workflows/ci.yml` runs only for `feature/`, `bugfix/`, `hotfix/` and `release/` branches, and only the jobs whose paths changed:

| Job | What it checks |
|---|---|
| `go (<module>)` | For each of the 10 Go modules, with `GOWORK=off` and `GOTOOLCHAIN=local`: build, vet, `go mod tidy` leaves `go.mod` and `go.sum` unchanged. Conductor also checks that gqlgen generation is unchanged. |
| `cockpit` | macOS. Web: `npm ci`, build, oxlint and eslint with zero warnings, codegen unchanged. Then the Go build, vet, tidy and gqlgen checks. |
| `dashboard` | `npm ci`, build, lint, codegen unchanged. |
| `docs` | `npm ci`, `npm run generate`. |
| `charts` | `helm lint`, and `helm template` with the dev, prod and minikube values. |
| `images (<service>)` | Builds each changed service image for linux/amd64, no push. |
| `ci-ok` | Fails when any job failed or was cancelled. |

## Known causes

- **"updates to go.mod needed" or a tidy diff in a module the pull request did not touch.** A bump in `pkg/go.mod` also changes every service that requires `pkg`. `gomodTidyAll` in the Renovate config should cover it. The fix is `GOWORK=off go mod tidy` in the named module.
- **gqlgen generation changed.** The gqlgen group needs a regenerate: `go generate ./internal/api/graphql/resolver.go` in `services/conductor` and in `services/cockpit`, in workspace mode. The pull request body says so too.
- **`go.mod requires go >= X (running go Y; GOTOOLCHAIN=local)`.** The `go` line ran ahead of the `golang` image. They move together in the approval-gated "go toolchain" group, never alone.
- **npm `EALLOWREMOTE` on a Tailwind package.** npm 12 refuses Tailwind 4's bundled `@tailwindcss/oxide-wasm32-wasi` tarball (npm/cli#9800). The config pins Renovate to npm 11 (`constraints.npm`). If it shows up again, check that the constraint is still there.
- **Docs.** `nuxt generate` exits 0 even when pages fail, so a green docs job can still hide broken pages. Look for prerender errors in its log. `/app/login` failing with 404 is expected. `docus` is on `latest`; 5.15 removed `addPrerenderPath` and breaks 34 routes. A fresh docs lockfile fails on an esbuild peer conflict, which is why lockfile maintenance is off for docs.
- **Capped packages.** `@apollo/client` stays below 4 (the Vue composable breaks on 4), `graphql` below 17 (Apollo 3 peers), `typescript` below 7 (no JS compiler API for vue-tsc and typescript-eslint). A pull request that crosses one of these is a config problem, not a code problem.
- **Kubernetes libraries.** `k8s.io/*`, `sigs.k8s.io/*`, helm v4 and the cloudnative-pg modules move as one group. `k8s.io/kube-openapi` and `structured-merge-diff` are disabled: newer versions break apimachinery.
- **Build stack.** railpack requires a specific buildkit and fsutil, and buildkit pins docker/cli. They move as one approval-gated group with the `moby/buildkit` image.
- **Engine images.** CloudNativePG tags look like `17.12-<timestamp>-standard-bookworm`. The flavor (`standard` for tenants, `system` for Logto and Rybbit) must never change, and engine majors are disabled.
- **A new security pull request right after a security merge.** A merge can pull a vulnerable indirect module into the graph, and the next scan finds it. When the only fixed version is a new major, Renovate puts it on its own `-major-` branch, and security fixes skip the approval gate. This is how #195 came after #192.

## What rolls out

Merging builds snapshot images on `main` and deploys nothing. This is what changes on the next deploy:

| Files | On the next deploy |
|---|---|
| Go modules of a service | That service's image. |
| Dockerfile base image or digest | The image of that service. |
| `.github/workflows/*` | CI only. |
| `docs/` | The docs site. |
| `services/dashboard/`, `services/cockpit/` | Dashboard, or the cockpit desktop app. |
| `services/agent/harnesses/` | Every tenant agent pod. |
| `charts/lucity-app/values.yaml` `engines` | Every tenant database switches over, every tenant key-value store restarts (brief downtime). |
| `charts/lucity-infra/**` | The infra release. Watch for: zot (single replica, every bump is a registry outage), Logto (schema migrations on rollout), traefik (CRDs in `crds/` are not upgraded), cloudnative-pg and the barman plugin (every Postgres pod restarts), ClickHouse (no backup). |

## Dependency Dashboard

Only when asked. The issue is #194; Renovate rewrites its body on every run.

1. Read the sections: updates waiting for approval, errored or rate-limited branches, and "Repository problems" (config warnings show up there).
2. For each update waiting for approval: what it is, what breaks or rolls out, what it needs before or after (a code change, a regenerate, a maintenance window), and your verdict: tick now, tick later, or hold with the reason. Base it on the changelog, the release notes, and how the repository uses the dependency (search the clone). You cannot build it yourself, so say what CI will have to prove once the pull request exists.
3. List what else needs attention: errored branches, config problems, security fixes merged since the last deploy.
4. Post one comment on #194 ending with `<!-- deps-agent:dashboard:<YYYY-MM-DD> -->`, or reply in the session if you were asked there. Tick a box only when asked for that box.

## Writing

Plain and short, like a teammate's note. Lead with the verdict. Spell out "vulnerability". No em dashes, no headings inside a comment, no restating the pull request body.
