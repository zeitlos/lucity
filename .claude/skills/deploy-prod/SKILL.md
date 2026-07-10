---
name: deploy-prod
description: >-
  Deploy the Lucity platform to production: upgrade the `lucity` control-plane
  chart and/or the `lucity-infra` chart on the lucity-prod cluster to a version
  published on ghcr.io. Use this whenever the user asks to deploy, release, ship,
  roll out, or upgrade production, "bump prod to <version>", "deploy the platform
  chart", "deploy the infra chart", "release to lucity-prod", or "upgrade the
  cluster", even if they don't say "skill". It lists published chart versions,
  reads what is currently installed, diffs the two with git, flags any breaking
  changes or manual migrations (new values, annotations, new secrets), and then
  applies with `make deploy-prod` / `make deploy-prod-infra`. This is for the
  platform's OWN charts, not for deploying user applications (that is the
  separate `lucity:deploy` skill).
---

# Deploy Lucity to production

Upgrade the Lucity platform on the **lucity-prod** cluster to a chart version already
published on ghcr.io. Two charts live here and this skill drives either or both:

| Chart          | What it is                                             | Make target           | Namespace     | Config (values / secrets)                                              |
| -------------- | ------------------------------------------------------ | --------------------- | ------------- | ---------------------------------------------------------------------- |
| `lucity`       | Control plane (conductor, cashier, dashboard, docs)    | `make deploy-prod`       | lucity-system | `deployments/lucity-prod/values.yaml` + `secrets.yaml`                 |
| `lucity-infra` | Cluster infra (Zot, CNPG, VictoriaMetrics/Logs, Grafana, Traefik, OTel) | `make deploy-prod-infra` | lucity-system | `deployments/lucity-prod/infra-values.yaml` + `infra-secrets.yaml`     |

Both charts share one version stream (the release workflow stamps every chart with the
same git-describe CalVer, so `26.7.1` and `26.7.1-rc.2-6-g1f6c151` are valid versions of
both). They are upgraded independently, and their currently-installed versions usually
differ.

## Safety rails (read first)

- **This is production.** Never apply without showing the user the resolved version and the
  change summary first, and getting an explicit go-ahead. A deploy is a fresh release
  revision that rolls the control plane.
- **Scope is a choice.** Ask up front which chart(s) to deploy: platform (`lucity`), infra
  (`lucity-infra`), or both. Default to whichever the user named; if they said "deploy prod"
  without qualifying, ask.
- **Secrets are the user's job.** If a new version needs a new secret key, the user adds it
  to the gitignored `secrets.yaml` / `infra-secrets.yaml` by hand. Never invent or write
  secret values.
- **Never commit or push** (repo rule). Value-file edits you make for a migration stay in the
  working tree for the user to review.
- The make targets pin `--kube-context lucity-prod`, so a deploy always targets prod
  regardless of your current kube context. Still, verify the context exists and is reachable.

## Output contract

Work quietly. Run the discovery, diff, and preview commands **without narrating each one** and
without printing per-step progress prose ("now checking…", "found X, next I'll…"). The user wants
the conclusions, not the play-by-play. Emit user-facing text at exactly these moments:

1. **One consolidated report** after Steps 1-3: the selected version vs what is installed, the
   change summary (commit list plus any notable value/template/secret diffs), and the migration
   classification split into `[I can apply]` and `[you must do]`. End it with the single
   confirmation ask.
2. **One short result** after Step 4: what deployed, the new revision, and rollout health.

The only thing that interrupts this is a **blocker** you must surface immediately: broken git
ancestry, an unreachable cluster, a detected downgrade, or missing/new secrets. Report those the
moment you hit them; do not bury them in the final summary.

## Step 1 — in parallel: pick a version, and read what is installed

Run 1a and 1b together; they are independent.

### 1a. List published versions (newest first)

Fetch tags so version strings resolve to local commits, then list the chart's tags on
ghcr.io ordered by commit date:

