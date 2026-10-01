---
name: deploy-platform
description: >-
  Deploy the Lucity platform's own Helm charts, the `lucity` control plane and the
  `lucity-infra` cluster infrastructure, to lucity-dev by default, or to lucity-prod
  only when the user explicitly asks for production. Use this whenever the user asks to
  deploy, release, ship, roll out, or upgrade the platform: "deploy to dev", "bump dev to
  <version>", "deploy the infra chart", "release to prod", "bump prod to <version>",
  "upgrade the cluster", even if they don't say "skill". It lists published chart
  versions, reads what each cluster runs, diffs the versions with git, flags anything that
  needs action first (new values, annotations, new secrets), and applies with the matching
  make target. A prod deploy first proves the same version is already live and healthy on
  dev. This is for the platform's OWN charts, not for deploying user applications (that is
  the separate `lucity:deploy` skill).
---

# Deploy the Lucity platform

Upgrade the Lucity platform on **lucity-dev** (the default) or **lucity-prod** (only when
explicitly asked) to a chart version already published on ghcr.io.

| Target | When | Kube context | Make targets | Values and secrets |
| --- | --- | --- | --- | --- |
| dev | default | `lucity-dev` | `deploy-dev`, `deploy-dev-infra` | `deployments/lucity-dev/` |
| prod | only on an explicit ask | `lucity-prod` | `deploy-prod`, `deploy-prod-infra` | `deployments/lucity-prod/` |

Two charts live here and this skill drives either or both:

| Chart | What it is | Make target | Values / secrets |
| --- | --- | --- | --- |
| `lucity` | Control plane (conductor, dashboard, docs, plus cashier on prod) | `deploy-<env>` | `values.yaml` + `secrets.yaml` |
| `lucity-infra` | Cluster infra (Zot, CNPG, Logto, VictoriaMetrics/Logs, alerting, Grafana, OTel) | `deploy-<env>-infra` | `infra-values.yaml` + `infra-secrets.yaml` |

Both charts share one version stream (the release workflow stamps every chart with the same
git-describe CalVer, so `26.7.1` and `26.7.1-rc.2-6-g1f6c151` are valid versions of both).
They are upgraded independently, and their installed versions usually differ. The examples
for the gitignored secrets files live once, at `deployments/secrets.yaml.example` and
`deployments/infra-secrets.yaml.example`.

## Safety rails (read first)

- **Dev is the default target.** "Deploy", "ship it", "release", "roll out", or a version
  bump without a named target means dev. Deploy to prod only when the user explicitly says
  prod, production, or lucity-prod. Never infer prod from context, urgency, or a previous
  prod deploy in the same conversation.
- **Prod requires dev first.** Before anything else on a prod deploy, Step 2 proves that the
  version is already live on dev and that dev is healthy. If that gate fails, stop and say
  so. The usual fix is to run this skill against dev first.
- **Never apply without showing the resolved version and the change summary first.** A deploy
  is a fresh release revision that rolls the workloads.
  - **Clean bumps are auto-approved.** When Step 4 finds no open tasks and no blocker (and, on
    prod, the dev gate passed), present the summary and apply in the same turn with no
    confirmation ask. This is the common case: an app-code version bump with no
    values/template/secret/immutability changes and no downgrade.
  - **Everything else stops for an explicit go-ahead.** If anything has to happen before the
    upgrade is safe, or there is any blocker (a failed dev gate, a new secret key, a downgrade,
    broken git ancestry, an unreachable cluster, or ambiguous scope), stop after the summary
    and wait for the user before applying.
- **Scope is a choice.** Deploy whichever chart the user named. If they only said "deploy",
  ask: platform (`lucity`), infra (`lucity-infra`), or both. For both, apply `lucity-infra`
  first, then `lucity`, because the platform builds on what infra provides.
- **Dev is shared.** Other people's environments run on lucity-dev, so a dev deploy still gets
  the same diff and health checks. Dev is the default target, not a scratch cluster.
