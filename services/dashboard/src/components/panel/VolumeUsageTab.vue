<script setup lang="ts">
import { HardDrive } from '@lucide/vue';
import QueryError from '@/components/QueryError.vue';
import EmptyState from '@/components/EmptyState.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { MetricWindow } from '@/gql/graphql';
import { formatBytes, parseQuantity } from '@/lib/format';
import {
  formatMetricTime,
  formatSize,
  latestValue,
  metricTicks,
  nearestTime,
  valueAt,
  windowOf,
  type MetricLine,
} from '@/lib/metrics';
import MetricHero from '@/components/metrics/MetricHero.vue';
import MetricsToolbar from '@/components/metrics/MetricsToolbar.vue';

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

const emit = defineEmits<{
  (e: 'grow'): void;
}>();

const selectedWindow = ref<MetricWindow>(MetricWindow.Last_24H);
const cursor = ref<number | null>(null);

const { result: metricsResult, loading: metricsLoading, error: metricsError, refetch: refetchMetrics } = useQuery(
  VolumeMetricsDocument,
  () => ({ id: props.volumeId, range: { window: selectedWindow.value } }),
  { pollInterval: 30000, fetchPolicy: 'cache-and-network' },
);

const period = computed(() => windowOf(selectedWindow.value));
const short = computed(() => period.value.ms <= 24 * 3_600_000);

const range = computed(() => {
  void metricsResult.value;
  const to = Date.now();
  return { from: to - period.value.ms, to };
});

const ticks = computed(() => metricTicks(range.value.from, range.value.to, short.value));

const capacity = computed(() => parseQuantity(props.volume.size));

const points = computed(() => metricsResult.value?.volume.metrics[0]?.points ?? []);

const lines = computed<MetricLine[]>(() => [{ key: 'used', label: '', color: 'var(--primary)', points: points.value }]);

const hasData = computed(() => points.value.some(p => p.value != null));

const focusTime = computed(() => (cursor.value == null ? null : nearestTime(points.value, cursor.value)));
const focusLabel = computed(() => (focusTime.value == null ? null : formatMetricTime(focusTime.value, short.value)));

const used = computed(() =>
  focusTime.value == null ? latestValue(points.value) : valueAt(points.value, focusTime.value),
);

const readings = computed(() => [
  { key: 'used', color: 'var(--primary)', text: used.value == null ? '–' : formatBytes(used.value) },
]);

const unit = computed(() => {
  if (!capacity.value) return null;
  const percent = used.value == null ? null : Math.min(100, Math.round((used.value / capacity.value) * 100));
  return `of ${formatSize(capacity.value)}${percent == null ? '' : ` · ${percent}%`}`;
});
</script>

<template>
  <div class="@container space-y-4">
    <MetricsToolbar v-model:window="selectedWindow" />

    <section v-if="hasData" class="overflow-hidden rounded-lg border">
      <MetricHero
        v-model:cursor="cursor"
        label="Used"
        :readings="readings"
        :unit="unit"
        :split="false"
        :focus-label="focusLabel"
        :lines="lines"
        :from="range.from"
        :to="range.to"
        :max="capacity || null"
        :ticks="ticks"
        edit-label="Grow volume"
        @edit="emit('grow')"
      />
    </section>

    <div v-else-if="metricsLoading" class="space-y-3">
      <Skeleton class="h-40 w-full" />
      <Skeleton class="h-40 w-full" />
    </div>
    <QueryError
      v-else-if="metricsError"
      title="Couldn't load metrics"
      :error="metricsError"
      @retry="refetchMetrics()"
    />
    <EmptyState
      v-else
      size="compact"
      :icon="HardDrive"
      title="No usage data yet"
      :description="volume.mount ? 'Usage appears within a minute or two.' : 'Mount this volume to a service to start collecting usage.'"
    />
  </div>
</template>
