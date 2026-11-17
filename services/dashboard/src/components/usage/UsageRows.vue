<script setup lang="ts">
import { computed } from 'vue';
import { Bot, Box, ChevronRight } from '@lucide/vue';
import ResourceIcon from '@/components/ResourceIcon.vue';
import UsageBars from '@/components/usage/UsageBars.vue';
import { usePanel } from '@/composables/usePanel';
import { UsageResourceKind } from '@/gql/graphql';
import type { ResourceKind } from '@/lib/resources';
import { dayDate, foldRow, formatCents, formatDay, formatShare, sum, type UsageLevel, type UsageRow } from '@/lib/usage';
import { cn } from '@/lib/utils';

const props = defineProps<{
  rows: UsageRow[];
  tail: UsageRow[];
  level: UsageLevel;
  totalLabel: string;
  totalNote?: string;
  periodStart: string;
  measuredUntil: string;
  days: number;
  lastDay: number;
  running: boolean;
}>();

const expanded = defineModel<Set<string>>('expanded', { required: true });

const { openPanel } = usePanel();

interface Line {
  row: UsageRow;
  level: number;
  expandable: boolean;
  open: boolean;
}

const headings: Record<UsageLevel, string> = {
  project: 'Project',
  environment: 'Environment',
  resource: 'Resource',
};

const resourceKinds: Partial<Record<UsageResourceKind, ResourceKind>> = {
  [UsageResourceKind.Database]: 'database',
  [UsageResourceKind.KeyValueStore]: 'keyValueStore',
  [UsageResourceKind.Bucket]: 'bucket',
  [UsageResourceKind.Volume]: 'volume',
};

const grid = 'grid grid-cols-[minmax(0,1fr)_2.5rem_5rem] items-center gap-x-4 px-4 sm:grid-cols-[minmax(0,1fr)_minmax(12rem,22rem)_3rem_6rem] sm:gap-x-6 sm:px-6';
const heights = ['h-11', 'h-9', 'h-8'];
const indents = ['', 'pl-6.5', 'pl-13'];

const axis = computed(() => [0, Math.floor((props.days - 1) / 2), props.days - 1].map(i => formatDay(dayDate(props.periodStart, i))));
const fold = computed(() => (props.tail.length ? foldRow(props.tail, props.level) : null));
const totalCents = computed(() => sum([...props.rows, ...props.tail].map(row => row.amountCents)));

const peak = computed(() => {
  const rows = [...props.rows, ...props.tail];
  if (fold.value) rows.push(fold.value);
  return Math.max(0, ...rows.flatMap(row => row.daily));
});

function toggle(key: string) {
  const next = new Set(expanded.value);
  if (next.has(key)) next.delete(key);
  else next.add(key);
  expanded.value = next;
}

function lines(row: UsageRow): Line[] {
  const open = expanded.value.has(row.key);
  const list: Line[] = [{ row, level: 0, expandable: row.children.length > 0, open }];
  if (!open) return list;

  for (const child of row.children) {
    const childOpen = expanded.value.has(child.key);
    list.push({ row: child, level: 1, expandable: child.children.length > 0, open: childOpen });
    if (!childOpen) continue;

    for (const leaf of child.children) {
      list.push({ row: leaf, level: 2, expandable: false, open: false });
    }
  }

  return list;
}

const groups = computed(() => {
  const list = props.rows.map(row => ({ key: row.key, lines: lines(row) }));
  if (!fold.value) return list;

  const open = expanded.value.has(fold.value.key);
  list.push({ key: fold.value.key, lines: [{ row: fold.value, level: 0, expandable: true, open }] });
  if (open) list.push(...props.tail.map(row => ({ key: row.key, lines: lines(row) })));

  return list;
});

function sources(row: UsageRow) {
  const colored = row.children.filter(child => child.color);
  return colored.length ? colored : [row];
}

function share(line: Line) {
  if (line.level > 0) return '';
  return formatShare(totalCents.value > 0 ? line.row.amountCents / totalCents.value : 0);
}

async function follow(event: MouseEvent, navigate: (event?: MouseEvent) => Promise<unknown>, row: UsageRow) {
  const plain = event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  await navigate(event);
  if (plain && row.panel) await openPanel(row.panel);
}
</script>

