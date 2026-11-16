<script setup lang="ts">
import { computed } from 'vue';
import { Bot, Box, ChevronRight } from '@lucide/vue';
import ResourceIcon from '@/components/ResourceIcon.vue';
import { UsageResourceKind } from '@/gql/graphql';
import type { ResourceKind } from '@/lib/resources';
import { dayDate, formatCents, formatDay, formatShare, type UsageProject } from '@/lib/usage';
import { cn } from '@/lib/utils';

const props = defineProps<{
  projects: UsageProject[];
  colors: Map<string, string>;
  totalCents: number;
  periodStart: string;
  days: number;
  lastDay: number;
}>();

const expanded = defineModel<Set<string>>('expanded', { required: true });

const resourceKinds: Partial<Record<UsageResourceKind, ResourceKind>> = {
  [UsageResourceKind.Database]: 'database',
  [UsageResourceKind.KeyValueStore]: 'keyValueStore',
  [UsageResourceKind.Bucket]: 'bucket',
  [UsageResourceKind.Volume]: 'volume',
};

const PROJECT_ROW = 44;
const ENVIRONMENT_ROW = 36;
const RESOURCE_ROW = 32;

const axis = computed(() => [0, Math.floor((props.days - 1) / 2), props.days - 1].map(i => formatDay(dayDate(props.periodStart, i))));

function toggle(key: string) {
  const next = new Set(expanded.value);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  expanded.value = next;
}

function deletedNote(deletedAt?: string | null) {
  return deletedAt ? `deleted ${formatDay(new Date(deletedAt))}` : '';
}

const sections = computed(() => props.projects.map((project) => {
  const open = expanded.value.has(project.projectId);
  const environments = open
    ? project.environments.map(environment => ({
        environment,
        open: expanded.value.has(environment.environmentId),
      }))
    : [];

  let rowCount = 1;
  let height = PROJECT_ROW;

  for (const { environment, open: environmentOpen } of environments) {
    rowCount += 1;
    height += ENVIRONMENT_ROW;

    if (environmentOpen) {
      rowCount += environment.resources.length;
      height += environment.resources.length * RESOURCE_ROW;
    }
  }

  const peak = Math.max(...project.daily, 0.000001);
  const days = Array.from({ length: props.days }, (_, i) => {
    const date = formatDay(dayDate(props.periodStart, i));

    if (i > props.lastDay) {
      return { key: i, title: date, future: true, segments: [] };
    }

    return {
      key: i,
      title: `${date}: ${formatCents(project.daily[i] ?? 0)}`,
      future: false,
      segments: project.environments
        .filter(environment => (environment.daily[i] ?? 0) > 0)
        .map(environment => ({
          key: environment.environmentId,
          height: `${Math.max(0.5, ((environment.daily[i] ?? 0) / peak) * 100).toFixed(2)}%`,
          color: props.colors.get(environment.environmentId),
        })),
    };
  });

  return {
    project,
    open,
    environments,
    rowCount,
    chartHeight: open ? Math.max(28, Math.min(96, height - 20)) : 28,
    days,
  };
}));
</script>