- **Secrets are the user's job.** If a new version needs a new secret key, the user adds it to
  the gitignored `secrets.yaml` / `infra-secrets.yaml` by hand. Never invent or write secret
  values.
- **Never commit or push** (repo rule). Value-file edits you make stay in the working tree for
  the user to review.
- The make targets pin `--kube-context`, so a deploy always reaches the intended cluster
  regardless of your current context. Still, verify the context exists and is reachable.
- Always pass `VERSION=` explicitly. Without it, Helm resolves the highest stable tag, which
  ranks below every prerelease and snapshot build, and silently downgrades.

## Output contract

Work quietly. Run the discovery, gate, diff, and preview commands **without narrating each
one** and without per-step progress prose ("now checking…", "found X, next I'll…"). The user
wants the conclusions. Emit user-facing text at exactly these moments:

1. **One consolidated report** after Steps 1-4: the target cluster, the selected version vs
   what is installed, on prod the dev gate result (live on dev since when, health), the change
   summary (commit list plus any notable value/template/secret diffs), and anything that has
   to happen before the upgrade is safe, under two plain headings: **Before I deploy** (what
   you will do, such as values edits or an annotation stamping script) and **Needs you** (what
   only the user can do, such as adding new secret keys). Ordinary prose bullets. Omit a heading
   entirely when it has nothing under it.
   - **Nothing to do and no blocker** → end the report by stating you are proceeding (clean
     bump, auto-approved) and continue straight to Step 5 in the **same turn**.
   - **Anything outstanding, or a blocker** → end the report with the single confirmation ask
     and wait.
2. **One short result** after Step 5: what deployed where, the new revision, and rollout health.

A **blocker** interrupts this and gets reported the moment you hit it: a failed dev gate,
broken git ancestry, an unreachable cluster, a detected downgrade, or missing/new secrets.

## Step 1 — in parallel: pick the target, list versions, read what is installed

Settle the target first (dev unless prod was explicitly asked), then run 1a and 1b together.

```sh
ENV=dev            # or prod, only on an explicit ask
CTX=lucity-$ENV
```

### 1a. List published versions (newest first)

Fetch tags so version strings resolve to local commits, then list the chart's tags on ghcr.io
ordered by commit date:

```sh
git fetch --tags --quiet origin

CHART=lucity   # or lucity-infra
crane ls ghcr.io/zeitlos/lucity/charts/$CHART | while read -r v; do
  if [[ "$v" =~ -g[0-9a-f]{7,}$ ]]; then ref="${v##*-g}"; else ref="v$v"; fi
  read -r ts ds <<<"$(git log -1 --format='%ct %cs' "$ref" 2>/dev/null || echo '0 unknown')"
  printf "%s\t%s\t%s\n" "$ts" "$ds" "$v"
done | sort -rn | head -15 | cut -f2,3 | column -t
```

Which version to propose depends on the target:

- **dev**: the newest published build. Snapshots (`-g<sha>`) are normal on dev, because that
  is where builds get validated.
- **prod**: the version currently live on dev for that chart (from 1b), because that is the
  one that has been validated. If the user names a different version, the dev gate in Step 2
  still applies to it.

### 1b. Read the installed versions

```sh
helm list -n lucity-system --kube-context "$CTX"
helm list -n lucity-system --kube-context lucity-dev     # prod only: dev's versions feed the gate
```

The `CHART` column reads `lucity-<version>` / `lucity-infra-<version>`. Strip the chart-name
prefix to get the version. The target cluster's installed version is the **from** point for
the diff.

## Step 2 — prod only: the dev gate

Skip this step on a dev deploy. On prod, every check below must pass for each chart being
deployed. A failure is a blocker: report it and stop.

### 2a. The version is already live on dev

Map both versions to commits (see the version scheme at the end) and require that dev runs the
target or a newer build that contains it:

