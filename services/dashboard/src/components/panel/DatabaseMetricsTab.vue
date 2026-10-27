<script setup lang="ts">
import CopyButton from '@/components/CopyButton.vue';
import QueryError from '@/components/QueryError.vue';
import EmptyState from '@/components/EmptyState.vue';
import { Skeleton } from '@/components/ui/skeleton';
import IconButton from '@/components/IconButton.vue';
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ChevronDown, ChevronUp } from '@lucide/vue';
import { graphql } from '@/gql';
import { DatabaseInstanceRole, DatabaseMetric, MetricGrouping, MetricWindow } from '@/gql/graphql';
import { formatBytes, parseCpu, parseStorageSize, cn } from '@/lib/utils';
import {
  formatCoreLimit,
  formatCores,
  formatCount,
  formatMetricTime,
  formatMilliseconds,
  formatRate,
  formatRatio,
  formatSeconds,
  formatSize,
  latestValue,
  lineColors,
  metricTicks,
  nearestTime,
  timeOf,
  valueAt,
  windowOf,
  type MetricLine,
  type MetricPoint,
  type MetricReading,
} from '@/lib/metrics';
import MetricHero from '@/components/metrics/MetricHero.vue';
import MetricLineChart from '@/components/metrics/MetricLineChart.vue';
import MetricsToolbar from '@/components/metrics/MetricsToolbar.vue';

const DatabaseMetricsDocument = graphql(`
  query DatabaseMetrics($id: DatabaseID!, $range: MetricsRange!, $grouping: MetricGrouping!) {
    database(id: $id) {
      id
      instances
      size
      resources {
        cpu
        memory
      }
      connectionLimit
      metrics(
        metrics: [
          CPU_USAGE
          MEMORY_USAGE
          STORAGE_USED
          CONNECTIONS
          TRANSACTIONS
          CACHE_HIT_RATIO
          ROWS_READ
          ROWS_WRITTEN
          DEADLOCKS
          REPLICATION_LAG
        ]
        range: $range
        grouping: $grouping
      ) {
        metric
        instance
        role
        points {
          timestamp
          value
        }
      }
    }
  }
`);

const DatabaseQueryStatisticsDocument = graphql(`
  query DatabaseQueryStatistics($id: DatabaseID!) {
    database(id: $id) {
      id
      queryStatistics {
        enabled
        since
        queries {
          query
          calls
          rows
          totalMilliseconds
          meanMilliseconds
          share
        }
      }
    }
  }
`);

const props = defineProps<{
  databaseId: string;
}>();

const emit = defineEmits<{
  (e: 'editResources'): void;
}>();

const selectedWindow = ref<MetricWindow>(MetricWindow.Last_24H);
const perInstance = ref(false);
const cursor = ref<number | null>(null);
const openQuery = ref<number | null>(null);

const grouping = computed(() => (perInstance.value ? MetricGrouping.PerReplica : MetricGrouping.Total));

const { result, loading, error, refetch } = useQuery(
  DatabaseMetricsDocument,
  () => ({ id: props.databaseId, range: { window: selectedWindow.value }, grouping: grouping.value }),
  { pollInterval: 30000, fetchPolicy: 'cache-and-network' },
);

const { result: statisticsResult, loading: statisticsLoading, error: statisticsError } = useQuery(
  DatabaseQueryStatisticsDocument,
  () => ({ id: props.databaseId }),
  { pollInterval: 60000, fetchPolicy: 'cache-and-network' },
);

const database = computed(() => result.value?.database);
const period = computed(() => windowOf(selectedWindow.value));
const short = computed(() => period.value.ms <= 24 * 3_600_000);

const range = computed(() => {
  void result.value;
  const to = Date.now();
  return { from: to - period.value.ms, to };
});

const ticks = computed(() => metricTicks(range.value.from, range.value.to, short.value));

type Line = MetricLine & { role: DatabaseInstanceRole | null };

function instanceNumber(instance: string): string {
  return instance.split('-').pop() ?? instance;
}

