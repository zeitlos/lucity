<script setup lang="ts">
import type { Mark } from '~/utils/comparisons/types';

defineProps<{
  mark: Mark;
}>();
</script>

<template>
  <svg v-if="mark === 'yes'" class="mt-0.5 size-4.5 shrink-0 text-emerald-700 dark:text-emerald-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d="M20 6 9 17l-5-5" />
  </svg>
  <svg v-else-if="mark === 'partly'" class="mt-0.5 size-4.5 shrink-0 text-amber-700 dark:text-amber-400" viewBox="0 0 24 24" aria-hidden="true">
    <circle cx="12" cy="12" r="8" fill="none" stroke="currentColor" stroke-width="2.2" />
    <path d="M12 4a8 8 0 0 1 0 16Z" fill="currentColor" />
  </svg>
  <svg v-else-if="mark === 'no'" class="mt-0.5 size-4.5 shrink-0 text-stone-500 dark:text-stone-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" aria-hidden="true">
    <path d="M6 12h12" />
  </svg>
</template>
