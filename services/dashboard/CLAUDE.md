# Dashboard

Vue 3 SPA for managing projects, environments, services, and deployments.

## Run

```sh
npm run dev      # Vite dev server
npm run build    # Production build
npm run lint     # ESLint with auto-fix
npm run codegen  # Regenerate GraphQL TypeScript types
```

Requires gateway running on `:8080`.

## Key URLs

- Dashboard: http://localhost:5173/
- GraphQL proxy: `/graphql` → gateway :8080

## Architecture

Vue 3 + Vite + TypeScript SPA with Vue Router. Apollo Client for GraphQL.

### GraphQL

- Schema source: `../gateway/graphql/schema/*.graphqls`
- Queries and mutations are defined **inline in the consuming component or composable** using the `graphql()` tagged template from `@/gql`. Codegen ([.graphqlrc.yaml](.graphqlrc.yaml)) scans `./src/**/*.vue` and `./src/**/*.ts` for these calls.
- `npm run codegen` generates `src/gql/{gql,graphql,index}.ts` — never hand-edit these files.
- Input types, result types, and enums (e.g. `DnsStatus`, `DeployPhase`) are imported from `@/gql/graphql`. The `graphql()` helper itself is imported from `@/gql`.
- `<script setup>` cannot `export`, so inline documents are declared as local `const`s. If a document needs to be shared across files, put it in a plain `.ts` module and import it.

Example:

```ts
import { graphql } from '@/gql';
import type { GenerateDomainInput } from '@/gql/graphql';

const GenerateDomainDocument = graphql(`
  mutation GenerateDomain($input: GenerateDomainInput!) {
    generateDomain(input: $input) { hostname type dnsStatus tlsStatus }
  }
`);
```

After editing any `graphql()` call or changing the gateway schema, run `npm run codegen`.

### Key Composables

- `useAuth` — authentication state
- `useProjects` — project list and selection
- `useToast` — notification queue

### UI

- shadcn-vue + Reka UI components in `src/components/ui/`
- Tailwind CSS v4
- lucide-vue-next icons
- `cn()` helper in `src/lib/utils.ts` for conditional classes

## Typecheck

```sh
npm run type-check
```
