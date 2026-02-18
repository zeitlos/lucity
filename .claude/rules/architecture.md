# Architecture & Design Principles

## Simplicity First

Prefer simplicity over abstraction. Simple code is easier to read, debug, and operate. Three similar lines of code are better than a premature abstraction. Don't build frameworks — build features.

If a design feels complicated, step back and ask: is there a simpler way? Usually there is.

## Stateless Platform

No central database. All state is derived from external systems:

- **Git (Soft-serve)**: GitOps repos, Helm values, environment configuration
- **Kubernetes**: namespaces, labels, ArgoCD Applications, operator CRDs
- **OCI Registry (Zot)**: built images, tags, digests
- **Identity Provider (OIDC)**: users, roles, authentication

If you're tempted to add a database, reconsider. The right answer is almost always to store state in one of the systems above. A read-optimized cache may be acceptable when query performance demands it, but it must be derivable from the sources above.

## Ejectability

Every feature must be ejectable to standard Kubernetes, Helm, and ArgoCD configurations. If a feature can't be represented as standard infrastructure-as-code after ejection, it doesn't belong in the platform.

Test this by asking: "If a user runs `lucity eject` right now, does this feature survive?"

## User's Repo is Sacred

The platform never writes to the user's source repository. Not a commit, not a file, not a webhook configuration file. The user's repo is read-only to the platform. All platform-managed configuration lives in the GitOps repository.

## Discovery Over Definition

Query Kubernetes for truth via labels and annotations. Don't define custom CRDs. Don't maintain mapping tables.

- A "Project" is namespaces with `lucity.dev/project` labels
- A "Service" is a Deployment discovered via Helm values or K8s API
- A "Database" is a CNPG Cluster CRD with platform labels

Standard `kubectl` works for everything. No special tooling needed.

## Loose Coupling

Services communicate via gRPC for commands but don't hold connections for long-running operations. Use polling and observation:

- Watch the OCI registry for built images to appear
- Poll ArgoCD for sync status
- Query Kubernetes for deployment state

Services don't need to know about each other's internals. Each service owns its domain and exposes it via gRPC.

## Minimal Day-2 Operations

Features should be operable by a small team. Ask:

- Does this add ongoing maintenance burden?
- Can it self-heal or does it need manual intervention?
- What happens when it breaks — does the blast radius stay small?
- Is observability built in, not bolted on?

If a feature can't be run without a dedicated on-call team, it's too complex.

## Idempotent Operations

Operations that touch external state — creating repos, deploying ArgoCD apps, pushing images, creating namespaces — must be idempotent. If something already exists, detect it and handle it gracefully instead of failing.

- **Create operations**: check if the resource already exists and is in the expected state. If yes, return success. If it exists but is incomplete (partial failure), recover by completing the remaining steps.
- **Delete operations**: if the resource is already gone, return success — don't error on "not found".
- **Update operations**: verify current state before applying changes. Don't assume a clean slate.

This matters because the platform is stateless and distributed. Retries, crashes, and partial failures are normal. Every operation should be safe to repeat.

## Don't Reinvent

If ArgoCD, Helm, a Kubernetes operator, or the OCI registry already manages a piece of state, use it. Don't duplicate it. Don't wrap it in an unnecessary abstraction. Leverage what's already there.
