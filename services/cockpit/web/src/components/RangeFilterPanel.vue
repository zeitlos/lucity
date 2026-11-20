<script setup lang="ts">
import { computed } from 'vue';

import type { Range, RangeFilter, Workspace } from '@/lib/workspaces';

const STEPS = 200;
const BINS = 40;

const thumb = [
  'pointer-events-none absolute inset-0 h-4 w-full appearance-none bg-transparent',
  '[&::-webkit-slider-runnable-track]:h-4',
  '[&::-webkit-slider-thumb]:pointer-events-auto [&::-webkit-slider-thumb]:size-4 [&::-webkit-slider-thumb]:cursor-grab',
  '[&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:border-2',
  '[&::-webkit-slider-thumb]:border-foreground [&::-webkit-slider-thumb]:bg-card [&::-webkit-slider-thumb]:shadow-xs',
];

const props = defineProps<{
  filter: RangeFilter;
  workspaces: Workspace[];
}>();

const range = defineModel<Range>({ required: true });

function transform(value: number): number {
  return props.filter.scale === 'log' ? Math.log1p(value) : value;
}

function inverse(value: number): number {
  return props.filter.scale === 'log' ? Math.expm1(value) : value;
}

const values = computed(() =>
  props.workspaces.map((workspace) => props.filter.value(workspace)).filter((value): value is number => value != null),
);

const extent = computed(() => ({ min: Math.min(...values.value), max: Math.max(...values.value) }));
const low = computed(() => transform(extent.value.min));
const span = computed(() => transform(extent.value.max) - low.value);

function position(value: number | null, fallback: number): number {
  if (value == null || span.value === 0) return fallback;
  return Math.min(Math.max(Math.round(((transform(value) - low.value) / span.value) * STEPS), 0), STEPS);
}

function valueAt(step: number): number {
  const value = inverse(low.value + (step / STEPS) * span.value);
  return props.filter.integer ? Math.round(value) : value;
}

const start = computed(() => position(range.value.min, 0));
const end = computed(() => position(range.value.max, STEPS));

function setStart(event: Event) {
  const input = event.target as HTMLInputElement;
  const step = Math.min(Number(input.value), end.value);
  const value = valueAt(step);

  input.value = String(step);
  range.value = { ...range.value, min: value <= extent.value.min ? null : value };
}

function setEnd(event: Event) {
  const input = event.target as HTMLInputElement;
  const step = Math.max(Number(input.value), start.value);
  const value = valueAt(step);

  input.value = String(step);
  range.value = { ...range.value, max: value >= extent.value.max ? null : value };
}

const bins = computed(() => {
  const counts = Array.from({ length: BINS }, () => 0);

  for (const value of values.value) {
    const index = span.value === 0 ? 0 : Math.min(Math.floor(((transform(value) - low.value) / span.value) * BINS), BINS - 1);
    counts[index] = (counts[index] ?? 0) + 1;
  }

  const peak = Math.sqrt(Math.max(...counts, 1));

  return counts.map((count, index) => ({
    count,
    height: count === 0 ? 0 : Math.max((Math.sqrt(count) / peak) * 100, 6),
    center: ((index + 0.5) / BINS) * STEPS,
  }));
});

function percent(step: number): string {
  return `${(step / STEPS) * 100}%`;
}
</script>

<template>
  <div class="flex w-80 flex-col gap-2 p-2">
    <p v-if="values.length === 0" class="py-6 text-center text-muted-foreground">No values</p>
    <template v-else>
      <div class="mx-2 flex h-16 items-end gap-px" aria-hidden="true">
        <span
          v-for="(bin, index) in bins"
          :key="index"
          class="flex-1 rounded-t-[2px]"
          :class="bin.center >= start && bin.center <= end ? 'bg-foreground/70' : 'bg-foreground/15'"
          :style="{ height: `${bin.height}%` }"
        />
      </div>

      <div class="relative h-4">
        <div class="absolute inset-x-2 top-1/2 h-1 -translate-y-1/2 rounded-full bg-muted">
          <div class="absolute inset-y-0 rounded-full bg-foreground" :style="{ left: percent(start), right: percent(STEPS - end) }" />
        </div>
        <input
          type="range"
          min="0"
          :max="STEPS"
          :value="start"
          :class="[thumb, start > STEPS / 2 ? 'z-10' : '']"
          :aria-label="`Minimum ${filter.label}`"
          @input="setStart"
        >
        <input
          type="range"
          min="0"
          :max="STEPS"
          :value="end"
          :class="thumb"
          :aria-label="`Maximum ${filter.label}`"
          @input="setEnd"
        >
      </div>

      <div class="flex justify-between text-muted-foreground">
        <span>{{ filter.format(range.min ?? extent.min) }}</span>
        <span>{{ filter.format(range.max ?? extent.max) }}</span>
      </div>
    </template>
  </div>
</template>