```sh
git fetch --tags --quiet origin

CHART=lucity   # or lucity-infra
crane ls ghcr.io/zeitlos/lucity/charts/$CHART | while read -r v; do
  if [[ "$v" =~ -g[0-9a-f]{7,}$ ]]; then ref="${v##*-g}"; else ref="v$v"; fi
  read -r ts ds <<<"$(git log -1 --format='%ct %cs' "$ref" 2>/dev/null || echo '0 unknown')"
  printf "%s\t%s\t%s\n" "$ts" "$ds" "$v"
done | sort -rn | head -15 | cut -f2,3 | column -t
```

Present the recent versions to the user (with their commit dates) and let them pick.
Point out the newest final release (no `-rc.` and no `-g<sha>` snapshot suffix) as the
usual choice; snapshots and `-rc` builds are fine when deliberately chosen.

### 1b. Read the currently-installed version

```sh
helm list -n lucity-system --kube-context lucity-prod
```

The `CHART` column reads `lucity-<version>` / `lucity-infra-<version>`; strip the chart-name
prefix to get the installed version. That version is the **from** point for the diff.

## Step 2 — derive what changed (git)

Every version string encodes the commit it was built from. Map version to git ref with:

- Snapshot `26.6.2-16-gda4470d` (ends in `-g<sha>`): the ref is the short SHA after the
  last `-g`, i.e. `da4470d`.
- Clean tag or rc `26.7.1`, `26.7.1-rc.2`: the ref is `v` + the version, i.e. `v26.7.1`.

```sh
ref_of() { case "$1" in *-g[0-9a-f]*) echo "${1##*-g}";; *) echo "v$1";; esac; }
FROM=$(ref_of "<installed-version>")
TO=$(ref_of "<selected-version>")

git log --oneline "$FROM..$TO"      # commits being applied
```

If `git log` errors or shows nothing, the ancestry may be broken (rebased history) or you
picked an older target than what is installed. Compare with `git log --oneline "$TO..$FROM"`:
if that has commits, this is a **downgrade / rollback**. Call that out explicitly before
proceeding.

Then diff the migration surfaces (only the paths that can require operator action), scoped
to the chart being deployed:

```sh
# platform (lucity)
git diff "$FROM..$TO" -- charts/lucity/values.yaml charts/lucity/templates \
  charts/lucity-app/values.yaml services/conductor/internal/deployer/values \
  deployments/lucity-prod/values.yaml deployments/lucity-prod/secrets.yaml.example

# infra (lucity-infra)
git diff "$FROM..$TO" -- charts/lucity-infra/Chart.yaml charts/lucity-infra/values.yaml \
  charts/lucity-infra/templates \
  deployments/lucity-prod/infra-values.yaml deployments/lucity-prod/infra-secrets.yaml.example
```

## Step 3 — classify: breaking changes or manual migrations?

Read the diff and decide whether the upgrade is safe to apply as-is or needs a migration
step first. What to look for, and who fixes it:

**Values schema changes** (`values.yaml`):
- A new **required** value with no default: add it to the prod values file. You can propose
  and apply this edit (it is a tracked, non-secret file).
- A **renamed or removed** key that the live release still carries: the platform re-applies
  onto persisted release values, so a bare rename can silently drop the old data. Flag it and
  add the new key; note the old one is now dead. See the values-reshape gotcha in memory.

**Template / immutability changes** (`templates/`, `Chart.yaml` subchart bumps):
- Changes to immutable fields (StatefulSet `volumeClaimTemplates`, selector labels, a PVC),
  or a subchart CRD major bump, can make `helm upgrade` fail. These are not auto-fixable by a
  values edit; warn the user, and if it wedges, the fix is usually a `--cascade=orphan`
  recreate or `deploymentStrategy: Recreate`. Known prod cases: the Valkey VCT immutability
  wedge and the Grafana RWO rolling-update deadlock (infra chart).

**New secret keys** (`secrets.yaml.example` / `infra-secrets.yaml.example`):
- Any added key here is a secret the user must add to their gitignored `secrets.yaml` /
  `infra-secrets.yaml` **by hand** before deploying. List exactly which keys are new. Do not
  fill them in.

