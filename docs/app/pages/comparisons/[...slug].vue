<script setup lang="ts">
import { findComparison } from '~/utils/comparisons';

definePageMeta({
  layout: 'landing',
  header: false,
  footer: false,
});

const route = useRoute();
const slug = computed(() => {
  const parts = route.params.slug;
  return Array.isArray(parts) ? parts.join('/') : parts ?? '';
});
const comparison = computed(() => findComparison(slug.value));
</script>

<template>
  <ComparePage v-if="comparison" :key="comparison.slug" :comparison="comparison" />
  <SolutionArticle
    v-else
    section="Comparisons"
    section-to="/comparisons"
    has-index
  />
</template>
