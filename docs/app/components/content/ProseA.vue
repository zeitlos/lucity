<script setup lang="ts">
const props = defineProps<{
  href?: string;
  target?: string;
  /** Nuxt Content puts rel="nofollow" on external links while parsing. */
  rel?: string | string[];
  class?: unknown;
  ui?: { base?: string };
}>();

const { load, find } = useDocPreviews();

await load();

const preview = computed(() => find(props.href));

/** Anything pointing off the site opens in a new tab and says so. */
const isExternal = computed(() => /^https?:\/\//.test(props.href ?? ''));

const linkTarget = computed(() => isExternal.value ? (props.target ?? '_blank') : props.target);

const linkRel = computed(() => {
  const parsed = Array.isArray(props.rel) ? props.rel : props.rel ? [props.rel] : [];
  const values = isExternal.value ? [...parsed, 'noopener', 'noreferrer'] : parsed;
  return values.length ? [...new Set(values)].join(' ') : undefined;
});
</script>

<template>
  <UPopover
    v-if="preview"
    mode="hover"
    :open-delay="220"
    :close-delay="80"
    :ui="{ content: 'w-80' }"
  >
    <ULink
      :href="props.href"
      :target="props.target"
      class="doc-link"
      raw
    >
      <span class="doc-link-text"><slot /></span>
      <UIcon
        name="i-lucide-book-open-text"
        class="doc-link-icon"
        aria-hidden="true"
      />
    </ULink>

    <template #content>
      <ULink
        :href="props.href"
        class="doc-preview"
        raw
      >
        <p class="doc-preview-title">{{ preview.title }}</p>
        <p
          v-if="preview.description"
          class="doc-preview-text"
        >
          {{ preview.description }}
        </p>
        <span class="doc-preview-cta">
          Learn more
          <UIcon
            name="i-lucide-arrow-right"
            class="size-3.5"
          />
        </span>
      </ULink>
    </template>
  </UPopover>

  <ULink
    v-else
    :href="props.href"
    :target="linkTarget"
    :rel="linkRel"
    class="doc-link"
    raw
  >
    <span class="doc-link-text"><slot /></span>
    <UIcon
      v-if="isExternal"
      name="i-lucide-arrow-up-right"
      class="doc-link-icon doc-link-icon-external"
      aria-hidden="true"
    />
    <span
      v-if="isExternal"
      class="sr-only"
    >(opens in a new tab)</span>
  </ULink>
</template>

<style scoped>
.doc-preview {
  display: block;
  padding: 0.875rem 0.9375rem;
  border-radius: inherit;
}

.doc-preview:hover .doc-preview-cta {
  gap: 0.4375rem;
}

.doc-preview-title {
  margin: 0;
  font-weight: 600;
  font-size: 0.9375rem;
  color: var(--ui-text-highlighted);
}

.doc-preview-text {
  margin: 0.25rem 0 0;
  font-size: 0.8125rem;
  line-height: 1.5;
  color: var(--ui-text-muted);
  text-wrap: pretty;
}

.doc-preview-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.75rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--ui-primary);
  transition: gap 0.15s ease;
}
</style>
