<script setup lang="ts">
import type { MetricWindow } from '@/gql/graphql';
import { metricWindows } from '@/lib/metrics';
import { cn } from '@/lib/utils';

defineProps<{
  splitLabel: string;
  legend: { key: string; label: string; color: string }[];
}>();

const split = defineModel<boolean>('split', { required: true });
const selected = defineModel<MetricWindow>('window', { required: true });
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="flex flex-wrap items-center gap-4">
      <div class="inline-flex overflow-hidden rounded-md border text-xs">
        <button
          type="button"
          :class="cn(
            'px-2.5 py-1 text-muted-foreground transition-colors hover:text-foreground',
            !split && 'bg-muted font-medium text-foreground',
          )"
          @click="split = false"
        >
          Total
        </button>
        <button
          type="button"
          :class="cn(
            'border-l px-2.5 py-1 text-muted-foreground transition-colors hover:text-foreground',
            split && 'bg-muted font-medium text-foreground',
          )"
          @click="split = true"
        >
          {{ splitLabel }}
        </button>
      </div>

      <div v-if="legend.length" class="flex flex-wrap items-center gap-x-3.5 gap-y-1 text-xs text-muted-foreground">
        <span v-for="item in legend" :key="item.key" class="inline-flex items-center gap-1.5">
          <span class="size-[7px] rounded-full" :style="{ background: item.color }" />
          {{ item.label }}
        </span>
      </div>
    </div>

    <div class="inline-flex overflow-hidden rounded-md border text-xs">
      <button
        v-for="option in metricWindows"
        :key="option.value"
        type="button"
        :class="cn(
          'px-2.5 py-1 text-muted-foreground transition-colors not-first:border-l hover:text-foreground',
          selected === option.value && 'bg-muted font-medium text-foreground',
        )"
        @click="selected = option.value"
      >
        {{ option.label }}
      </button>
    </div>
  </div>
</template>
