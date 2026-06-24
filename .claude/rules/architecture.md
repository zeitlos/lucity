# Architecture

The principles that shape every design decision. When a change conflicts with one of these, reconsider the change, not the principle.

## Stateless

No central database. All state is derived from external systems of record: Kubernetes (namespaces, labels, Helm release state, operator CRDs), the OCI registry (images, tags, digests), and the identity provider (users, roles, workspaces). If you reach for a database, you are almost certainly solving the wrong problem. A read-optimized cache is acceptable only when it is fully derivable from those sources.

## Ejectable

Every feature must survive ejection as standard Kubernetes and Helm config with no Lucity dependency. If a feature can't be expressed as plain infrastructure-as-code, it doesn't belong in the platform.

## Discovery over definition

Truth lives in Kubernetes, queried via `lucity.dev/*` labels and annotations. Don't invent CRDs or mapping tables when a label selector answers the question. Don't duplicate state that Helm, an operator, or the registry already owns.

## The user's repo is sacred

The platform reads source repos and writes to them never: not a commit, not a file, not a hook. All platform-managed config lives outside the user's repo, as the Helm values the control plane computes and applies.

## Zero-trust

Users and workspaces are hostile by default; the platform runs arbitrary user code on shared infrastructure. Three trust boundaries, hardest first:

1. **Platform vs. workloads** — the platform namespace is the crown jewel. Workloads must never reach it. This boundary never breaks.
2. **Workspace vs. workspace** — hard isolation. No tenant ever sees or touches another's data, images, or workloads.
3. **Environment vs. environment** — soft isolation, same owner, separate blast radius.

Never trust a user-provided value that crosses into Kubernetes (names, env keys, refs, domains, start commands), shell, or a template. Validate at the boundary; pass values as structured data, never string-interpolated. Workspace identity comes from the verified token, never from a parameter. Every new feature is secure by default: assume a malicious user controls each input and ask whether it can escape a boundary, reach another tenant, or exhaust shared resources.

## Deployment model

One Helm release per project per environment, applied imperatively and idempotently. The control plane generates chart *values*; it never touches chart templates. A change (replicas, image, variable) is a fresh release revision. Promotion copies an image digest between environments and re-applies; it never rebuilds. Operations on external state must be safe to repeat: detect existing state and converge, treat "already gone" as success.

## Simplicity

Prefer three plain lines to one premature abstraction. Build features, not frameworks. A little copying beats a little dependency. Features must be operable by a small team: self-healing over manual intervention, small blast radius, observability built in.