<template>
  <div role="treegrid" aria-label="Usage by project and environment" class="min-w-160 text-sm tabular-nums">
    <div role="row" class="grid grid-cols-[minmax(0,1fr)_minmax(12rem,22rem)_7rem] gap-x-6 pb-1.5 text-xs text-muted-foreground">
      <span role="columnheader" class="font-medium">Project</span>
      <span role="columnheader" class="flex justify-between">
        <span v-for="label in axis" :key="label">{{ label }}</span>
      </span>
      <span role="columnheader" class="text-right font-medium">Amount</span>
    </div>

    <div
      v-for="section in sections"
      :key="section.project.projectId"
      role="rowgroup"
      :aria-label="section.project.name"
      class="grid grid-cols-[minmax(0,1fr)_minmax(12rem,22rem)_7rem] gap-x-6 border-t pt-0.5 pb-1.5"
    >
      <div role="row" aria-level="1" :aria-expanded="section.open" class="contents">
        <span role="rowheader" class="col-start-1 row-start-1 flex h-11 min-w-0 items-center">
          <button
            type="button"
            :aria-expanded="section.open"
            class="flex h-11 min-w-0 items-center gap-2.5 text-left"
            @click="toggle(section.project.projectId)"
          >
            <ChevronRight :size="16" :class="cn('shrink-0 text-muted-foreground transition-transform', section.open && 'rotate-90')" />
            <span class="size-2 shrink-0 rounded-xs" :style="{ background: colors.get(section.project.projectId) }" />
            <span class="truncate font-semibold">{{ section.project.name }}</span>
            <span class="text-xs whitespace-nowrap text-muted-foreground">
              {{ formatShare(totalCents > 0 ? section.project.amountCents / totalCents : 0) }}
              <template v-if="section.project.deletedAt">· {{ deletedNote(section.project.deletedAt) }}</template>
            </span>
          </button>
        </span>
        <div
          role="img"
          :aria-label="`Daily cost of ${section.project.name} by environment`"
          :class="cn('col-start-2 flex items-end gap-0.5 border-b', section.open ? 'mt-2.5 self-start' : 'self-center')"
          :style="{ gridRow: `1 / span ${section.rowCount}`, height: `${section.chartHeight}px` }"
        >
          <div
            v-for="day in section.days"
            :key="day.key"
            :title="day.title"
            class="flex h-full min-w-0 flex-1 flex-col-reverse overflow-hidden rounded-t-xs"
          >
            <div v-if="day.future" class="h-0.5 bg-muted" />
            <div
              v-for="segment in day.segments"
              :key="segment.key"
              :style="{ height: segment.height, background: segment.color }"
            />
          </div>
        </div>
        <span role="cell" class="col-start-3 row-start-1 flex h-11 items-center justify-end font-semibold">
          {{ formatCents(section.project.amountCents) }}
        </span>
      </div>

      <template v-for="{ environment, open } in section.environments" :key="environment.environmentId">
        <div role="row" aria-level="2" :aria-expanded="open" class="contents">
          <span role="rowheader" class="col-start-1 flex h-9 min-w-0 items-center pl-6.5">
            <button
              type="button"
              :aria-expanded="open"
              class="flex h-9 min-w-0 items-center gap-2 text-left"
              @click="toggle(environment.environmentId)"
            >
              <ChevronRight :size="16" :class="cn('shrink-0 text-muted-foreground transition-transform', open && 'rotate-90')" />
              <span class="size-2 shrink-0 rounded-xs" :style="{ background: colors.get(environment.environmentId) }" />
              <span class="truncate">{{ environment.name }}</span>
              <span v-if="environment.deletedAt" class="text-xs whitespace-nowrap text-muted-foreground">
                {{ deletedNote(environment.deletedAt) }}
              </span>
            </button>
          </span>
          <span role="cell" class="col-start-3 flex h-9 items-center justify-end">
            {{ formatCents(environment.amountCents) }}
          </span>
        </div>

        <template v-if="open">
          <div
            v-for="resource in environment.resources"
            :key="`${resource.kind}:${resource.resourceId}`"
            role="row"
            aria-level="3"
            class="contents"
          >
            <span role="rowheader" class="col-start-1 flex h-8 min-w-0 items-center gap-2 pl-13">
              <span class="flex size-4 shrink-0 items-center justify-center text-muted-foreground">
                <Bot v-if="resource.kind === UsageResourceKind.Agent" :size="16" />
                <ResourceIcon v-else-if="resourceKinds[resource.kind]" :kind="resourceKinds[resource.kind]!" :size="16" />
                <Box v-else :size="16" />
              </span>
              <span class="truncate">{{ resource.name }}</span>
            </span>
            <span role="cell" class="col-start-3 flex h-8 items-center justify-end text-muted-foreground">
              {{ formatCents(resource.amountCents) }}
            </span>
          </div>
        </template>
      </template>
    </div>
  </div>
</template>
