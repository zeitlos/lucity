# Security Model

## Zero-Trust Posture

Users and workspaces are malicious by default. Every user is assumed to want to harm, exploit, or escape the platform. This is not a hypothetical: Lucity runs arbitrary user code on shared infrastructure. Act accordingly.

Three trust boundaries, in order of hardness:

1. **Platform vs. user workloads** — `lucity-system` namespace is the crown jewel. User workloads must never reach it, read its secrets, or influence its behavior. This boundary must never be crossed.
2. **Workspace vs. workspace** — hard isolation. Workspace A must never see, access, or affect workspace B's data, images, repos, or workloads. No exceptions.
3. **Environment vs. environment** — soft isolation within the same workspace. Same owner, but separate blast radius. A compromised development environment must not be able to affect production.

## Input Validation

Never trust any user-provided value. Every input that crosses a trust boundary must be validated before use.

- **Kubernetes resource names** (project names, service names, environment names): validate against DNS label regex (`[a-z0-9][a-z0-9-]*[a-z0-9]`, max 63 chars). These values get concatenated into namespace names, label values, and Helm release names. A malicious name can break YAML, inject labels, or cause resource collisions.
- **Environment variable keys**: alphanumeric and underscore only (`[A-Z_][A-Z0-9_]*`). No shell metacharacters. A key like `FOO;rm -rf /` must be rejected.
- **Environment variable values**: never interpolate into shell commands or templates. Always pass as structured data. Values are opaque strings that may contain anything.
- **Git refs and branch names**: validate against safe patterns. Reject refs containing `..`, shell metacharacters, or path traversal sequences.
- **Repository URLs**: must match expected patterns (HTTPS GitHub URLs, Soft-serve SSH URLs). Reject `file://`, `javascript:`, or unexpected schemes.
- **Custom start commands**: these execute via `sh -c` in the deployment template. This is inherently dangerous. Validate strictly: no shell expansion characters beyond what's necessary. Consider this the single most dangerous user input in the system.
- **Domain names**: RFC 1123 compliant. Must not be wildcards (`*`). Must not overlap with platform domains (`lucity.cloud`, `lucity.app`). A user must not be able to claim a platform domain.
- **GraphQL inputs**: use the `@constraint` directive on all user-facing input fields. No input field that accepts free-form strings should go unvalidated.
- **Webhook payloads**: validate signatures before processing. Treat the entire payload as untrusted even after signature verification (the source repo is user-controlled).

## Injection Prevention

User-provided values will end up in YAML files, shell commands, Helm templates, Kubernetes manifests, and Git commits. Every one of these is an injection vector.

- **YAML injection**: use structured YAML marshaling (`yaml.Marshal` on `map[string]any`). Never construct YAML with `fmt.Sprintf`, string concatenation, or Go `text/template`. The packager's values generation must always use structured data, never string interpolation.
- **Command injection**: never concatenate user values into shell commands. Pass user values as environment variables to subprocesses, never as command arguments that go through shell expansion. The builder sets `BUILD_SOURCE_URL`, `BUILD_GIT_REF`, etc. as container env vars. These must be validated before use.
- **Helm template injection**: always use `| quote` for user-provided string values in Helm templates. Never use bare `{{ .Values.x }}` for any value that originates from user input. Review every template addition for unquoted user values.
- **Label/annotation injection**: Kubernetes label values are limited to 63 characters and must match `[a-z0-9A-Z][a-z0-9A-Z._-]*`. Validate before setting. A malicious label value can break label selectors across the cluster.
- **Git commit injection**: user-provided values in commit messages (service names, environment names) must not contain newlines or Git-special sequences that could alter commit metadata.

## Workspace Isolation

- **API layer**: every GraphQL query and mutation must be scoped to the authenticated user's workspace. The workspace comes from the JWT, never from user input. No query should ever accept a workspace ID as a parameter.
- **gRPC propagation**: the `X-Lucity-Workspace` header propagates workspace context to backend services. Backend services must verify this claim against the actual Kubernetes resource labels. If a request says "workspace: acme" but the target namespace has `lucity.dev/workspace: other`, reject it.
- **Namespace ownership**: before operating on any namespace, verify it carries the expected `lucity.dev/workspace` label. Never assume a namespace belongs to a workspace just because the name matches a pattern.
- **Registry isolation**: image paths must be namespaced by workspace (`registry/{workspace}/{project}/{service}`). A workspace must not be able to pull, push, or list images from another workspace's path. Registry credentials scoped per workspace.
- **GitOps repo isolation**: Soft-serve repositories are scoped per workspace. A user in workspace A must never read, write, or discover workspace B's GitOps repos.
- **ArgoCD Application isolation**: ArgoCD Applications must be labeled with workspace ownership. Operations on Applications must verify workspace labels before proceeding.

