<script setup lang="ts">
import { computed } from 'vue';

import type { RoutinesQuery } from '@/gql/graphql';
import { ago, change, hour, seconds } from '@/lib/platform';

type Routine = RoutinesQuery['routines'][number];
type Hour = Routine['hours'][number];

const props = defineProps<{
  routines: Routine[] | null;
  error: string | null;
}>();

const names: Record<string, string> = {
  releases: 'Releases',
  admission: 'Admission',
  domains_pending: 'Pending domains',
  domains_sync: 'Domain sync',
  backups: 'Backups',
  metering: 'Metering',
  object_storage: 'Object storage',
  edge_zones: 'Edge zones',
};

const order = Object.keys(names);
const slower = 0.5;
const columns = 'grid-cols-[minmax(140px,1fr)_238px_132px_88px]';

function name(routine: Routine): string {
  const known = names[routine.name];

  if (known) return known;

  const words = routine.name.replaceAll('_', ' ');
  return words.charAt(0).toUpperCase() + words.slice(1);
}

function rank(routine: Routine): number {
  const index = order.indexOf(routine.name);
  return index < 0 ? order.length : index;
}

function slowed(routine: Routine): boolean {
  return routine.averageDurationChange != null && routine.averageDurationChange >= slower;
}

function cellClass(cell: Hour): string {
  if (cell.failed > 0) return 'bg-linear-to-b from-[oklch(0.58_0.19_25)] to-[oklch(0.47_0.19_25)] shadow-[inset_0_1px_0_oklch(1_0_0/0.22)]';
  if (cell.passed > 0) return 'bg-linear-to-b from-foreground/42 to-foreground/55 shadow-[inset_0_1px_0_oklch(1_0_0/0.22)]';
  return 'border border-foreground/30';
}

function cellLabel(cell: Hour): string {
  const parts = [cell.failed ? `${cell.failed} failed` : '', cell.passed ? `${cell.passed} passed` : ''].filter(Boolean);
  return `${hour(cell.start)}, ${parts.join(', ') || 'no runs'}`;
}

const sorted = computed(() => [...(props.routines ?? [])].sort((left, right) => rank(left) - rank(right) || left.name.localeCompare(right.name)));

const halves = computed(() => {
  const half = Math.ceil(sorted.value.length / 2);
  return [sorted.value.slice(0, half), sorted.value.slice(half)];
});

const ticks = computed(() => {
  const hours = sorted.value[0]?.hours ?? [];
  return [0, 8, 16].filter((index) => hours[index]).map((index) => ({ index, text: hour(hours[index]!.start) }));
});
</script>

<template>
  <section aria-labelledby="routines-title" class="border-t">
    <div class="flex min-h-11 flex-wrap items-center gap-x-4 gap-y-2 px-4">
      <h2 id="routines-title" class="font-semibold">Routines</h2>
      <span class="ml-auto flex items-center gap-1.5 text-muted-foreground">
        <span class="h-3 w-2 rounded-[2px] bg-foreground/50" aria-hidden="true" />Passed
      </span>
      <span class="flex items-center gap-1.5 text-muted-foreground">
        <span class="h-3 w-2 rounded-[2px] bg-status-danger" aria-hidden="true" />Failed
      </span>
      <span class="flex items-center gap-1.5 text-muted-foreground">
        <span class="h-3 w-2 rounded-[2px] border border-foreground/35" aria-hidden="true" />No runs
      </span>
    </div>

    <p v-if="error" class="border-t border-hairline px-4 py-3 text-status-danger">{{ error }}</p>
    <p v-else-if="routines && routines.length === 0" class="border-t border-hairline px-4 py-3 text-muted-foreground">No routine ran in the last 7 days.</p>

    <div v-else class="overflow-x-auto">
      <div class="grid min-w-[1340px] grid-cols-2 border-t border-hairline">
        <div v-for="(half, column) in halves" :key="column" :class="column === 0 ? 'border-r border-hairline' : ''">
          <div class="grid h-9 items-center gap-x-5 border-b border-hairline bg-sidebar px-4 text-muted-foreground" :class="columns">
            <span>Routine</span>
            <span class="relative h-full">
              <span
                v-for="tick in ticks"
                :key="tick.index"
                class="absolute top-2 whitespace-nowrap"
                :style="{ left: `${tick.index * 10}px` }"
              >{{ tick.text }}</span>
              <span class="absolute top-2 right-0">now</span>
            </span>
            <span class="text-right">Avg duration</span>
            <span class="text-right">Last passed</span>
          </div>

          <template v-if="!routines">
            <div v-for="index in 4" :key="index" class="grid min-h-10 items-center gap-x-5 border-b border-hairline px-4" :class="columns">
              <span class="h-4 w-28 animate-pulse rounded-md bg-muted" />
              <span class="h-3.5 w-full animate-pulse rounded-[2px] bg-muted" />
            </div>
          </template>

          <div
            v-for="routine in half"
            :key="routine.name"
            class="grid min-h-10 items-center gap-x-5 border-b border-hairline px-4 tabular-nums"
            :class="columns"
          >
            <span class="truncate font-medium">{{ name(routine) }}</span>
            <span class="flex gap-0.5">
              <span
                v-for="cell in routine.hours"
                :key="cell.start"
                role="img"
                :aria-label="cellLabel(cell)"
                class="box-border h-3.5 w-2 shrink-0 rounded-[2px]"
                :class="cellClass(cell)"
              />
            </span>
            <span class="flex items-center justify-end gap-2">
              <span>{{ routine.averageDuration == null ? '–' : seconds(routine.averageDuration) }}</span>
              <span
                class="min-w-13 rounded-md px-1.5 text-right"
                :class="slowed(routine) ? 'bg-status-warn/14 font-semibold text-status-warn' : 'text-muted-foreground'"
              >{{ routine.averageDurationChange == null ? '' : slowed(routine) ? `↑ ${change(routine.averageDurationChange).slice(1)}` : change(routine.averageDurationChange) }}</span>
            </span>
            <span class="text-right text-muted-foreground">{{ routine.lastPassedAt ? ago(routine.lastPassedAt) : '–' }}</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