function linesFor(metric: DatabaseMetric): Line[] {
  const series = (database.value?.metrics ?? [])
    .filter(s => s.metric === metric)
    .sort((a, b) => {
      const rank = (role?: DatabaseInstanceRole | null) => (role === DatabaseInstanceRole.Primary ? 0 : 1);
      return rank(a.role) - rank(b.role) || (a.instance ?? '').localeCompare(b.instance ?? '');
    });

  let replicas = 0;
  return series.map(s => {
    if (!s.instance) {
      return { key: 'total', label: '', role: null, color: 'var(--primary)', points: s.points };
    }
    const primary = s.role === DatabaseInstanceRole.Primary;
    return {
      key: s.instance,
      label: `${instanceNumber(s.instance)} · ${primary ? 'primary' : 'replica'}`,
      role: s.role ?? null,
      color: primary ? lineColors[0]! : lineColors[1 + (replicas++ % (lineColors.length - 1))]!,
      points: s.points,
    };
  });
}

const allPoints = computed<MetricPoint[]>(() => (database.value?.metrics ?? []).flatMap(s => s.points));

const focusTime = computed(() => (cursor.value == null ? null : nearestTime(allPoints.value, cursor.value)));
const focusLabel = computed(() => (focusTime.value == null ? null : formatMetricTime(focusTime.value, short.value)));

function shown(points: MetricPoint[]): number | null {
  return focusTime.value == null ? latestValue(points) : valueAt(points, focusTime.value);
}

function samples(lines: Line[]): { time: number; value: number }[] {
  return lines.flatMap(l =>
    l.points.filter(p => p.value != null).map(p => ({ time: timeOf(p), value: p.value as number })),
  );
}

function formatTime(time: number): string {
  return formatMetricTime(time, short.value);
}

const instanceCount = computed(() => Math.max(1, database.value?.instances ?? 1));
const scale = computed(() => (perInstance.value ? 1 : instanceCount.value));

function readings(lines: Line[], format: (value: number) => string): MetricReading[] {
  return lines.map(l => {
    const value = shown(l.points);
    return { key: l.key, color: l.color, text: value == null ? '–' : format(value) };
  });
}

function hero(
  metric: DatabaseMetric,
  label: string,
  limit: number,
  format: (value: number) => string,
  formatLimit: (value: number) => string,
) {
  const lines = linesFor(metric);
  const max = limit > 0 ? limit * scale.value : null;
  const unit = max == null ? null : perInstance.value ? `of ${formatLimit(limit)} each` : `of ${formatLimit(max)}`;
  return { metric, label, lines, readings: readings(lines, format), unit, max };
}

const heroes = computed(() => [
  hero(DatabaseMetric.CpuUsage, 'CPU', parseCpu(database.value?.resources.cpu ?? ''), formatCores, formatCoreLimit),
  hero(
    DatabaseMetric.MemoryUsage,
    'Memory',
    parseStorageSize(database.value?.resources.memory ?? ''),
    formatBytes,
    formatSize,
  ),
]);

type Cell = {
  metric: DatabaseMetric;
  label: string;
  lines: Line[];
  readings: MetricReading[];
  unit: string;
  context: string;
  min: number;
  max: number | null;
  area: boolean;
  fills: { key: string; color: string; percent: number }[] | null;
};

function peakText(lines: Line[], format: (value: number) => string, suffix = ''): string {
  const values = samples(lines).map(s => s.value);
  if (!values.length) return '';
  const peak = `peak ${format(Math.max(...values))}${suffix}`;
  if (perInstance.value) return peak;
  const average = values.reduce((sum, v) => sum + v, 0) / values.length;
  return `${peak} · avg ${format(average)}${suffix}`;
}

function extremeText(lines: Line[], word: string, pick: 'max' | 'min', format: (value: number) => string): string {
  const all = samples(lines);
  if (!all.length) return '';
  const extreme = all.reduce((best, s) => ((pick === 'max' ? s.value > best.value : s.value < best.value) ? s : best));
  return `${word} ${format(extreme.value)} at ${formatTime(extreme.time)}`;
}

function cell(
  metric: DatabaseMetric,
  label: string,
  format: (value: number) => string,
  unit: string,
  context: (lines: Line[]) => string,
): Cell {
  const lines = linesFor(metric);
  return {
    metric,
    label,
    lines,
    readings: readings(lines, format),
    unit,
    context: focusLabel.value ? `at ${focusLabel.value}` : context(lines),
    min: 0,
    max: null,
    area: true,
    fills: null,
  };
}

