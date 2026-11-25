<script setup lang="ts">
import type { Component } from 'vue';
import { computed } from 'vue';
import { brandMark } from '@/lib/brandMarks';
import { cn } from '@/lib/utils';
import { fallbackColorIndex, workspaceColorStyle } from '@/lib/workspaceColor';

const props = withDefaults(defineProps<{
  name: string;
  brand?: string | null;
  icon?: Component;
  size?: 'sm' | 'md';
}>(), {
  brand: undefined,
  icon: undefined,
  size: 'md',
});

const mark = computed(() => {
  if (props.icon) return null;
  return (props.brand ? brandMark(props.brand) : null) ?? brandMark(props.name);
});
const letter = computed(() => props.name.trim().charAt(0).toUpperCase() || '?');
const style = computed(() => {
  if (props.icon) return undefined;
  if (mark.value) return { background: mark.value.background, color: mark.value.color };
  return workspaceColorStyle(fallbackColorIndex(props.name.trim().toLowerCase()));
});
</script>

<template>
  <span
    :class="cn(
      'flex shrink-0 items-center justify-center rounded-sm font-semibold',
      size === 'sm' ? 'size-5 text-xs' : 'size-8 text-sm',
      icon && 'bg-muted text-muted-foreground',
    )"
    :style="style"
    aria-hidden="true"
  >
    <component :is="icon" v-if="icon" :class="size === 'sm' ? 'size-3' : 'size-4.5'" />
    <svg
      v-else-if="mark"
      :viewBox="mark.viewBox ?? '0 0 24 24'"
      :class="size === 'sm' ? 'size-3' : 'size-4.5'"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path :d="mark.path" />
    </svg>
    <template v-else>{{ letter }}</template>
  </span>
</template>
