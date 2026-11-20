<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';

import { graphql } from '@/gql';
import { bucketStart, bucketStarts, fullDate, shortDate } from '@/lib/trends';

const DAYS = 30;
const WIDTH = 300;
const HEIGHT = 64;

const plans = [
  { key: 'pro', label: 'pro', fill: 'fill-foreground' },
  { key: 'hobby', label: 'hobby', fill: 'fill-foreground/55' },
  { key: 'trial', label: 'trial', fill: 'fill-foreground/30' },
  { key: 'none', label: 'no subscription', fill: 'fill-foreground/12' },
] as const;

const PlanHistoryDocument = graphql(`
  query PlanHistory($since: Time!) {
    workspacePlans(since: $since) { at none trial hobby pro }
  }
`);

const now = new Date();
const since = new Date(now.getFullYear(), now.getMonth(), now.getDate() - (DAYS - 1));

const { result, error } = useQuery(PlanHistoryDocument, { since: since.toISOString() }, { fetchPolicy: 'cache-and-network' });

const samples = computed(() => result.value?.workspacePlans ?? []);

const days = computed(() => {
  const last = new Map<number, (typeof samples.value)[number]>();

  for (const sample of samples.value) {
    last.set(bucketStart(new Date(Date.parse(sample.at) - 1), 'day'), sample);
  }

  const starts = bucketStarts(since, now.getTime(), 'day');
  const peak = Math.max(1, ...[...last.values()].map((sample) => sample.none + sample.trial + sample.hobby + sample.pro));
  const slot = WIDTH / starts.length;

  return starts.map((start, index) => {
    const sample = last.get(start);
    const date = fullDate(new Date(start));
    let base = 0;

    if (!sample) return { start, title: `${date}: no data`, rects: [] };

    const rects = plans.flatMap((plan) => {
      const count = sample[plan.key];

      if (!count) return [];

      const height = (count / peak) * HEIGHT;
      const rect = {
        fill: plan.fill,
        x: index * slot + 1,
        y: HEIGHT - base - height,
        width: slot - 2,
        height: Math.max(height - (base ? 1 : 0), 0.5),
      };

      base += height;
      return [rect];
    });

    return { start, title: `${date}: ${plans.map((plan) => `${sample[plan.key]} ${plan.label}`).join(', ')}`, rects };
  });
});
</script>

<template>
  <div class="mx-2 mt-1 flex flex-col gap-1.5 border-t border-hairline pt-2.5 pb-1 text-xs">
    <p v-if="error" class="flex h-16 items-center text-status-danger">{{ error.message }}</p>
    <p v-else-if="result && !samples.length" class="flex h-16 items-center text-muted-foreground">No plan history yet</p>
    <svg
      v-else
      :viewBox="`0 0 ${WIDTH} ${HEIGHT}`"
      preserveAspectRatio="none"
      class="block h-16 w-full"
      role="img"
      :aria-label="`Workspaces by plan at the end of each of the last ${DAYS} days`"
    >
      <g v-for="day in days" :key="day.start">
        <title>{{ day.title }}</title>
        <rect
          v-for="rect in day.rects"
          :key="rect.fill"
          :x="rect.x"
          :y="rect.y"
          :width="rect.width"
          :height="rect.height"
          :class="rect.fill"
        />
      </g>
    </svg>
    <div class="flex justify-between text-muted-foreground">
      <span>{{ shortDate(since) }}</span>
      <span>Today</span>
    </div>
  </div>
</template>
