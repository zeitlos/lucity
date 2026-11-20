<script setup lang="ts">
import icons from './icons.json';

const props = withDefaults(defineProps<{
  name: string;
  size?: number;
  label?: string;
}>(), {
  size: 24,
  label: undefined,
});

const icon = computed(() => (icons as Record<string, { body: string; viewBox?: string }>)[props.name]!);
</script>

<template>
  <svg
    class="shrink-0"
    :width="size"
    :height="size"
    :viewBox="icon.viewBox ?? '0 0 24 24'"
    :role="label ? 'img' : undefined"
    :aria-label="label"
    :aria-hidden="label ? undefined : 'true'"
    v-html="icon.body"
  />
</template>
