<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <span
    :class="cn(
      'inline-block rounded border border-border bg-card px-2 py-0.5 font-mono text-[0.6rem] font-semibold uppercase leading-none tracking-widest text-muted-foreground',
      props.class,
    )"
  >
    <slot />
  </span>
</template>
