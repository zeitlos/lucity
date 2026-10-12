<script setup lang="ts">
import type { MetricLine, MetricMarker, MetricReading } from '@/lib/metrics';
import { cn } from '@/lib/utils';
import MetricLineChart from './MetricLineChart.vue';

withDefaults(
  defineProps<{
    label: string;
    readings: MetricReading[];
    unit: string | null;
    split: boolean;
    focusLabel: string | null;
    lines: MetricLine[];
    from: number;
    to: number;
    max: number | null;
    ticks: string[];
    markers?: MetricMarker[];
    editLabel?: string;
  }>(),
  { markers: () => [], editLabel: 'Edit limit' },
);

const emit = defineEmits<{
  (e: 'edit'): void;
}>();

const cursor = defineModel<number | null>('cursor', { default: null });
</script>

<template>
  <div class="flex min-w-0 flex-col gap-3.5 bg-card px-4 pb-3 pt-4">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="text-xs font-medium text-muted-foreground">{{ label }}</div>
        <div class="mt-0.5 flex flex-wrap items-baseline gap-x-1.5">
          <span :class="cn('inline-flex flex-wrap items-baseline gap-x-3.5', !split && 'whitespace-nowrap')">
            <span v-for="reading in readings" :key="reading.key" class="inline-flex items-baseline gap-1.5">
              <span v-if="split" class="size-[7px] -translate-y-1 rounded-full" :style="{ background: reading.color }" />
              <span
                :class="cn('font-semibold tabular-nums tracking-tight', split ? 'text-2xl' : 'text-[28px] leading-9')"
              >{{ reading.text }}</span>
            </span>
          </span>
          <span v-if="unit" class="whitespace-nowrap text-[13px] text-muted-foreground">{{ unit }}</span>
        </div>
      </div>
      <span
        v-if="focusLabel"
        class="shrink-0 rounded-md bg-muted px-1.5 py-0.5 text-xs font-medium tabular-nums"
      >{{ focusLabel }}</span>
      <button
        v-else
        type="button"
        class="shrink-0 py-0.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
        @click="emit('edit')"
      >
        {{ editLabel }}
      </button>
    </div>

    <MetricLineChart
      v-model:cursor="cursor"
      :series="lines"
      :from="from"
      :to="to"
      :max="max"
      :height="150"
      :stroke-width="2"
      :area="lines.length === 1"
      :limit="max != null"
      :markers="markers"
      grid
    />

    <div class="-mt-1.5 flex justify-between text-[10px] text-muted-foreground">
      <span v-for="(tick, index) in ticks" :key="index">{{ tick }}</span>
    </div>
  </div>
</template>