```sh
ref_of() { case "${1}" in *-g[0-9a-f]*) echo "${1##*-g}";; *) echo "v${1}";; esac; }
TO=$(ref_of "<selected-version>")
DEV=$(ref_of "<dev-installed-version>")

git merge-base --is-ancestor "$TO" "$DEV" && echo "live on dev"
```

- Dev runs exactly the target: the best case.
- Dev runs a newer build that contains the target: acceptable, but say that dev validated it as
  part of a larger set of changes.
- Otherwise the change is not on dev. Stop and offer to deploy the version to dev first.

Report how long it has been live there:

```sh
helm history "$CHART" --kube-context lucity-dev -n lucity-system --max 20 -o json \
  | jq -r '.[] | "\(.revision)\t\(.updated)\t\(.status)\t\(.chart)"'
```

### 2b. Dev is healthy

```sh
helm list -n lucity-system --kube-context lucity-dev       # STATUS must be "deployed", not failed or pending-*

kubectl --context lucity-dev -n lucity-system get deploy,sts -o json \
  | jq -r --arg r "$CHART" '.items[] | select(.metadata.annotations["meta.helm.sh/release-name"] == $r)
      | [.kind, .metadata.name, (.status.readyReplicas // 0), .spec.replicas, .metadata.generation, (.status.observedGeneration // 0)] | @tsv' \
  | column -t                                              # READY == WANT and OBSERVED == GEN on every row

kubectl --context lucity-dev -n lucity-system get pods -o json \
  | jq -r '.items[] | select(.status.phase != "Succeeded") | . as $pod | [.status.containerStatuses[]? | select(.ready | not)]
      | select(length > 0) | "\($pod.metadata.name)\t\($pod.status.phase)\t\(map(.state.waiting.reason // .state.terminated.reason // "starting") | join(","))"'

kubectl --context lucity-dev get --raw \
  '/api/v1/namespaces/lucity-system/services/lucity-infra-victoria-metrics-alert-alertmanager:9093/proxy/api/v2/alerts?active=true&silenced=false&inhibited=false' \
  | jq -r '.[] | "\(.labels.severity)\t\(.labels.alertname)\t\(.annotations.summary)"'
```

Gate rules:

- The release is `deployed`, and every workload it owns is fully rolled out.
- The pod check prints nothing: every container in `lucity-system` is ready. Also glance at
  `kubectl get pods` for recent restarts (the `RESTARTS` column shows "(5m ago)"); on a
  platform pod those count against the gate too.
- No critical alert is firing on dev. A firing warning is not automatically a blocker, but name
  it in the report and say whether it could relate to this change. If the Alertmanager service
  does not exist, say that alerting is off on dev and rely on the other checks.

## Step 3 — derive what changed (git)

Every version string encodes the commit it was built from:

- Snapshot `26.6.2-16-gda4470d` (ends in `-g<sha>`): the ref is the short SHA after the last
  `-g`, i.e. `da4470d`.
- Clean tag or rc `26.7.1`, `26.7.1-rc.2`: the ref is `v` + the version, i.e. `v26.7.1`.

```sh
FROM=$(ref_of "<installed-version-on-target>")
TO=$(ref_of "<selected-version>")

git log --oneline "$FROM..$TO"      # commits being applied
```

If `git log` errors or shows nothing, the ancestry may be broken (rebased history, or a side
branch built on dev) or you picked an older target than what is installed. Compare with
`git log --oneline "$TO..$FROM"`: if that has commits, this is a **downgrade / rollback**. Call
that out explicitly before proceeding.

Then diff the surfaces that can require operator action, scoped to the chart being deployed.
The secrets examples moved from `deployments/lucity-<env>/` to `deployments/` on 2026-10-01;
keeping the old path in the pathspec lets git show a rename when `FROM` predates the move.

