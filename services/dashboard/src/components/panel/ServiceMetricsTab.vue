<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { MetricWindow, MetricGrouping, ResourceMetric } from '@/gql/graphql';
import { formatBytes, cn } from '@/lib/utils';
import MetricAreaChart from '@/components/metrics/MetricAreaChart.vue';

const ServiceMetricsDocument = graphql(`
  query ServiceMetrics($id: ServiceID!, $range: MetricsRange!, $grouping: MetricGrouping!) {
    service(id: $id) {
      id
      metrics(metrics: [CPU_USAGE, MEMORY_USAGE], range: $range, grouping: $grouping) {
        metric
        replica
        points {
          timestamp
          value
        }
      }
    }
  }
`);

const props = defineProps<{
  serviceId: string;
}>();

const windowOptions: { value: MetricWindow; label: string }[] = [
  { value: MetricWindow.Last_1H, label: '1h' },
  { value: MetricWindow.Last_6H, label: '6h' },
  { value: MetricWindow.Last_24H, label: '24h' },
  { value: MetricWindow.Last_7D, label: '7d' },
  { value: MetricWindow.Last_30D, label: '30d' },
];

const windowMs: Record<MetricWindow, number> = {
  [MetricWindow.Last_1H]: 3_600_000,
  [MetricWindow.Last_6H]: 6 * 3_600_000,
  [MetricWindow.Last_24H]: 24 * 3_600_000,
  [MetricWindow.Last_7D]: 7 * 24 * 3_600_000,
  [MetricWindow.Last_30D]: 30 * 24 * 3_600_000,
};

const selectedWindow = ref<MetricWindow>(MetricWindow.Last_24H);
const perReplica = ref(false);

const grouping = computed(() => (perReplica.value ? MetricGrouping.PerReplica : MetricGrouping.Total));

const { result, loading } = useQuery(
  ServiceMetricsDocument,
  () => ({ id: props.serviceId, range: { window: selectedWindow.value }, grouping: grouping.value }),
  { pollInterval: 30000, fetchPolicy: 'cache-and-network' },
);

const range = computed(() => {
  void result.value;
  const to = Date.now();
  return { from: to - windowMs[selectedWindow.value], to };
});

function shortReplica(replica?: string | null): string | undefined {
  if (!replica) return undefined;
  const parts = replica.split('-');
  return parts[parts.length - 1];
}

function seriesFor(metric: ResourceMetric) {
  const all = result.value?.service.metrics ?? [];
  return all
    .filter(s => s.metric === metric)
    .map(s => ({ label: shortReplica(s.replica), points: s.points }));
}

const cpuSeries = computed(() => seriesFor(ResourceMetric.CpuUsage));
const memorySeries = computed(() => seriesFor(ResourceMetric.MemoryUsage));

function hasData(series: { points: { value?: number | null }[] }[]): boolean {
  return series.some(s => s.points.some(p => p.value != null));
}

const hasAny = computed(() => hasData(cpuSeries.value) || hasData(memorySeries.value));

function formatCpu(value: number): string {
  if (value < 1) return `${Math.round(value * 1000)}m`;
  return value.toFixed(2);
}
</script>

<template>
  <div class="space-y-5">
    <div class="flex items-center justify-between gap-3 px-1">
      <div class="inline-flex overflow-hidden rounded-md border text-xs">
        <button
          type="button"
          :class="cn(
            'px-2.5 py-1 text-muted-foreground transition-colors hover:text-foreground',
            !perReplica && 'bg-muted font-medium text-foreground',
          )"
          @click="perReplica = false"
        >
          Total
        </button>
        <button
          type="button"
          :class="cn(
            'border-l px-2.5 py-1 text-muted-foreground transition-colors hover:text-foreground',
            perReplica && 'bg-muted font-medium text-foreground',
          )"
          @click="perReplica = true"
        >
          Per replica
        </button>
      </div>

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

    <template v-if="hasAny">
      <section class="space-y-2">
        <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          CPU
        </h3>
        <div class="rounded-lg border p-4">
          <MetricAreaChart
            :series="cpuSeries"
            :from="range.from"
            :to="range.to"
            :format-value="formatCpu"
          />
        </div>
      </section>

      <section class="space-y-2">
        <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Memory
        </h3>
        <div class="rounded-lg border p-4">
          <MetricAreaChart
            :series="memorySeries"
            :from="range.from"
            :to="range.to"
            :format-value="formatBytes"
          />
        </div>
      </section>
    </template>

    <div
      v-else
      class="flex h-[180px] items-center justify-center text-center text-sm text-muted-foreground"
    >
      <span v-if="loading">Loading metrics…</span>
      <span v-else>No metrics yet. They appear within a minute or two of the service running.</span>
    </div>
  </div>
</template>
