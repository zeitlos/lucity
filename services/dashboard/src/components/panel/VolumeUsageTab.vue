<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { MetricWindow } from '@/gql/graphql';
import { formatBytes, parseStorageSize, cn } from '@/lib/utils';
import MetricAreaChart from '@/components/metrics/MetricAreaChart.vue';

const VolumeMetricsDocument = graphql(`
  query VolumeMetrics($id: VolumeID!, $range: MetricsRange!) {
    volume(id: $id) {
      id
      metrics(metrics: [STORAGE_USED], range: $range) {
        metric
        points {
          timestamp
          value
        }
      }
    }
  }
`);

const props = defineProps<{
  volumeId: string;
  volume: {
    name: string;
    size: string;
    mount?: { service: string; path: string } | null;
  };
}>();

const windowOptions: { value: MetricWindow; label: string }[] = [
  { value: MetricWindow.Last_1H, label: '1h' },
  { value: MetricWindow.Last_6H, label: '6h' },
  { value: MetricWindow.Last_24H, label: '24h' },
  { value: MetricWindow.Last_7D, label: '7d' },
  { value: MetricWindow.Last_30D, label: '30d' },
];

const selectedWindow = ref<MetricWindow>(MetricWindow.Last_24H);

const windowMs: Record<MetricWindow, number> = {
  [MetricWindow.Last_1H]: 3_600_000,
  [MetricWindow.Last_6H]: 6 * 3_600_000,
  [MetricWindow.Last_24H]: 24 * 3_600_000,
  [MetricWindow.Last_7D]: 7 * 24 * 3_600_000,
  [MetricWindow.Last_30D]: 30 * 24 * 3_600_000,
};

const { result: metricsResult, loading: metricsLoading } = useQuery(
  VolumeMetricsDocument,
  () => ({ id: props.volumeId, range: { window: selectedWindow.value } }),
  { pollInterval: 30000, fetchPolicy: 'cache-and-network' },
);

const capacityBytes = computed(() => parseStorageSize(props.volume.size));

const points = computed(() => metricsResult.value?.volume.metrics[0]?.points ?? []);

const range = computed(() => {
  void metricsResult.value;
  const to = Date.now();
  return { from: to - windowMs[selectedWindow.value], to };
});

const hasData = computed(() => points.value.some(p => p.value != null));

const currentUsage = computed(() => {
  const withValues = points.value.filter(p => p.value != null);
  const last = withValues[withValues.length - 1];
  return last?.value ?? null;
});

const usagePercent = computed(() => {
  if (currentUsage.value == null || !capacityBytes.value) return null;
  return Math.min(100, Math.round((currentUsage.value / capacityBytes.value) * 100));
});
</script>

<template>
  <div class="space-y-6">
    <!-- Usage -->
    <section class="space-y-2">
      <div class="flex items-center justify-between px-1">
        <h3 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Usage
        </h3>
        <div class="inline-flex overflow-hidden rounded-md border text-xs">
          <button
            v-for="option in windowOptions"
            :key="option.value"
            type="button"
            :class="cn(
              'px-2.5 py-1 text-muted-foreground transition-colors not-first:border-l hover:text-foreground',
              selectedWindow === option.value && 'bg-muted font-medium text-foreground',
            )"
            @click="selectedWindow = option.value"
          >
            {{ option.label }}
          </button>
        </div>
      </div>

      <div class="rounded-lg border p-4">
        <div class="mb-3 flex items-baseline gap-2">
          <span class="text-2xl font-semibold text-foreground">
            {{ currentUsage != null ? formatBytes(currentUsage) : '—' }}
          </span>
          <span class="text-sm text-muted-foreground">
            / {{ formatBytes(capacityBytes) }}
            <template v-if="usagePercent != null">({{ usagePercent }}%)</template>
          </span>
        </div>

        <MetricAreaChart
          v-if="hasData"
          :points="points"
          :from="range.from"
          :to="range.to"
          :format-value="formatBytes"
        />

        <div
          v-else
          class="flex h-[180px] items-center justify-center text-center text-sm text-muted-foreground"
        >
          <span v-if="metricsLoading">Loading usage…</span>
          <span v-else-if="!volume.mount">
            Mount this volume to a service to start collecting usage.
          </span>
          <span v-else>No usage data yet. Metrics appear within a minute or two.</span>
        </div>
      </div>
    </section>
  </div>
</template>
