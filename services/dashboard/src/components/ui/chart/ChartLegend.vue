<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  items?: Array<{ name: string; color: string }>;
}>();
</script>

<template>
  <div :class="cn('flex items-center gap-4', props.class)">
    <div v-for="item in items" :key="item.name" class="flex items-center gap-1.5 text-xs">
      <div class="h-2 w-2 shrink-0 rounded-[2px]" :style="{ backgroundColor: item.color }" />
      <span class="text-muted-foreground">{{ item.name }}</span>
    </div>
    <slot />
  </div>
</template>