```sh
# platform (lucity)
git diff "$FROM..$TO" -- charts/lucity/values.yaml charts/lucity/templates \
  charts/lucity-app/values.yaml services/conductor/internal/deployer/values \
  deployments/lucity-$ENV/values.yaml \
  deployments/secrets.yaml.example deployments/lucity-$ENV/secrets.yaml.example

# infra (lucity-infra)
git diff "$FROM..$TO" -- charts/lucity-infra/Chart.yaml charts/lucity-infra/values.yaml \
  charts/lucity-infra/templates deployments/lucity-$ENV/infra-values.yaml \
  deployments/infra-secrets.yaml.example deployments/lucity-$ENV/infra-secrets.yaml.example
```

## Step 4 — classify: is anything needed before this is safe to apply?

Read the diff and decide whether the upgrade is safe to apply as-is, or whether something has
to happen first. What to look for, and who does it:

**Values schema changes** (`values.yaml`):
- A new **required** value with no default: add it to the target's values file. You can propose
  and apply this edit (it is a tracked, non-secret file).
- A **renamed or removed** key that the live release still carries: the platform re-applies
  onto persisted release values, so a bare rename can silently drop the old data. Flag it and
  add the new key; note the old one is now dead. See the values-reshape gotcha in memory.

**Template / immutability changes** (`templates/`, `Chart.yaml` subchart bumps):
- Changes to immutable fields (StatefulSet `volumeClaimTemplates`, selector labels, a PVC),
  or a subchart CRD major bump, can make `helm upgrade` fail. These are not auto-fixable by a
  values edit; warn the user, and if it wedges, the fix is usually a `--cascade=orphan`
  recreate or `deploymentStrategy: Recreate`. Known cases: the Valkey VCT immutability wedge and
  the Grafana RWO rolling-update deadlock (infra chart).

**New secret keys** (`deployments/secrets.yaml.example` / `deployments/infra-secrets.yaml.example`):
- Any added key is a secret the user must add to the target's gitignored `secrets.yaml` /
  `infra-secrets.yaml` **by hand** before deploying. List exactly which keys are new, and check
  whether the target's file already has them. Do not fill them in.

**Annotations / one-off stamping on existing resources:**
- If the new version expects annotations or labels on already-running releases (e.g. an
  autodeploy annotation), that is a one-off stamping step, not part of `helm upgrade`. Propose
  a small `kubectl annotate` script over the affected releases; you can run it after the user
  confirms.
- Reconciler-stamped defaults (e.g. a raised ResourceQuota) **are** re-applied automatically:
  the two-minute `ReconcileServices` loop calls `environment.Ensure` per environment, which
  re-runs the namespace scaffolding (quota, LimitRange, NetworkPolicies, pull secret). A changed
  default reaches every live environment within ~2 min of the conductor rollout, with no
  Settings save and no `kubectl patch`. Verified 2026-08-19: a quota bump landed on all 34 envs
  60s after `helm upgrade` returned. The corollary is that `kubectl patch` on these resources is
  useless as a fix, since `ensureQuota` overwrites it on the next tick.

**Decision:**
- **Something is outstanding** → tell the user under the two plain headings, **Before I deploy**
  and **Needs you**. Propose the concrete edits/commands, apply your own after confirmation, and
  wait for the user to confirm secrets are in place before applying.
- **Nothing outstanding (and no blocker)** → show the summary (the `git log` commit list plus a
  rendered manifest diff, see below) and **apply immediately in the same turn; this case is
  auto-approved, no confirmation ask.** A clean bump means: the diff over the surfaces above is
  empty (no `values.yaml` / template / `lucity-app` / secret-example changes), no new required
  env var, no immutable-field or subchart-CRD change, the version is an upgrade (not a
  downgrade), and on prod the dev gate passed. Any blocker cancels the auto-approval and you
  stop for the user, even if nothing else migrates.

Preview the exact manifest changes against the live release with the `helm diff` plugin (it is
installed). This is the most reliable "what will actually change" view and catches
immutable-field errors before they happen:

