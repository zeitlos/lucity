# Frontend (Vue 3)

`<script setup lang="ts">` throughout. Standard Vue/TS idioms assumed. Project-specific rules:

- **No `export` inside `<script setup>`** — it's not a module context. If a component must export something (e.g. `cva` variants), use a separate `<script lang="ts">` block.
- **Global context, not local pickers.** Project and environment context come from shared composables (the header breadcrumb / environment switcher are the single source). Never add a local project or environment selector to a page.
- **SPA navigation only.** App-local links use `RouterLink` / `router.push`, never `<a href>` (full reload). For UI-library components that render `<a>`, use `as-child` with a `RouterLink` slot.
- **Server state in the Apollo cache, shared local state in composables.** No Vuex/Pinia.
- **shadcn-vue components are generated** — regenerate via the CLI rather than hand-editing, except to fix a build/lint break.
- Style is enforced by ESLint/Prettier; don't restate it here, just run the linter before finishing frontend work.
