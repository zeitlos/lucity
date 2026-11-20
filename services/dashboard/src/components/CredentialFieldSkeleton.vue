<script setup lang="ts">
import { Skeleton } from '@/components/ui/skeleton';

defineProps<{
  rows: number;
}>();

const labels = ['w-24', 'w-10', 'w-8', 'w-16', 'w-10', 'w-16'];
const values = ['w-28', 'w-56', 'w-10', 'w-8', 'w-10', 'w-28'];
</script>

<template>
  <div class="space-y-1.5">
    <div
      v-for="row in rows"
      :key="row"
      class="flex h-10 items-center gap-2 rounded-md bg-muted/40 px-3"
    >
      <span class="w-28 shrink-0">
        <Skeleton :class="['h-3', labels[(row - 1) % labels.length]]" />
      </span>
      <Skeleton :class="['h-3', values[(row - 1) % values.length]]" />
    </div>
  </div>
</template>
