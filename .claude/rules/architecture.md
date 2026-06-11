# Architecture & Design Principles

## Simplicity First

Prefer simplicity over abstraction. Simple code is easier to read, debug, and operate. Three similar lines of code are better than a premature abstraction. Don't build frameworks — build features.

If a design feels complicated, step back and ask: is there a simpler way? Usually there is.

## Stateless Platform

No central database. All state is derived from external systems:

- **Kubernetes**: namespaces, labels, Helm release state (Secrets), operator CRDs
- **OCI Registry (Zot)**: built images, tags, digests
- **Identity Provider (OIDC/Logto)**: users, roles, authentication, workspace metadata

If you're tempted to add a database, reconsider. The right answer is almost always to store state in one of the systems above. A read-optimized cache may be acceptable when query performance demands it, but it must be derivable from the sources above.

## Ejectability

Every feature must be ejectable to standard Kubernetes, Helm, and ArgoCD configurations. If a feature can't be represented as standard infrastructure-as-code after ejection, it doesn't belong in the platform.

Test this by asking: "If a user runs `lucity eject` right now, does this feature survive?"

## User's Repo is Sacred

The platform never writes to the user's source repository. Not a commit, not a file, not a webhook configuration file. The user's repo is read-only to the platform. All platform-managed configuration lives outside the user's repo, as the Helm release values the conductor computes and applies.

## Zero-Trust Security

Users and workspaces are hostile by default. Every user is assumed to be malicious. Never trust user-provided values. Never trust user-generated code. All inputs crossing a trust boundary must be validated and sanitized. Workspaces are hard isolation boundaries. Build-time execution of user code is the highest-risk operation on the platform.

See `security.md` for comprehensive rules covering input validation, injection prevention, workspace isolation, build-time security, runtime isolation, and platform service protection.

## Discovery Over Definition

Query Kubernetes for truth via labels and annotations. Don't define custom CRDs. Don't maintain mapping tables.

- A "Project" is namespaces with `lucity.dev/project` labels
- A "Service" is a Deployment discovered via Helm values or K8s API
- A "Database" is a CNPG Cluster CRD with platform labels

Standard `kubectl` works for everything. No special tooling needed.

## Loose Coupling

The control plane is a single binary (conductor); the one cross-service boundary, conductor ↔ cashier, talks over gRPC for commands but doesn't hold connections open for long-running operations. Use polling and observation:

- Watch the OCI registry for built images to appear
- Poll Helm/Kubernetes for rollout status
- Query Kubernetes for deployment state

Components own their domain behind narrow interfaces. For long-running work, observe state rather than holding a connection open.

## Minimal Day-2 Operations

Features should be operable by a small team. Ask:

- Does this add ongoing maintenance burden?
- Can it self-heal or does it need manual intervention?
- What happens when it breaks — does the blast radius stay small?
- Is observability built in, not bolted on?

If a feature can't be run without a dedicated on-call team, it's too complex.

## Idempotent Operations

Operations that touch external state — applying Helm releases, pushing images, creating namespaces — must be idempotent. If something already exists, detect it and handle it gracefully instead of failing.

- **Create operations**: check if the resource already exists and is in the expected state. If yes, return success. If it exists but is incomplete (partial failure), recover by completing the remaining steps.
- **Delete operations**: if the resource is already gone, return success — don't error on "not found".
- **Update operations**: verify current state before applying changes. Don't assume a clean slate.

This matters because the platform is stateless and distributed. Retries, crashes, and partial failures are normal. Every operation should be safe to repeat.

## No Backward Compatibility

Don't support legacy API versions. Just change APIs directly — no fallbacks, no deprecation shims, no version negotiation. This is an early-stage project with no external consumers. Clean breaks are better than compatibility layers.

## Don't Reinvent

If Helm, a Kubernetes operator, or the OCI registry already manages a piece of state, use it. Don't duplicate it. Don't wrap it in an unnecessary abstraction. Leverage what's already there.
