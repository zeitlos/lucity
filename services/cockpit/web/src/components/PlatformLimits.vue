<script setup lang="ts">
import { computed } from 'vue';

import { LimitUnit, type LimitsQuery } from '@/gql/graphql';
import { chartPaths, count, gibibytes, percent, shortDate, tone, toneText } from '@/lib/platform';

type Group = LimitsQuery['limits'][number];
type Limit = Group['limits'][number];

const props = defineProps<{
  groups: Group[] | null;
  error: string | null;
}>();

const WIDTH = 300;
const HEIGHT = 40;
const DAY = 24 * 60 * 60 * 1000;

function present(limit: Limit) {
  const bytes = limit.unit === LimitUnit.Bytes;
  const format = (value: number) => (bytes ? gibibytes(value) : count(value));
  const unit = bytes ? ' GiB' : '';
  const ratio = limit.cap ? limit.value / limit.cap : null;
  const level = ratio == null ? 'none' : tone(ratio, 0.8);

  const end = Date.now();
  const start = end - 30 * DAY;
  const points = [...limit.history.map((point) => ({ at: Date.parse(point.time), value: point.value })), { at: end, value: limit.value }];
  const top = Math.max(limit.cap ?? 0, ...points.map((point) => point.value)) * 1.08;
  const peak = limit.history.reduce<{ time: string; value: number } | null>((best, point) => (!best || point.value > best.value ? point : best), null);

  return {
    name: limit.name,
    value: limit.cap == null ? `${format(limit.value)}${unit}` : format(limit.value),
    of: limit.cap == null ? '' : `of ${format(limit.cap)}${unit}`,
    share: ratio == null ? '' : percent(ratio),
    note: peak && peak.value > limit.value * 1.05 ? `Peak ${format(peak.value)}${unit}, ${shortDate(peak.time)}` : '',
    level,
    chart: chartPaths(points, start, end, top, WIDTH, HEIGHT),
    capY: limit.cap == null || top <= 0 ? null : HEIGHT - 1 - ((HEIGHT - 4) * limit.cap) / top,
  };
}

const columns = computed(() => (props.groups ?? []).map((group) => ({ name: group.name, rows: group.limits.map(present) })));
</script>

<template>
  <section aria-labelledby="limits-title" class="mt-4 border-t">
    <div class="flex min-h-11 items-center px-4">
      <h2 id="limits-title" class="font-semibold">Limits</h2>
      <span class="ml-auto text-muted-foreground">30 days</span>
    </div>

    <p v-if="error" class="border-t border-hairline px-4 py-3 text-status-danger">{{ error }}</p>

    <div v-else class="grid grid-cols-[repeat(auto-fit,minmax(320px,1fr))] border-t border-hairline tabular-nums">
      <template v-if="!groups">
        <div v-for="index in 3" :key="index" class="flex flex-col border-r border-hairline last:border-r-0">
          <div class="h-9 border-b border-hairline bg-sidebar" />
          <div class="flex flex-col gap-2 border-b border-hairline px-4 py-3">
            <span class="h-4 w-24 animate-pulse rounded-md bg-muted" />
            <span class="h-8 w-32 animate-pulse rounded-md bg-muted" />
            <span class="h-10 w-full animate-pulse rounded-md bg-muted" />
          </div>
        </div>
      </template>

      <div v-for="column in columns" :key="column.name" class="flex min-w-0 flex-col border-r border-hairline last:border-r-0">
        <div class="flex h-9 items-center border-b border-hairline bg-sidebar px-4 font-semibold">{{ column.name }}</div>
        <p v-if="column.rows.length === 0" class="border-b border-hairline px-4 py-3 text-muted-foreground">Nothing measured</p>
        <div v-for="row in column.rows" :key="row.name" class="flex flex-col gap-0.5 border-b border-hairline px-4 py-3">
          <span class="flex items-baseline gap-2">
            <span class="truncate text-muted-foreground">{{ row.name }}</span>
            <span class="ml-auto shrink-0 text-muted-foreground">{{ row.note }}</span>
          </span>
          <span class="flex items-baseline gap-1.5">
            <span class="text-2xl font-semibold tracking-tight" :class="toneText[row.level]">{{ row.value }}</span>
            <span class="text-muted-foreground">{{ row.of }}</span>
            <span class="ml-auto font-semibold" :class="toneText[row.level]">{{ row.share }}</span>
          </span>
          <svg
            :viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
            preserveAspectRatio="none"
            role="img"
            :aria-label="`${row.name} over the last 30 days`"
            class="mt-1 block h-10 w-full overflow-visible"
          >
            <path :d="row.chart.area" :class="row.level === 'none' ? 'fill-foreground/7' : 'fill-status-warn/12'" />
            <line
              v-if="row.capY != null"
              x1="0"
              :y1="row.capY"
              :x2="WIDTH"
              :y2="row.capY"
              stroke-dasharray="4 3"
              vector-effect="non-scaling-stroke"
              :class="row.level === 'none' ? 'stroke-foreground/35' : 'stroke-status-danger/70'"
            />
            <path
              :d="row.chart.line"
              fill="none"
              stroke-width="1.75"
              stroke-linejoin="round"
              vector-effect="non-scaling-stroke"
              :class="row.level === 'none' ? 'stroke-foreground/78' : 'stroke-status-warn'"
            />
            <line x1="0" :y1="HEIGHT - 0.5" :x2="WIDTH" :y2="HEIGHT - 0.5" vector-effect="non-scaling-stroke" class="stroke-border" />
          </svg>
        </div>
      </div>
    </div>
  </section>
</template>