const storageCell = computed<Cell>(() => {
  const lines = linesFor(DatabaseMetric.StorageUsed);
  const size = parseStorageSize(database.value?.size ?? '');
  const capacity = size > 0 ? size * scale.value : 0;

  let context = '';
  if (focusLabel.value) {
    context = `at ${focusLabel.value}`;
  } else if (capacity > 0 && lines.length) {
    const used = lines.reduce((sum, l) => sum + (latestValue(l.points) ?? 0), 0);
    const first = lines.reduce((sum, l) => sum + (l.points.find(p => p.value != null)?.value ?? 0), 0);
    const growth = used - first;
    const percent = Math.round((used / (size * instanceCount.value)) * 100);
    context = `${percent}% · ${growth > 0 ? `+${formatBytes(growth)}` : 'no growth'} in ${period.value.label}`;
  }

  return {
    metric: DatabaseMetric.StorageUsed,
    label: 'Storage',
    lines,
    readings: readings(lines, formatBytes),
    unit: capacity > 0 ? (perInstance.value ? `of ${formatSize(size)} each` : `of ${formatSize(capacity)}`) : '',
    context,
    min: 0,
    max: null,
    area: false,
    fills:
      capacity > 0
        ? lines.map(l => ({
            key: l.key,
            color: l.color,
            percent: Math.min(100, ((shown(l.points) ?? 0) / capacity) * 100),
          }))
        : [],
  };
});

const deadlocksCell = computed<Cell>(() => {
  const lines = linesFor(DatabaseMetric.Deadlocks);
  const total = (l: Line) => l.points.reduce((sum, p) => sum + (p.value ?? 0), 0);
  const last = samples(lines)
    .filter(s => s.value >= 0.5)
    .reduce<number | null>((latestTime, s) => (latestTime == null || s.time > latestTime ? s.time : latestTime), null);

  return {
    metric: DatabaseMetric.Deadlocks,
    label: 'Deadlocks',
    lines,
    readings: lines.map(l => {
      const value = focusTime.value == null ? total(l) : valueAt(l.points, focusTime.value);
      return { key: l.key, color: l.color, text: value == null ? '–' : formatCount(value) };
    }),
    unit: focusTime.value == null ? `in ${period.value.label}` : '',
    context: focusLabel.value ? `at ${focusLabel.value}` : last == null ? 'none' : `last at ${formatTime(last)}`,
    min: 0,
    max: null,
    area: true,
    fills: null,
  };
});

const cells = computed<Cell[]>(() => {
  const connectionLimit = database.value?.connectionLimit ?? 0;
  const connectionsUnit =
    connectionLimit > 0
      ? perInstance.value
        ? `of ${formatCount(connectionLimit)} each`
        : `of ${formatCount(connectionLimit * instanceCount.value)}`
      : '';

  const cacheLines = linesFor(DatabaseMetric.CacheHitRatio);
  const cacheValues = samples(cacheLines).map(s => s.value);
  const cacheMin = Math.max(0, Math.min(0.98, (cacheValues.length ? Math.min(...cacheValues) : 1) - 0.005));

  const list: Cell[] = [
    storageCell.value,
    cell(DatabaseMetric.Connections, 'Connections', formatCount, connectionsUnit, lines => peakText(lines, formatCount)),
    cell(DatabaseMetric.Transactions, 'Transactions', formatRate, '/s', lines => peakText(lines, formatRate, '/s')),
    {
      ...cell(DatabaseMetric.CacheHitRatio, 'Cache hit ratio', formatRatio, '', lines =>
        extremeText(lines, 'lowest', 'min', formatRatio),
      ),
      min: cacheMin,
      max: 1,
      area: false,
    },
    cell(DatabaseMetric.RowsRead, 'Rows read', formatRate, '/s', lines => peakText(lines, formatRate, '/s')),
    cell(DatabaseMetric.RowsWritten, 'Rows written', formatRate, '/s', lines => peakText(lines, formatRate, '/s')),
    deadlocksCell.value,
  ];

  if (instanceCount.value > 1) {
    list.push(
      cell(DatabaseMetric.ReplicationLag, 'Replication lag', formatSeconds, '', lines => {
        const lag = samples(lines).map(s => s.value);
        if (lag.length && Math.max(...lag) < 0.0005) return 'no lag';
        return extremeText(lines, 'peak', 'max', formatSeconds);
      }),
    );
  }

  return list;
});

