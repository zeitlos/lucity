<script setup lang="ts">
import { computed } from 'vue';
import type { MetricWindow } from '@/gql/graphql';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { metricWindows } from '@/lib/metrics';

const props = withDefaults(
  defineProps<{
    splitLabel?: string;
    legend?: { key: string; label: string; color: string }[];
  }>(),
  { splitLabel: undefined, legend: () => [] },
);

const split = defineModel<boolean>('split', { default: false });
const selected = defineModel<MetricWindow>('window', { required: true });

const scope = computed({
  get: () => (split.value ? 'split' : 'total'),
  set: (value: 'total' | 'split') => {
    split.value = value === 'split';
  },
});

const scopeOptions = computed(() => [
  { value: 'total' as const, label: 'Total' },
  { value: 'split' as const, label: props.splitLabel ?? '' },
]);
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="flex flex-wrap items-center gap-4">
      <SegmentedControl v-if="splitLabel" v-model="scope" label="Scope" :options="scopeOptions" />

      <div v-if="legend.length" class="flex flex-wrap items-center gap-x-3.5 gap-y-1 text-xs text-muted-foreground">
        <span v-for="item in legend" :key="item.key" class="inline-flex items-center gap-1.5">
          <span class="size-1.5 rounded-full" :style="{ background: item.color }" />
          {{ item.label }}
        </span>
      </div>
    </div>

    <SegmentedControl v-model="selected" label="Time range" :options="metricWindows" />
  </div>
</template>
