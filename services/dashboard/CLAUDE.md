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
- Per-page query files: `src/pages/<domain>/graphql.ts`
- Fragment-based reuse for shared fields

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