**Annotations / runtime migrations on existing resources:**
- If the new version expects annotations or labels on already-running releases (e.g. an
  autodeploy annotation), that is a one-off stamping step, not part of `helm upgrade`.
  Propose a small `kubectl annotate` script over the affected releases; you can run it after
  the user confirms.
- Reconciler-stamped defaults (e.g. a raised ResourceQuota) are not re-applied by the two-minute
  reconciler; if a default changed, note it needs a re-Ensure (a Settings save or `kubectl patch`).

**Decision:**
- **Migrations needed** → tell the user, split into `[I can apply]` (values edits, an annotation
  stamping script) and `[you must do]` (add new secret keys, since those are yours). Propose the
  concrete edits/commands, apply the auto-fixable ones after confirmation, and wait for the user
  to confirm secrets are in place before applying.
- **No migrations** → show the summary (the `git log` commit list plus a rendered manifest diff,
  see below) and go straight to confirm-then-apply.

Preview the exact manifest changes against the live release with the `helm diff` plugin (it
is installed). This is the most reliable "what will actually change" view and catches
immutable-field errors before they happen:

```sh
# platform
helm diff upgrade lucity oci://ghcr.io/zeitlos/lucity/charts/lucity --version <selected-version> \
  --kube-context lucity-prod -n lucity-system \
  -f deployments/lucity-prod/values.yaml -f deployments/lucity-prod/secrets.yaml
```

(For infra, swap the release name, chart, and the two `infra-*` files. The `make ... HELM_ARGS="--dry-run"`
form works too, but `helm diff` shows only what changes.)

## Step 4 — apply, then check rollout health

After the user confirms the version and any migrations are handled:

```sh
make deploy-prod       VERSION=<selected-version>     # platform
make deploy-prod-infra VERSION=<selected-version>     # infra
```

Extra helm flags pass through `HELM_ARGS` if ever needed, e.g. `HELM_ARGS="--timeout 15m"`.

Then check rollout health with a **non-blocking snapshot** (do not use `rollout status` without
`--watch=false`, and do not use `helm --wait`: both block the session, which is the wrong shape
for an agent). For the platform chart, `READY == WANT` and `OBSERVED == GEN` on every deployment
means the new revision is fully rolled out:

```sh
kubectl --context lucity-prod -n lucity-system get deploy \
  lucity-conductor lucity-cashier lucity-dashboard lucity-docs \
  -o custom-columns=NAME:.metadata.name,READY:.status.readyReplicas,WANT:.spec.replicas,GEN:.metadata.generation,OBSERVED:.status.observedGeneration
```

For the infra chart the workloads span many subcharts, so snapshot the pods:

```sh
kubectl --context lucity-prod -n lucity-system get pods | grep -v Running    # anything not Running/Completed is suspect
```

The rollout takes a moment after `helm upgrade` returns, so a snapshot right away may show it
still progressing (`READY` below `WANT`). That is not a failure: report it as "still rolling out"
and re-run the snapshot once or twice rather than blocking. Only if it is still not `READY` after
a couple of checks (or a pod is Pending/CrashLoopBackOff) is it a real stall, handled below.

Confirm the release landed:

```sh
helm list -n lucity-system --kube-context lucity-prod        # REVISION bumped, STATUS deployed, CHART shows the new version
```

If a rollout never completes or a pod stays Pending, do not force it blindly: report the error
and reach for the known remedies from Step 3 (orphan-recreate, delete the stuck old pod,
`Recreate` strategy).

## Reference — the version scheme

The release workflow (`.github/workflows/release.yml`) stamps every image and chart with one
git-describe CalVer:

- On a `v*` tag: the version is the tag without the `v` (`v26.7.1` → `26.7.1`).
- Otherwise: `git describe --tags --match 'v*'` → `<nearest-tag>-<commits-since>-g<short-sha>`
  (e.g. `26.7.1-rc.2-6-g1f6c151` is 6 commits past `v26.7.1-rc.2`, at commit `1f6c151`).

That trailing `-g<sha>` (and the `v<tag>` for clean versions) is what lets you map any
published version back to a commit and diff two of them with plain git.
