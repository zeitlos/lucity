# Vue Conventions

## Framework

Vue 3 + Vite + TypeScript. Always use `<script setup lang="ts">`.

## Components

- PascalCase filenames: `BaseButton.vue`, `ProjectCard.vue`
- `Base*` prefix for atomic/reusable primitives
- Feature or domain prefix for page-specific components
- Polymorphic components: use `useAttrs()` to detect `to`/`href` and render `RouterLink`, `<a>`, or `<button>`

## Props & Events

```vue
<script setup lang="ts">
const props = defineProps<{
  color?: ButtonColors;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update', value: string): void;
}>();
</script>
```

## Composables

- `use<Name>` convention in `composables/` directory
- Small, focused, heavily composed
- Examples: `useAuth`, `useProjects`, `useToast`, `useConfirmation`, `useLoading`

## State Management

- Apollo cache for server state (GraphQL)
- Composables for local/shared state
- `provide`/`inject` for hierarchical state
- No Vuex or Pinia

## Styling

- Tailwind CSS v4 with `@tailwindcss/vite` plugin
- `cn()` helper (clsx + tailwind-merge) for conditional classes
- Icons: `lucide-vue-next`

## UI Libraries

- shadcn-vue + Reka UI (`components/ui/`)
- **Never use `export` inside `<script setup>`** — `<script setup>` cannot contain ES module exports. If a component needs to export a value (e.g. `cva` variants), use a separate `<script lang="ts">` block for the export and keep component logic in `<script setup lang="ts">`
- shadcn-vue components live in `src/components/ui/` and are exempt from `vue/multi-word-component-names` via eslint config
- Never manually edit shadcn-vue components unless fixing a build/lint issue — regenerate with the CLI instead

## Imports

- `@/` alias for `src/` directory
- `import type { ... }` for type-only imports

## GraphQL

- Codegen from gateway schema
- `graphql()` template tag for queries and mutations
- Fragment-based reuse for shared fields
- Per-page `graphql.ts` files for queries/mutations

## ESLint (enforced)

- Single quotes
- Semicolons required
- Max 3 attributes per single-line element
- Props don't require defaults
- Always run `npx eslint .` from `services/dashboard/` before committing frontend changes
- shadcn-vue overrides in `eslint.config.ts`: `multi-word-component-names`, `no-explicit-any`, `no-unused-vars` are off for `src/components/ui/**`
