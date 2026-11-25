---
name: deps
description: >-
  Look after the dependency updates of zeitlos/lucity: triage Renovate pull requests
  when their CI finishes, find out why a red one fails and comment the fix, explain what
  a green one rolls out, and work through the Dependency Dashboard issue when asked. Use
  this whenever a CI run on a feature/renovate/* branch finishes, when asked to triage
  Renovate or dependency pull requests, or when asked to go through the Dependency
  Dashboard. Merges green routine updates; never pushes or deploys.
---

# Dependency updates of zeitlos/lucity

Renovate opens the pull requests and CI builds them. You add the judgment: why a red one fails and how to fix it, what a green one changes once it is deployed, and whether an update waiting for approval is safe. You merge green routine updates yourself. A person merges everything else, and a person deploys, always.

## Never

- Merge a pull request that fails any condition under "Merge".
- Close or approve a pull request.
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

Check which tools the pod has before you rely on one. Read pull requests, comments, check runs and job logs through the GitHub connector. When a cause needs a tool you cannot get to confirm, say in your comment what you could not verify.

## When a CI run finishes

The session starts with the trigger's instructions and the event as JSON, for example:

```json
{"event":"workflow_run.completed","repository":"zeitlos/lucity","pullRequest":123,"branch":"feature/renovate/<name>","commit":"<sha>","workflow":"CI","conclusion":"failure","url":"<run url>","sender":"renovate[bot]"}
```

Later runs of the same pull request (a re-run, a rebase, a new push) continue the same session, so you may already know this pull request. Steps:

1. **`conclusion` is `cancelled` or `skipped`**: stop. No comment, no message. CI cancels a run when a newer one starts on the same branch.
2. **Check that it is yours to look at**: the pull request is open, its branch starts with the `branchPrefix` from `renovate.json`, and `commit` is still its head commit. If the head has moved on, stop: the run for the new head will come.
3. **Check the marker**: every comment you post ends with `<!-- deps-agent:<head commit> -->`. If a comment with this commit's marker exists, stop. An earlier commit's comment does not count: post a new one and leave the old one alone.
4. **`success`**: merge it if every condition under "Merge" holds, then post the green comment either way.
5. **`failure`**: find the cause, then post the red comment.

### Merge

Merge only when all of these hold, checked right before you merge, not from memory:

- Renovate opened the pull request (`renovate[bot]`).
- It does not carry the `planned` label. Renovate adds that label to every update that needs a person (see `renovate.json`). Those get the green comment and wait.
- Every check on the head commit has passed, including `ci-ok`. None is pending or failed.
- `commit` is still the head commit, and GitHub reports the pull request as mergeable (no conflicts).

Squash merge, keeping Renovate's title as the commit title. If the merge is refused, don't retry: say why in the green comment.

### Green

At most four lines: whether you merged it and why not if you didn't, what changed, what it rolls out once deployed, and anything the maintainer has to do. Nothing more: no praise, no restating the pull request body, no `<details>` block.

```markdown
**<Merged. | Ready for you to merge: <the reason, e.g. labeled planned>.>** <one line: what this updates, in plain words>

- Rolls out: <what changes on the next deploy, see "What rolls out">
- Before deploying: <only if something is needed, e.g. a maintenance window; leave the line out otherwise>

<!-- deps-agent:<head commit> -->
```

For a security fix, add one line on whether the vulnerable code is reachable. Follow the import path from the vulnerable package to the code that ships: a vulnerability that only a test package reaches is not in any binary.

### Red

1. List the check runs of the CI run and read the logs of the failed jobs. Find the first real error, not the last line.
2. Check whether the files under "Where the answers are" already explain it, and confirm from the log and the diff.
3. Reproduce what you can in the clone, and read the changelog or release notes of the dependency that moved.
4. Comment. The visible part is at most five lines: the cause, the error, the fix. Everything that shows how you got there (reruns, timelines, commits you compared) goes into the `<details>` block, briefly. Leave the block out when the error speaks for itself.

```markdown
**CI fails: <the cause, at most about 15 words>.**

`<job>`: `<the one error line that matters>`

Fix: <a command, a short diff, "nothing, Renovate's next rebase fixes it", or "hold this update" with the reason>. Goes on <this branch, main, or renovate.json>.

<details><summary>How I got there</summary>

- <a fact you checked, one line each, at most five>

</details>

<!-- deps-agent:<head commit> -->
```

If the fix belongs in `renovate.json`, name the rule and why. A fix pushed to a Renovate branch by someone else stops Renovate from rebasing that branch; mention it when that matters.

## Where the answers are

Read the current files instead of relying on what you remember about this repository:

- `.github/renovate.json`: a rule's `description` says why it exists, what it groups, caps or disables, and what it rolls out. A failure that a rule already explains is a config question, not a code question.
- `.github/workflows/ci.yml`: which jobs run for which paths, with which settings, and what each step checks.
- The pull request: its body (Renovate's notes, release notes, advisory details), its diff, and Renovate's own comments if there are any.
- `main` since the branch was created: a fix may already be there, and Renovate's next rebase picks it up.

## What rolls out

Merging deploys nothing; a person deploys. What a change rolls out follows from where it sits: a service's modules and Dockerfile go into that service's image, a chart into its release, a workflow into CI only. When a rule in `renovate.json` describes a wider impact (restarts, downtime, migrations), repeat it in your comment.

## Dependency Dashboard

Only when asked. It is the open issue titled "Dependency Dashboard"; Renovate rewrites its body on every run.

1. Read the sections: updates waiting for approval, errored or rate-limited branches, and "Repository problems" (config warnings show up there).
2. For each update waiting for approval: what it is, what breaks or rolls out, what it needs before or after (a code change, a regenerate, a maintenance window), and your verdict: tick now, tick later, or hold with the reason. Base it on the changelog, the release notes, and how the repository uses the dependency (search the clone). If you cannot build it yourself, say what CI will have to prove once the pull request exists.
3. List what else needs attention: errored branches, config problems, security fixes merged since the last deploy.
4. Post one comment on the issue ending with `<!-- deps-agent:dashboard:<YYYY-MM-DD> -->`, or reply in the session if you were asked there. Tick a box only when asked for that box.

## Writing

Plain and short, like a teammate's note. Lead with the verdict. Spell out "vulnerability". No em dashes, no headings inside a comment, no restating the pull request body.
