<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';

import BarChart from '@/components/BarChart.vue';
import { graphql } from '@/gql';
import {
  bucketStart,
  bucketStarts,
  endTitle,
  fullDate,
  groupings,
  periods,
  spanTitle,
  ticks,
  time,
  type Bar,
  type Grouping,
  type Layer,
  type Period,
} from '@/lib/trends';

type Chart = 'signups' | 'workspaces';

const props = defineProps<{
  loadedAt?: string;
}>();

const TrendsDocument = graphql(`
  query Trends($since: Time!) {
    signups(since: $since) { start count }
    workspaceStandings(since: $since) { at good billing abuse }
  }
`);

const signupLayers: Layer[] = [
  { label: 'Signups', fill: 'fill-foreground/55', partialFill: 'fill-foreground/20', swatch: 'bg-foreground/55' },
];

const standingLayers: Layer[] = [
  { label: 'Abuse', fill: 'fill-status-danger', partialFill: 'fill-status-danger', swatch: 'bg-status-danger' },
  { label: 'Billing', fill: 'fill-status-warn/55', partialFill: 'fill-status-warn/55', swatch: 'bg-status-warn/55' },
  { label: 'Good', fill: 'fill-foreground/15', partialFill: 'fill-foreground/15', swatch: 'bg-foreground/15' },
];

const period = ref(periods[2]!);
const grouping = ref<Grouping>('hour');
const now = ref(new Date());
const hovered = ref<{ chart: Chart; index: number } | null>(null);

const since = computed(() => period.value.start(now.value));

const { result, error, refetch } = useQuery(TrendsDocument, () => ({ since: since.value.toISOString() }), {
  fetchPolicy: 'cache-and-network',
});

watch(() => props.loadedAt, () => {
  now.value = new Date();
  refetch();
});

function pickPeriod(next: Period) {
  if (!next.groupings.includes(grouping.value)) grouping.value = next.fallback;

  period.value = next;
  now.value = new Date();
  hovered.value = null;
}

function pickGrouping(next: Grouping) {
  grouping.value = next;
  hovered.value = null;
}

function hover(chart: Chart, index: number | null) {
  hovered.value = index == null ? null : { chart, index };
}

const signups = computed(() => result.value?.signups ?? []);
const standings = computed(() => result.value?.workspaceStandings ?? []);
const signupTotal = computed(() => signups.value.reduce((sum, hour) => sum + hour.count, 0));
const latest = computed(() => standings.value.at(-1));

const starts = computed(() => {
  const ends = [now.value.getTime()];
  const lastSignup = signups.value.at(-1);

  if (lastSignup) ends.push(Date.parse(lastSignup.start));
  if (latest.value) ends.push(Date.parse(latest.value.at) - 1);

  return bucketStarts(since.value, Math.max(...ends), grouping.value);
});

const signupBars = computed<Bar[]>(() => {
  const counts = new Map<number, number>();

  for (const hour of signups.value) {
    const key = bucketStart(new Date(hour.start), grouping.value);
    counts.set(key, (counts.get(key) ?? 0) + hour.count);
  }

  return starts.value.map((start, index) => {
    const partial = index === starts.value.length - 1;
    const title = spanTitle(start, grouping.value);

    return { title: partial ? `${title}, so far` : title, values: [counts.get(start) ?? 0], partial };
  });
});

const standingBars = computed<Bar[]>(() => {
  const last = new Map<number, (typeof standings.value)[number]>();

  for (const standing of standings.value) {
    last.set(bucketStart(new Date(Date.parse(standing.at) - 1), grouping.value), standing);
  }

  return starts.value.map((start, index) => {
    const standing = last.get(start);
    const partial = index === starts.value.length - 1;

    return {
      title: partial ? 'Now' : endTitle(start, grouping.value),
      values: standing ? [standing.abuse, standing.billing, standing.good] : [0, 0, 0],
      partial,
    };
  });
});

const tickLabels = computed(() => ticks(starts.value, period.value, grouping.value));
const range = computed(() => (period.value.key === 'today' ? `Since ${time(since.value)}` : `Since ${fullDate(since.value)}`));

