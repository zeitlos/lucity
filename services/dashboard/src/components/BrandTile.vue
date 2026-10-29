<script setup lang="ts">
import { computed } from 'vue';
import { cn } from '@/lib/utils';
import { fallbackColorIndex, workspaceColorStyle } from '@/lib/workspaceColor';

const props = withDefaults(defineProps<{
  name: string;
  size?: 'sm' | 'md';
}>(), {
  size: 'md',
});

const letter = computed(() => props.name.trim().charAt(0).toUpperCase() || '?');
const style = computed(() => workspaceColorStyle(fallbackColorIndex(props.name.trim().toLowerCase())));
</script>

<template>
  <span
    :class="cn(
      'flex shrink-0 items-center justify-center rounded-sm font-semibold',
      size === 'sm' ? 'size-5 text-xs' : 'size-8 text-sm',
    )"
    :style="style"
    aria-hidden="true"
  >{{ letter }}</span>
</template>