<template>
  <div role="table" :aria-label="`Usage by ${level}`" class="-mx-4 text-sm tabular-nums sm:-mx-6">
    <div role="row" :class="cn(grid, 'pb-1.5 text-xs text-muted-foreground')">
      <span role="columnheader" class="font-medium">{{ headings[level] }}</span>
      <span role="columnheader" aria-label="Daily cost" class="flex justify-between max-sm:hidden">
        <span v-for="label in axis" :key="label">{{ label }}</span>
      </span>
      <span role="columnheader" class="text-right font-medium">Share</span>
      <span role="columnheader" class="text-right font-medium">Amount</span>
    </div>

    <div v-for="group in groups" :key="group.key" role="rowgroup" class="border-t">
      <div
        v-for="line in group.lines"
        :key="line.row.key"
        role="row"
        :class="cn(grid, line.expandable && 'cursor-pointer hover:bg-muted/60')"
        @click="line.expandable && toggle(line.row.key)"
      >
        <span role="rowheader" :class="cn('col-start-1 row-start-1 flex min-w-0 items-center gap-2', heights[line.level], indents[line.level])">
          <button
            v-if="line.expandable"
            type="button"
            :aria-expanded="line.open"
            :aria-label="`Details of ${line.row.name}`"
            class="-m-1 flex shrink-0 rounded-sm p-1 text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none"
          >
            <ChevronRight :size="16" :class="cn('transition-transform', line.open && 'rotate-90')" />
          </button>
          <span v-if="line.row.color" class="size-2 shrink-0 rounded-xs" :style="{ background: line.row.color }" />
          <span v-if="line.row.kind" class="flex size-4 shrink-0 items-center justify-center text-muted-foreground">
            <Bot v-if="line.row.kind === UsageResourceKind.Agent" :size="16" />
            <ResourceIcon v-else-if="resourceKinds[line.row.kind]" :kind="resourceKinds[line.row.kind]!" :size="16" />
            <Box v-else :size="16" />
          </span>
          <RouterLink v-if="line.row.to" v-slot="{ href, navigate }" :to="line.row.to" custom>
            <a
              :href="href"
              :class="cn(
                'truncate rounded-sm hover:underline focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none',
                line.level === 0 && 'font-semibold',
              )"
              @click.stop="follow($event, navigate, line.row)"
            >
              {{ line.row.name }}
            </a>
          </RouterLink>
          <span v-else :class="cn('truncate', line.level === 0 && 'font-semibold')">{{ line.row.name }}</span>
          <span v-if="line.row.deletedAt" class="min-w-0 flex-1 truncate text-xs text-muted-foreground">
            deleted {{ formatDay(new Date(line.row.deletedAt)) }}
          </span>
        </span>
        <UsageBars
          v-if="line.level === 0"
          class="col-span-3 row-start-2 mb-2 max-sm:h-6 sm:col-span-1 sm:col-start-2 sm:row-start-1 sm:mb-0"
          :label="`Daily cost of ${line.row.name}`"
          :sources="sources(line.row)"
          :period-start="periodStart"
          :measured-until="measuredUntil"
          :days="days"
          :last-day="lastDay"
          :running="running"
          :peak="peak"
        />
        <span role="cell" class="col-start-2 row-start-1 text-right text-muted-foreground sm:col-start-3">{{ share(line) }}</span>
        <span
          role="cell"
          :class="cn(
            'col-start-3 row-start-1 flex items-center justify-end sm:col-start-4',
            line.level === 0 && 'font-semibold',
            line.level > 0 && !line.expandable && 'text-muted-foreground',
          )"
        >
          {{ formatCents(line.row.amountCents) }}
        </span>
      </div>
    </div>

    <div role="row" class="grid grid-cols-[minmax(0,1fr)_auto] gap-x-6 border-t px-4 py-3 font-semibold sm:px-6">
      <span role="rowheader">
        {{ totalLabel }}<span v-if="totalNote" class="font-normal text-muted-foreground"> · {{ totalNote }}</span>
      </span>
      <span role="cell" class="text-right">{{ formatCents(totalCents) }}</span>
    </div>
  </div>
</template>