## Build-Time Security

Building user code is the highest-risk operation. The builder executes arbitrary code from user repositories.

- **Namespace isolation**: build Jobs run in `lucity-builds`, physically separated from `lucity-system` and all workload namespaces.
- **Resource limits**: strict CPU and memory limits on build pods. These are security controls preventing resource exhaustion attacks, not just resource management.
- **No privilege**: build pods run as non-root, all capabilities dropped, `allowPrivilegeEscalation: false`, seccomp profile enforced.
- **BuildKit process sandbox**: mandatory. Isolates `RUN` steps so user Dockerfile commands cannot escape the build context.
- **Kubernetes user namespaces**: enabled for build pods to map container root to unprivileged host UID.
- **No API access**: build pods must not have access to the Kubernetes API. ServiceAccount token automount disabled or ServiceAccount has zero RBAC permissions.
- **Network restrictions**: build pods should only reach the source repo (GitHub), the OCI registry (for push), and public package registries. No access to cluster-internal services, no access to `lucity-system`, no access to other workload namespaces.
- **Build timeouts**: enforced via `activeDeadlineSeconds` on Jobs. A build that runs forever is either broken or malicious. Kill it.
- **No secrets in builds**: never mount platform credentials, registry push tokens (beyond what's needed for the specific image push), or workspace secrets into build pods.

## Runtime Workload Isolation

User workloads run in per-environment namespaces. They are untrusted containers running untrusted code.

- **NetworkPolicy**: mandatory for every workload namespace. Default deny ingress and egress. Allow only: ingress from the Gateway API controller (for HTTP traffic), egress to the internet (for the workload's own external calls). Deny all cross-namespace traffic. Deny access to cluster CIDRs (metadata API, `lucity-system` services, other workload namespaces).
- **Pod security**: non-root (UID 1000), `allowPrivilegeEscalation: false`, all capabilities dropped, read-only root filesystem where possible, seccomp profile enforced.
- **Resource limits**: LimitRange in every workload namespace. Limits are a security control: they prevent a single workload from starving the node and affecting other tenants.
- **No API access**: workload pods must not have access to the Kubernetes API. No ServiceAccount token. A compromised workload must not be able to enumerate or modify cluster resources.
- **No platform access**: workloads must not be able to reach `lucity-system` services (gateway, builder, packager, deployer) directly. All platform interaction goes through the public API via the Gateway API ingress.
- **Image provenance**: workloads must only run images from the platform's OCI registry. No pulling arbitrary images from Docker Hub or other registries.

## Platform Service Protection

The `lucity-system` namespace runs all platform services. It is the most privileged namespace in the cluster.

- **NetworkPolicy**: restrict ingress to only the Gateway API controller (for external API access) and inter-service gRPC traffic within the namespace. No ingress from workload namespaces.
- **Credentials**: all platform credentials (ArgoCD tokens, Soft-serve SSH keys, registry auth, Stripe keys, OIDC secrets) stored as Kubernetes Secrets in `lucity-system`. Never copy platform credentials to workload namespaces.
- **Registry pull secrets**: when cloned into workload namespaces, must provide pull-only access scoped to the workspace's image path. Never give workloads push access or access to other workspaces' images.
- **Internal gRPC**: unauthenticated between platform services (trusted network assumption). This is acceptable only because NetworkPolicy prevents workload pods from reaching gRPC ports. If NetworkPolicy is ever relaxed, add mTLS.
- **RBAC**: platform service ServiceAccounts have only the permissions they need. The builder SA can create Jobs in `lucity-builds`. The deployer SA can manage ArgoCD Applications and namespaces. No service has cluster-admin.

## Secure Defaults

Every new feature must be secure by default. When adding any new feature, ask:

- "What happens if a malicious user controls this value?"
- "Can this be used to escape the workspace boundary?"
- "Can this be used to access another workspace's data?"
- "Can this be used to reach `lucity-system` services?"
- "Can this be used to exhaust shared resources?"

If the answer to any of these is "yes" or "maybe", fix it before shipping. Security is not a follow-up task.
