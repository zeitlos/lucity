# GraphQL Conventions

## Approach

Schema-first with gqlgen code generation.

## Schema Organization

Domain-split files in `services/conductor/internal/api/graphql/schema/`:

- `schema.graphqls` — base types, directives, scalars, empty Query/Mutation
- `project.graphqls`, `environment.graphqls`, `service.graphqls`, etc. — domain files extend Query/Mutation

```graphql
extend type Query {
  projects: [Project!]! @hasRole(role: [USER])
  project(id: ID!): Project @hasRole(role: [USER])
}

extend type Mutation {
  createProject(input: CreateProjectInput!): Project! @hasRole(role: [USER])
  deleteProject(id: ID!): Boolean! @hasRole(role: [ADMIN])
}
```

## Authorization

`@hasRole(role: [Role!]!)` directive on fields. Roles:

- `ANONYMOUS` — public, no auth required
- `USER` — authenticated user
- `ADMIN` — admin privileges

## Multi-Tenant (Workspaces)

Each Lucity instance supports multiple workspaces. A workspace is the tenant boundary. Workspace context comes from the JWT token — resolvers read the active workspace from auth claims, not from a header or URL parameter.

## Input Validation

`@constraint(constraint: String!)` directive on input fields/arguments.

## Custom Scalars

`Time`, `Duration`

## gqlgen Directives

`@goModel`, `@goField`, `@goTag`, `@goExtraField` — for mapping GraphQL types to Go structs.

## Vendor-Agnostic Naming

The GraphQL API is a complete abstraction over the underlying technology. **Never leak implementation details into the schema.** The consumer should not know or care whether the platform uses Helm, CloudNativePG, Zot, or any other tool.

- `rolloutHealth`, not `helmStatus`
- `deployment`, not `helmRelease`
- `database`, not `cnpgCluster`
- `registry`, not `zotRegistry`

This applies to type names, field names, enum values, and descriptions. Implementation-specific names belong in Go code, not in the API surface.

## Resolvers

Thin resolvers that delegate to the conductor client and domain packages. Type conversion in `convert.go` files using `convert<Type>` functions.

## Code Generation

From the conductor service directory:

```sh
go generate ./internal/api/graphql/resolver.go
```

Dashboard TypeScript types:

```sh
cd services/dashboard && npm run codegen
```

Run `npm run codegen` after changing the conductor GraphQL schema or dashboard query/mutation definitions. This regenerates `src/gql/graphql.ts` with typed document nodes, result types, and variable types.