const legend = computed(() => (perInstance.value ? linesFor(DatabaseMetric.CpuUsage).filter(l => l.role) : []));

const hasData = computed(() => allPoints.value.some(p => p.value != null));

const statistics = computed(() => statisticsResult.value?.database.queryStatistics);

function plural(count: number, word: string): string {
  return `${count} ${word}${count === 1 ? '' : 's'}`;
}

const sinceRelative = computed(() => {
  const value = statistics.value?.since;
  if (!value) return null;
  const minutes = (Date.now() - new Date(value).getTime()) / 60_000;
  if (minutes < 60) return `last ${plural(Math.max(1, Math.round(minutes)), 'minute')}`;
  const hours = minutes / 60;
  if (hours < 48) return `last ${plural(Math.round(hours), 'hour')}`;
  return `last ${plural(Math.round(hours / 24), 'day')}`;
});

const sinceExact = computed(() => {
  const value = statistics.value?.since;
  if (!value) return undefined;
  return `Counted since ${new Date(value).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })}`;
});

function rowsPerCall(rows: number, calls: number): string {
  if (calls === 0) return '0';
  const value = rows / calls;
  return value < 10 ? String(Number(value.toFixed(1))) : formatCount(value);
}

function toggleQuery(index: number) {
  openQuery.value = openQuery.value === index ? null : index;
}

</script>

