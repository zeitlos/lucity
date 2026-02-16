# GraphQL Conventions

## Approach

Schema-first with gqlgen code generation.

## Schema Organization

Domain-split files in `services/gateway/graphql/schema/`:

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

## Single-Tenant

Each Lucity instance serves one organization. There is no organization header or multi-tenant scoping in the schema. The instance IS the organization.

## Input Validation

`@constraint(constraint: String!)` directive on input fields/arguments.

## Custom Scalars

`Time`, `Duration`

## gqlgen Directives

`@goModel`, `@goField`, `@goTag`, `@goExtraField` — for mapping GraphQL types to Go structs.

## Resolvers

Thin resolvers that delegate to the `handler` package. Type conversion in `convert.go` files using `convert<Type>` functions.

## Code Generation

From the gateway service directory:

```sh
go generate ./graphql/resolver.go
```

Dashboard TypeScript types:

```sh
cd services/dashboard && npm run codegen
```
