<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  title?: string;
  items?: Array<{ name: string; value: string | number; color: string }>;
}>();
</script>

<template>
  <div :class="cn('rounded-lg border bg-background p-2 shadow-sm', props.class)">
    <div v-if="title" class="mb-1 text-xs font-medium text-muted-foreground">
      {{ title }}
    </div>
    <div class="flex flex-col gap-0.5">
      <div v-for="item in items" :key="item.name" class="flex items-center gap-2 text-xs">
        <div class="h-2.5 w-2.5 shrink-0 rounded-[2px]" :style="{ backgroundColor: item.color }" />
        <span class="text-muted-foreground">{{ item.name }}</span>
        <span class="ml-auto font-mono font-medium tabular-nums text-foreground">{{ item.value }}</span>
      </div>
    </div>
    <slot />
  </div>
</template>