<template>
  <div class="@container space-y-4">
    <MetricsToolbar
      v-model:split="perInstance"
      v-model:window="selectedWindow"
      split-label="Per instance"
      :legend="legend"
    />

    <template v-if="hasData">
      <section class="grid gap-px overflow-hidden rounded-lg border bg-border @lg:grid-cols-2">
        <MetricHero
          v-for="item in heroes"
          :key="item.metric"
          v-model:cursor="cursor"
          :label="item.label"
          :readings="item.readings"
          :unit="item.unit"
          :split="perInstance"
          :focus-label="focusLabel"
          :lines="item.lines"
          :from="range.from"
          :to="range.to"
          :max="item.max"
          :ticks="ticks"
          @edit="emit('editResources')"
        />
      </section>

      <section class="grid grid-cols-2 gap-px overflow-hidden rounded-lg border bg-border @xl:grid-cols-4">
        <div
          v-for="item in cells"
          :key="item.metric"
          class="flex min-w-0 flex-col gap-1.5 bg-card px-3.5 pb-3 pt-3.5"
        >
          <div class="text-xs font-medium text-muted-foreground">{{ item.label }}</div>
          <div class="flex items-baseline gap-1 whitespace-nowrap">
            <span class="inline-flex items-baseline gap-2.5">
              <span v-for="reading in item.readings" :key="reading.key" class="inline-flex items-baseline gap-1.5">
                <span
                  v-if="perInstance"
                  class="size-[7px] -translate-y-0.5 rounded-full"
                  :style="{ background: reading.color }"
                />
                <span
                  :class="cn(
                    'font-semibold tabular-nums tracking-tight',
                    perInstance ? 'text-[17px]' : 'text-xl',
                  )"
                >{{ reading.text }}</span>
              </span>
            </span>
            <span v-if="item.unit" class="truncate text-xs text-muted-foreground">{{ item.unit }}</span>
          </div>

          <div v-if="item.fills" class="flex h-8 flex-col justify-center gap-1.5">
            <div
              v-for="fill in item.fills"
              :key="fill.key"
              :class="cn('overflow-hidden rounded-full bg-muted', item.fills.length > 1 ? 'h-1' : 'h-1.5')"
            >
              <div class="h-full rounded-full" :style="{ width: `${fill.percent}%`, background: fill.color }" />
            </div>
          </div>
          <MetricLineChart
            v-else
            v-model:cursor="cursor"
            :series="item.lines"
            :from="range.from"
            :to="range.to"
            :min="item.min"
            :max="item.max"
            :area="item.area && item.lines.length === 1"
          />

          <div
            :class="cn(
              'truncate text-[11px] tabular-nums',
              focusLabel ? 'font-medium text-foreground' : 'text-muted-foreground',
            )"
          >
            {{ item.context || ' ' }}
          </div>
        </div>
        <div v-if="cells.length % 2" class="bg-card" />
      </section>
    </template>

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
      description="They appear within a minute or two of the database running."
    />

    <div class="-mx-4 overflow-x-auto">
      <table class="w-full border-collapse text-[13px]">
        <thead>
          <tr class="border-t bg-muted/60 text-left">
            <th class="whitespace-nowrap py-2 pl-4 pr-3 align-baseline text-[13px] font-semibold">
              <span :title="sinceExact">
                Top queries
                <span v-if="sinceRelative" class="text-xs font-normal text-muted-foreground">· {{ sinceRelative }}</span>
              </span>
            </th>
            <th class="px-3 py-2 text-right align-baseline text-[11px] font-medium text-muted-foreground">Time</th>
            <th class="px-3 py-2 text-right align-baseline text-[11px] font-medium text-muted-foreground">Calls</th>
            <th class="px-3 py-2 text-right align-baseline text-[11px] font-medium text-muted-foreground">Mean</th>
            <th class="py-2 pl-3 pr-1 text-right align-baseline text-[11px] font-medium text-muted-foreground">Rows</th>
            <th class="w-10 pr-3"><span class="sr-only">Details</span></th>
          </tr>
        </thead>
        <tbody v-if="statistics?.enabled && statistics.queries.length">
          <template v-for="(statement, index) in statistics.queries" :key="index">
            <tr :class="cn('border-t', openQuery === index && 'bg-muted/40')">
              <td class="w-full max-w-0 py-2.5 pl-4 pr-3">
                <code class="block truncate font-mono text-xs">{{ statement.query }}</code>
              </td>
              <td class="whitespace-nowrap px-3 py-2.5 tabular-nums">
                <div class="flex items-center justify-end gap-2">
                  <div class="h-1 w-12 overflow-hidden rounded-full bg-muted">
                    <div class="h-full rounded-full bg-primary" :style="{ width: `${statement.share * 100}%` }" />
                  </div>
                  <span class="min-w-10 text-right">{{ formatRatio(statement.share) }}</span>
                </div>
              </td>
              <td class="whitespace-nowrap px-3 py-2.5 text-right tabular-nums">{{ formatCount(statement.calls) }}</td>
              <td class="whitespace-nowrap px-3 py-2.5 text-right tabular-nums">
                {{ formatMilliseconds(statement.meanMilliseconds) }}
              </td>
              <td class="whitespace-nowrap py-2.5 pl-3 pr-1 text-right tabular-nums text-muted-foreground">
                {{ formatCount(statement.rows) }}
              </td>
              <td class="py-1.5 pr-3 text-right">
                <IconButton
                  :label="openQuery === index ? 'Hide details' : 'Show details'"
                  size="icon-xs"
                  class="text-muted-foreground"
                  :aria-expanded="openQuery === index"
                  @click="toggleQuery(index)"
                >
                  <ChevronUp v-if="openQuery === index" />
                  <ChevronDown v-else />
                </IconButton>
              </td>
            </tr>
            <tr v-if="openQuery === index" class="bg-muted/40">
              <td colspan="6" class="px-4 pb-3.5">
                <code
                  class="block whitespace-pre-wrap break-words rounded-md border bg-card px-3 py-2.5 font-mono text-xs leading-relaxed"
                >{{ statement.query }}</code>
                <div class="mt-2.5 flex items-center gap-4 text-[11px] tabular-nums text-muted-foreground">
                  <span>{{ formatMilliseconds(statement.totalMilliseconds) }} in total</span>
                  <span>{{ rowsPerCall(statement.rows, statement.calls) }} rows per call</span>
                  <CopyButton label="Copy query" class="ml-auto" :value="statement.query" />
                </div>
              </td>
            </tr>
          </template>
        </tbody>
        <tbody v-else>
          <tr class="border-t">
            <td colspan="6" class="px-4 py-6 text-center text-sm text-muted-foreground">
              <span v-if="statisticsLoading && !statistics">Loading queries…</span>
              <span v-else-if="statisticsError">Couldn't read query statistics from the database.</span>
              <span v-else-if="statistics && !statistics.enabled">This database doesn't collect query statistics yet.</span>
              <span v-else>No queries recorded yet.</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
