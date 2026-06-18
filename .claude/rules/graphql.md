# GraphQL

Schema-first with gqlgen codegen. Project-specific rules:

- **Vendor-agnostic schema.** The API is a complete abstraction over the implementation. Never leak the underlying tech into type names, fields, enums, or descriptions: `rolloutHealth` not `helmStatus`, `database` not `cnpgCluster`, `registry` not `zotRegistry`. Implementation names belong in Go, not the API surface.
- **Authorization is mandatory and explicit.** Every field carries the role directive; an undirected root field is a bug. Workspace scope comes from the verified token, never from a query parameter or argument.
- **Validate every free-form input** with the constraint directive. No unvalidated user string reaches a resolver.
- **Thin resolvers.** Delegate to the conductor client and domain packages; keep type conversion in the `convert*` helpers.
- **Names stay consistent across layers** (schema, Go, proto): plural for lists, singular + id for one, action verbs for mutations, no obscure abbreviations.
- Regenerate after schema or query changes — both the Go side and the dashboard types.
