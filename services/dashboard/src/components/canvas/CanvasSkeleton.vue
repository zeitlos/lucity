<script setup lang="ts">
import { Plus } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { gridSize } from '@/composables/useCanvasLayout';

const emit = defineEmits<{
  (e: 'create'): void;
}>();

const nodes = [
  { name: 'w-20', lines: ['w-44', 'w-28'] },
  { name: 'w-24', lines: [] },
  { name: 'w-28', lines: [] },
];
</script>

<template>
  <div class="relative h-full w-full overflow-hidden">
    <svg class="absolute inset-0 size-full" aria-hidden="true">
      <pattern
        id="canvas-skeleton-dots"
        :width="gridSize"
        :height="gridSize"
        patternUnits="userSpaceOnUse"
        patternTransform="translate(-0.5,-0.5)"
      >
        <circle cx="0.5" cy="0.5" r="0.5" fill="#81818a" />
      </pattern>
      <rect width="100%" height="100%" fill="url(#canvas-skeleton-dots)" />
    </svg>

    <div class="absolute inset-0 flex items-center justify-center">
      <div class="flex items-start gap-12">
        <div
          v-for="node in nodes"
          :key="node.name"
          class="flex w-72 shrink-0 flex-col rounded-xl border bg-card px-6 py-5 shadow-sm"
        >
          <div class="flex items-center gap-3">
            <Skeleton class="size-7 rounded-full" />
            <Skeleton :class="['h-4', node.name]" />
          </div>
          <div v-if="node.lines.length" class="mt-3 space-y-1">
            <div v-for="line in node.lines" :key="line" class="flex h-4 items-center gap-1.5">
              <Skeleton class="size-3 rounded-full" />
              <Skeleton :class="['h-3', line]" />
            </div>
          </div>
          <div class="grow basis-3" />
          <div class="mt-1 border-t border-border/50 pt-4">
            <Skeleton class="h-5 w-16 rounded-full" />
          </div>
        </div>
      </div>
    </div>

    <div class="absolute right-4 top-4 flex items-center gap-2">
      <Skeleton class="size-9 rounded-none" />
      <Skeleton class="size-9 rounded-none" />
      <Button variant="outline" @click="emit('create')">
        <Plus />
        New
      </Button>
    </div>
  </div>
</template>