const signupMessage = computed(() => {
  if (error.value) return error.value.message;
  if (result.value && !signups.value.length) return 'No signup metrics for this period';
  return undefined;
});

const standingMessage = computed(() => {
  if (error.value) return error.value.message;
  if (result.value && !standings.value.length) return 'No workspace metrics for this period';
  return undefined;
});
</script>

<template>
  <section aria-label="Signups and workspaces" class="sticky left-0 border-b px-4 pt-4 pb-5">
    <div class="flex flex-wrap items-center gap-2">
      <div role="group" aria-label="Period" class="flex h-8 gap-0.5 rounded-sm border bg-background p-0.5">
        <button
          v-for="option in periods"
          :key="option.key"
          type="button"
          class="rounded-xs px-2.5 whitespace-nowrap"
          :class="option.key === period.key ? 'bg-card shadow-xs' : 'text-muted-foreground hover:text-foreground'"
          :aria-pressed="option.key === period.key"
          @click="pickPeriod(option)"
        >
          {{ option.label }}
        </button>
      </div>
      <span class="ml-2 text-muted-foreground">{{ range }}</span>
      <div role="group" aria-label="Group by" class="ml-auto flex h-8 gap-0.5 rounded-sm border bg-background p-0.5">
        <button
          v-for="option in groupings"
          :key="option.value"
          type="button"
          class="rounded-xs px-2.5 whitespace-nowrap disabled:cursor-not-allowed disabled:opacity-45"
          :class="option.value === grouping ? 'bg-card shadow-xs' : 'text-muted-foreground enabled:hover:text-foreground'"
          :aria-pressed="option.value === grouping"
          :disabled="!period.groupings.includes(option.value)"
          @click="pickGrouping(option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </div>

    <div class="mt-5 flex flex-wrap gap-x-12 gap-y-6">
      <div class="flex min-w-0 flex-[1_1_520px] flex-col gap-3.5">
        <div class="flex flex-col gap-0.5">
          <span class="text-muted-foreground">Signups</span>
          <span class="text-[28px] leading-9 font-semibold tabular-nums">{{ result ? signupTotal : '–' }}</span>
        </div>
        <BarChart
          :bars="signupBars"
          :layers="signupLayers"
          :ticks="tickLabels"
          :label="`Signups per ${grouping}, ${range.toLowerCase()}`"
          :hovered="hovered?.index ?? null"
          :tip="hovered?.chart === 'signups'"
          :message="signupMessage"
          :failed="!!error"
          @hover="hover('signups', $event)"
        />
      </div>

      <div class="flex min-w-0 flex-[1_1_520px] flex-col gap-3.5">
        <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
          <div class="flex flex-col gap-0.5">
            <span class="text-muted-foreground">Workspaces</span>
            <span class="text-[28px] leading-9 font-semibold tabular-nums">{{ latest ? latest.good + latest.billing + latest.abuse : '–' }}</span>
          </div>
          <span v-if="latest" class="flex flex-wrap items-center gap-x-3.5 gap-y-1 text-muted-foreground tabular-nums">
            <span class="flex items-center gap-1.5"><span class="size-2 rounded-[2px] bg-foreground/15" />{{ latest.good }} good</span>
            <span class="flex items-center gap-1.5" :class="latest.billing ? 'text-status-warn' : ''">
              <span class="size-2 rounded-[2px] bg-status-warn/55" />{{ latest.billing }} billing
            </span>
            <span class="flex items-center gap-1.5" :class="latest.abuse ? 'text-status-danger' : ''">
              <span class="size-2 rounded-[2px] bg-status-danger" />{{ latest.abuse }} abuse
            </span>
          </span>
        </div>
        <BarChart
          :bars="standingBars"
          :layers="standingLayers"
          :ticks="tickLabels"
          :label="`Workspaces by standing at the end of each ${grouping}, ${range.toLowerCase()}`"
          :hovered="hovered?.index ?? null"
          :tip="hovered?.chart === 'workspaces'"
          :message="standingMessage"
          :failed="!!error"
          @hover="hover('workspaces', $event)"
        />
      </div>
    </div>
  </section>
</template>
