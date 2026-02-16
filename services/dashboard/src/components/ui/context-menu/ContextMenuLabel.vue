<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ContextMenuLabel, type ContextMenuLabelProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ContextMenuLabelProps & { class?: HTMLAttributes['class']; inset?: boolean }>();
</script>

<template>
  <ContextMenuLabel :class="cn('px-2 py-1.5 text-sm font-semibold text-foreground', inset && 'pl-8', props.class)">
    <slot />
  </ContextMenuLabel>
</template>
