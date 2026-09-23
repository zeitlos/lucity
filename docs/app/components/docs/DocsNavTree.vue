<script setup lang="ts">
import type { ContentNavigationItem } from '@nuxt/content';

const route = useRoute();
const { sidebarNavigation } = useSubNavigation();

const emit = defineEmits<{ navigate: [] }>();

const overrides = ref<Record<string, boolean>>({});

function holdsCurrentPage(item: ContentNavigationItem) {
  return route.path === item.path || route.path.startsWith(`${item.path}/`);
}

function isOpen(item: ContentNavigationItem) {
  return overrides.value[item.path] ?? holdsCurrentPage(item);
}

function toggle(item: ContentNavigationItem) {
  overrides.value[item.path] = !isOpen(item);
}

/** A section links to its own index page when it has one. */
function indexPath(item: ContentNavigationItem) {
  return item.children?.some(child => child.path === item.path) ? item.path : undefined;
}

/** The index page is reachable from the section title, so drop the duplicate. */
function pages(item: ContentNavigationItem) {
  return (item.children ?? []).filter(child => child.path !== item.path);
}
</script>

<template>
  <nav class="docs-aside">
    <template v-for="item in sidebarNavigation" :key="item.path">
      <NuxtLink
        v-if="!item.children?.length"
        :to="item.path"
        class="docs-aside-link"
        :class="{ 'docs-aside-link-active': route.path === item.path }"
        @click="emit('navigate')"
      >
        {{ item.title }}
      </NuxtLink>

      <div
        v-else
        class="docs-aside-section"
      >
        <div class="docs-aside-head">
          <NuxtLink
            v-if="indexPath(item)"
            :to="indexPath(item)"
            class="docs-aside-link docs-aside-head-label"
            :class="{ 'docs-aside-link-active': route.path === item.path }"
            @click="overrides[item.path] = true; emit('navigate')"
          >
            {{ item.title }}
          </NuxtLink>
          <button
            v-else
            type="button"
            class="docs-aside-link docs-aside-head-label"
            @click="toggle(item)"
          >
            {{ item.title }}
          </button>

          <button
            type="button"
            class="docs-aside-toggle"
            :aria-expanded="isOpen(item)"
            :aria-label="isOpen(item) ? `Collapse ${item.title}` : `Expand ${item.title}`"
            @click="toggle(item)"
          >
            <UIcon
              name="i-lucide-chevron-right"
              class="size-4 transition-transform duration-200"
              :class="{ 'rotate-90': isOpen(item) }"
            />
          </button>
        </div>

        <ul
          v-show="isOpen(item)"
          class="docs-aside-children"
        >
          <li
            v-for="child in pages(item)"
            :key="child.path"
          >
            <NuxtLink
              :to="child.path"
              class="docs-aside-link docs-aside-child"
              :class="{ 'docs-aside-link-active': route.path === child.path }"
              @click="emit('navigate')"
            >
              {{ child.title }}
            </NuxtLink>
          </li>
        </ul>
      </div>
    </template>
  </nav>
</template>

<style scoped>
.docs-aside {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  font-size: 0.875rem;
}

.docs-aside-link {
  display: block;
  padding: 0.375rem 0.625rem;
  border-radius: 0.5rem;
  color: var(--ui-text-toned);
  font-weight: 500;
  text-decoration: none;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.docs-aside-link:hover {
  color: var(--ui-text-highlighted);
  background: var(--ui-bg-elevated);
}

.docs-aside-link-active,
.docs-aside-link-active:hover {
  color: var(--ui-text-highlighted);
  background: color-mix(in oklab, var(--ui-primary) 14%, var(--ui-bg-elevated));
  font-weight: 600;
}

.docs-aside-section {
  margin-top: 0.125rem;
}

.docs-aside-head {
  display: flex;
  align-items: center;
  gap: 0.125rem;
}

.docs-aside-head-label {
  flex: 1;
  text-align: start;
  cursor: pointer;
}

.docs-aside-toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.3125rem;
  margin-inline-end: 0.125rem;
  border-radius: 0.375rem;
  color: var(--ui-text-dimmed);
  cursor: pointer;
  transition: color 0.15s ease, background-color 0.15s ease;
}

.docs-aside-toggle:hover {
  color: var(--ui-text-highlighted);
  background: var(--ui-bg-accented);
}

.docs-aside-toggle:focus-visible {
  outline: 2px solid var(--ui-primary);
  outline-offset: 1px;
}

.docs-aside-children {
  margin: 0.125rem 0 0.25rem 0.75rem;
  padding-inline-start: 0.625rem;
  border-inline-start: 1px solid var(--ui-border);
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.docs-aside-child {
  position: relative;
}

.docs-aside-child.docs-aside-link-active::before {
  content: '';
  position: absolute;
  inset-block: 0.25rem;
  inset-inline-start: -0.6875rem;
  width: 2px;
  border-radius: 999px;
  background: var(--ui-primary);
}
</style>

<style>
/* Unscoped: the scoped compiler drops the descendant part of :global(.dark). */
.dark .docs-aside-link-active,
.dark .docs-aside-link-active:hover {
  background: color-mix(in oklab, white 7%, var(--ui-bg-elevated));
  color: var(--ui-text-highlighted);
}
</style>
