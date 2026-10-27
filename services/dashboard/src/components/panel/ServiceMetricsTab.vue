<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import EmptyState from '@/components/EmptyState.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { MetricWindow, MetricGrouping, ResourceMetric } from '@/gql/graphql';
import { formatBytes, parseCpu, parseStorageSize } from '@/lib/utils';
import {
  formatCoreLimit,
  formatCores,
  formatMetricTime,
  formatSize,
  latestValue,
  lineColors,
  metricTicks,
  nearestTime,
  valueAt,
  windowOf,
  type MetricLine,
  type MetricMarker,
  type MetricPoint,
  type MetricReading,
} from '@/lib/metrics';
import MetricHero from '@/components/metrics/MetricHero.vue';
import MetricsToolbar from '@/components/metrics/MetricsToolbar.vue';

const ServiceMetricsDocument = graphql(`
  query ServiceMetrics($id: ServiceID!, $range: MetricsRange!, $grouping: MetricGrouping!) {
    service(id: $id) {
      id
      replicas {
        desired
      }
      resources {
        cpu
        memory
      }
      deployments {
        createdAt
      }
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

const emit = defineEmits<{
  (e: 'editResources'): void;
}>();

const selectedWindow = ref<MetricWindow>(MetricWindow.Last_24H);
const perReplica = ref(false);
const cursor = ref<number | null>(null);

const grouping = computed(() => (perReplica.value ? MetricGrouping.PerReplica : MetricGrouping.Total));

const { result, loading, error, refetch } = useQuery(
  ServiceMetricsDocument,
  () => ({ id: props.serviceId, range: { window: selectedWindow.value }, grouping: grouping.value }),
  { pollInterval: 30000, fetchPolicy: 'cache-and-network' },
);

const service = computed(() => result.value?.service);
const period = computed(() => windowOf(selectedWindow.value));
const short = computed(() => period.value.ms <= 24 * 3_600_000);

const range = computed(() => {
  void result.value;
  const to = Date.now();
  return { from: to - period.value.ms, to };
});

const ticks = computed(() => metricTicks(range.value.from, range.value.to, short.value));

function shortReplica(replica: string): string {
  return replica.split('-').pop() ?? replica;
}

function linesFor(metric: ResourceMetric): MetricLine[] {
  return (service.value?.metrics ?? [])
    .filter(s => s.metric === metric)
    .sort((a, b) => (a.replica ?? '').localeCompare(b.replica ?? ''))
    .map((s, index) =>
      s.replica
        ? { key: s.replica, label: shortReplica(s.replica), color: lineColors[index % lineColors.length]!, points: s.points }
        : { key: 'total', label: '', color: 'var(--primary)', points: s.points },
    );
}

const allPoints = computed<MetricPoint[]>(() => (service.value?.metrics ?? []).flatMap(s => s.points));

const focusTime = computed(() => (cursor.value == null ? null : nearestTime(allPoints.value, cursor.value)));
const focusLabel = computed(() => (focusTime.value == null ? null : formatMetricTime(focusTime.value, short.value)));

function readings(lines: MetricLine[], format: (value: number) => string): MetricReading[] {
  return lines.map(line => {
    const value = focusTime.value == null ? latestValue(line.points) : valueAt(line.points, focusTime.value);
    return { key: line.key, color: line.color, text: value == null ? '–' : format(value) };
  });
}

const replicaCount = computed(() => Math.max(1, service.value?.replicas.desired ?? 1));
const scale = computed(() => (perReplica.value ? 1 : replicaCount.value));

function hero(
  metric: ResourceMetric,
  label: string,
  limit: number,
  format: (value: number) => string,
  formatLimit: (value: number) => string,
) {
  const lines = linesFor(metric);
  const max = limit > 0 ? limit * scale.value : null;
  const unit = max == null ? null : perReplica.value ? `of ${formatLimit(limit)} each` : `of ${formatLimit(max)}`;
  return { metric, label, lines, readings: readings(lines, format), unit, max };
}

const heroes = computed(() => [
  hero(ResourceMetric.CpuUsage, 'CPU', parseCpu(service.value?.resources.cpu ?? ''), formatCores, formatCoreLimit),
  hero(
    ResourceMetric.MemoryUsage,
    'Memory',
    parseStorageSize(service.value?.resources.memory ?? ''),
    formatBytes,
    formatSize,
  ),
]);

const markers = computed<MetricMarker[]>(() =>
  (service.value?.deployments ?? []).map(d => {
    const time = new Date(d.createdAt).getTime();
    return { time, label: `Deployed ${formatMetricTime(time, short.value)}` };
  }),
);

const legend = computed(() => (perReplica.value ? linesFor(ResourceMetric.CpuUsage).filter(l => l.label) : []));

const hasData = computed(() => allPoints.value.some(p => p.value != null));
</script>

<template>
  <div class="@container space-y-4">
    <MetricsToolbar
      v-model:split="perReplica"
      v-model:window="selectedWindow"
      split-label="Per replica"
      :legend="legend"
    />

    <section v-if="hasData" class="grid gap-px overflow-hidden rounded-lg border bg-border @lg:grid-cols-2">
      <MetricHero
        v-for="item in heroes"
        :key="item.metric"
        v-model:cursor="cursor"
        :label="item.label"
        :readings="item.readings"
        :unit="item.unit"
        :split="perReplica"
        :focus-label="focusLabel"
        :lines="item.lines"
        :from="range.from"
        :to="range.to"
        :max="item.max"
        :ticks="ticks"
        :markers="markers"
        @edit="emit('editResources')"
      />
    </section>

    <div v-else-if="loading" class="space-y-3">
      <Skeleton class="h-40 w-full" />
      <Skeleton class="h-40 w-full" />
    </div>
    <QueryError
      v-else-if="error"
      title="Couldn't load metrics"
      :error="error"
      @retry="refetch()"
    />
    <EmptyState
      v-else
      size="compact"
      title="No metrics yet"
      description="They appear within a minute or two of the service running."
    />
  </div>
</template>