```sh
# platform
helm diff upgrade lucity oci://ghcr.io/zeitlos/lucity/charts/lucity --version <selected-version> \
  --kube-context "$CTX" -n lucity-system --suppress-secrets \
  -f deployments/lucity-$ENV/values.yaml -f deployments/lucity-$ENV/secrets.yaml

# infra
helm diff upgrade lucity-infra oci://ghcr.io/zeitlos/lucity/charts/lucity-infra --version <selected-version> \
  --kube-context "$CTX" -n lucity-system --suppress-secrets \
  -f deployments/lucity-$ENV/infra-values.yaml -f deployments/lucity-$ENV/infra-secrets.yaml
```

## Step 5 — apply, then check rollout health

Once approved, either **automatically** (Step 4 found nothing outstanding and no blocker) or by
the user's explicit go-ahead, apply. For both charts, infra goes first:

```sh
make deploy-$ENV-infra VERSION=<selected-version>     # infra
make deploy-$ENV       VERSION=<selected-version>     # platform
```

Extra helm flags pass through `HELM_ARGS` if ever needed, e.g. `HELM_ARGS="--timeout 15m"`.

Then check rollout health with a **non-blocking snapshot** (do not use `rollout status` without
`--watch=false`, and do not use `helm --wait`: both block the session, which is the wrong shape
for an agent). Every workload the release owns should show `READY == WANT` and
`OBSERVED == GEN`:

```sh
kubectl --context "$CTX" -n lucity-system get deploy,sts -o json \
  | jq -r --arg r "<lucity or lucity-infra>" '.items[] | select(.metadata.annotations["meta.helm.sh/release-name"] == $r)
      | [.kind, .metadata.name, (.status.readyReplicas // 0), .spec.replicas, .metadata.generation, (.status.observedGeneration // 0)] | @tsv' \
  | column -t

kubectl --context "$CTX" -n lucity-system get pods -o json \
  | jq -r '.items[] | select(.status.phase != "Succeeded") | . as $pod | [.status.containerStatuses[]? | select(.ready | not)]
      | select(length > 0) | "\($pod.metadata.name)\t\($pod.status.phase)\t\(map(.state.waiting.reason // .state.terminated.reason // "starting") | join(","))"'
```

The rollout takes a moment after `helm upgrade` returns, so a snapshot right away may show it
still progressing (`READY` below `WANT`). That is not a failure: report it as "still rolling
out" and re-run the snapshot once or twice rather than blocking. Only if it is still not ready
after a couple of checks (or a pod is Pending/CrashLoopBackOff) is it a real stall.

Confirm the release landed:

```sh
helm list -n lucity-system --kube-context "$CTX"     # REVISION bumped, STATUS deployed, CHART shows the new version
```

If a rollout never completes or a pod stays Pending, do not force it blindly: report the error
and reach for the known remedies from Step 4 (orphan-recreate, delete the stuck old pod,
`Recreate` strategy).

## Reference — the version scheme

The release workflow (`.github/workflows/release.yml`) stamps every image and chart with one
git-describe CalVer:

- On a `v*` tag: the version is the tag without the `v` (`v26.7.1` → `26.7.1`).
- Otherwise: `git describe --tags --match 'v*'` → `<nearest-tag>-<commits-since>-g<short-sha>`
  (e.g. `26.7.1-rc.2-6-g1f6c151` is 6 commits past `v26.7.1-rc.2`, at commit `1f6c151`).

That trailing `-g<sha>` (and the `v<tag>` for clean versions) is what lets you map any
published version back to a commit and diff two of them with plain git. Snapshot builds from a
branch other than `main` map to commits that may not be ancestors of `main`, which is normal on
dev and is why Step 2 checks ancestry against dev's commit rather than against `main`.

The skill loader substitutes positional placeholders (a dollar sign followed by a digit) with
invocation arguments, so the shell and jq snippets here avoid them on purpose. Keep it that way
when editing: write `${1}` in shell, and use named variables elsewhere.
