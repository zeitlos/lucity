<script setup lang="ts">
import { computed } from 'vue';

import type { AppChartVersionsQuery, PlatformTotalsQuery } from '@/gql/graphql';
import { count, percent, shortDate, sparkline } from '@/lib/platform';

type Totals = PlatformTotalsQuery['platformTotals'];
type Version = AppChartVersionsQuery['appChartVersions'][number];
type Tone = 'none' | 'warn' | 'alert';

const props = defineProps<{
  totals: Totals | null;
  totalsError: string | null;
  versions: Version[] | null;
  versionsError: string | null;
}>();

function sinceYesterday(trend: number[], value: number): string {
  if (trend.length === 0) return '';

  const difference = Math.round(value - trend[0]!);

  if (difference === 0) return 'Same as yesterday';
  return `${difference > 0 ? '+' : '−'}${count(Math.abs(difference))} since yesterday`;
}

function figure(label: string, value: number | undefined, sub: string, trend: number[], tone: Tone = 'none') {
  return { label, value: value == null ? '–' : count(value), sub, tone, spark: sparkline(trend, 200, 28) };
}

const figures = computed(() => {
  const totals = props.totals;

  if (!totals) {
    return ['Pods running', 'Containers', 'Workspaces running', 'Releases', 'Firing'].map((label) => figure(label, undefined, '', []));
  }

  const waiting = [
    totals.pendingPods ? `${count(totals.pendingPods)} pending` : '',
    totals.failedPods ? `${count(totals.failedPods)} failed` : '',
  ].filter(Boolean).join(' · ');

  return [
    figure('Pods running', totals.runningPods, waiting || 'None pending', totals.runningPodsTrend, waiting ? 'warn' : 'none'),
    figure('Containers', totals.runningContainers, sinceYesterday(totals.runningContainersTrend, totals.runningContainers), totals.runningContainersTrend),
    figure('Workspaces running', totals.runningWorkspaces, `of ${count(totals.workspaces)} workspaces`, totals.runningWorkspacesTrend),
    figure(
      'Releases',
      totals.releases,
      totals.failedReleases ? `${count(totals.failedReleases)} failed` : 'None failed',
      totals.releasesTrend,
      totals.failedReleases ? 'warn' : 'none',
    ),
    figure(
      'Firing',
      totals.firingAlerts.length,
      totals.firingAlerts.join(', ') || 'Nothing firing',
      totals.firingAlertsTrend,
      totals.firingAlerts.length ? 'alert' : 'none',
    ),
  ];
});

const rollout = computed(() => {
  const versions = props.versions;

  if (!versions || versions.length === 0) return null;

  const total = versions.reduce((sum, version) => sum + version.releases, 0);
  const newest = versions[0]!;

  return { total, newest, share: newest.releases / total };
});
</script>

<template>
  <section aria-label="Totals" class="hero m-4 flex flex-wrap overflow-hidden rounded-3xl text-ink-foreground tabular-nums">
    <p v-if="totalsError && !totals" class="basis-full px-5 pt-4 text-[oklch(0.82_0.12_65)]">{{ totalsError }}</p>
    <div
      v-for="item in figures"
      :key="item.label"
      class="flex min-w-0 flex-[1_1_170px] flex-col gap-1.5 border-r border-white/9 px-5 pt-5 pb-4.5"
    >
      <span class="text-ink-muted-foreground">{{ item.label }}</span>
      <span
        class="text-4xl font-semibold tracking-tight"
        :class="item.tone === 'alert'
          ? 'text-[oklch(0.74_0.16_28)] [text-shadow:0_0_24px_oklch(0.65_0.2_28/0.55)]'
          : '[text-shadow:0_0_24px_oklch(0.95_0.04_80/0.25)]'"
      >{{ item.value }}</span>
      <span class="truncate" :class="item.tone === 'none' ? 'text-ink-muted-foreground' : 'text-[oklch(0.82_0.12_65)]'">{{ item.sub }}</span>
      <svg viewBox="0 0 200 28" preserveAspectRatio="none" class="mt-1.5 block h-7 w-full overflow-visible" aria-hidden="true">
        <path :d="item.spark.area" fill="oklch(1 0 0 / 0.06)" />
        <path
          :d="item.spark.line"
          fill="none"
          :stroke="item.tone === 'alert' ? 'oklch(0.74 0.16 28 / 0.8)' : 'oklch(0.95 0.03 80 / 0.7)'"
          stroke-width="1.5"
          stroke-linejoin="round"
          vector-effect="non-scaling-stroke"
        />
      </svg>
    </div>

    <div class="flex min-w-0 flex-[2_1_380px] flex-col gap-2.5 bg-white/3 px-5 pt-5 pb-4.5">
      <span class="flex items-baseline gap-2">
        <span class="text-ink-muted-foreground">App chart rollout</span>
        <span v-if="rollout" class="ml-auto text-ink-muted-foreground">Since {{ shortDate(rollout.newest.since) }}</span>
      </span>
      <template v-if="rollout">
        <span class="flex flex-wrap items-baseline gap-x-2.5">
          <span class="text-4xl font-semibold tracking-tight">{{ percent(rollout.share) }}</span>
          <span class="text-ink-muted-foreground">
            {{ count(rollout.newest.releases) }} of {{ count(rollout.total) }} releases on
            <span class="font-mono text-ink-foreground">{{ rollout.newest.version }}</span>
          </span>
        </span>
        <span
          role="img"
          :aria-label="`${count(rollout.newest.releases)} of ${count(rollout.total)} releases on ${rollout.newest.version}`"
          class="relative mt-2 flex h-3 overflow-hidden rounded-md bg-white/8 shadow-[inset_0_1px_2px_oklch(0_0_0/0.35)]"
        >
          <span
            class="rounded-md bg-linear-to-b from-[oklch(0.97_0.02_80)] to-[oklch(0.84_0.04_75)] shadow-[0_0_16px_oklch(0.95_0.04_80/0.55)]"
            :style="{ width: `${rollout.share * 100}%` }"
          />
          <span v-for="tick in 9" :key="tick" class="absolute inset-y-0 w-px bg-black/18" :style="{ left: `${tick * 10}%` }" />
        </span>
        <span class="flex flex-wrap gap-x-5 gap-y-1 font-mono">
          <span v-for="(version, index) in versions" :key="version.version" class="flex items-center gap-2">
            <span class="size-2 rounded-[2px]" :class="index === 0 ? 'bg-[oklch(0.93_0.03_80)]' : 'bg-white/20'" aria-hidden="true" />
            {{ version.version }}
            <span class="text-ink-muted-foreground">{{ count(version.releases) }}</span>
          </span>
        </span>
      </template>
      <span v-else-if="versionsError" class="text-[oklch(0.82_0.12_65)]">{{ versionsError }}</span>
      <span v-else-if="versions" class="text-ink-muted-foreground">No releases deployed</span>
      <span v-else class="text-ink-muted-foreground">Reading releases…</span>
    </div>
  </section>
</template>

<style scoped>
.hero {
  background:
    radial-gradient(oklch(1 0 0 / 0.05) 1px, transparent 1.3px) 0 0 / 9px 9px,
    radial-gradient(130% 150% at 0% 0%, oklch(0.34 0.03 55) 0%, oklch(0.25 0.022 55) 48%, var(--ink) 100%);
  box-shadow:
    inset 0 1px 0 oklch(1 0 0 / 0.14),
    inset 0 -1px 0 oklch(0 0 0 / 0.25),
    0 24px 48px -28px oklch(0.22 0.03 55 / 0.75);
}
</style>
