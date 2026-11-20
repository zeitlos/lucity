<script setup lang="ts">
import { computed, ref } from 'vue';
import { ChevronRight } from '@lucide/vue';

import type { NodePoolsQuery } from '@/gql/graphql';
import { age } from '@/lib/format';
import { GiB, count, oneDecimal, tone, toneText, type Tone } from '@/lib/platform';

type Pool = NodePoolsQuery['nodePools'][number];
type Node = Pool['nodes'][number];

const props = defineProps<{
  pools: Pool[] | null;
  error: string | null;
}>();

const open = ref(true);

const fills: Record<Tone, string> = {
  none: 'bg-foreground/20',
  warn: 'bg-status-warn/60',
  danger: 'bg-status-danger',
};

function meter(name: string, value: number, cap: number, used: number | null, format: (value: number) => string) {
  const level = cap > 0 ? tone(value / cap, 0.85) : 'none';

  return {
    name,
    width: `${cap > 0 ? Math.min(100, (value / cap) * 100) : 0}%`,
    used: `${cap > 0 && used != null ? Math.min(100, (used / cap) * 100) : 0}%`,
    fill: fills[level],
    color: toneText[level],
    label: `${format(value)} / ${format(cap)}`,
  };
}

const gibibytes = (bytes: number) => oneDecimal(bytes / GiB);

function sum(nodes: Node[], value: (node: Node) => number | null): number {
  return nodes.reduce((total, node) => total + (value(node) ?? 0), 0);
}

function shape(nodes: Node[]): string {
  const types = new Map<string, number>();

  for (const node of nodes) {
    types.set(node.instanceType, (types.get(node.instanceType) ?? 0) + 1);
  }

  return [...types].map(([type, number]) => `${number} × ${type.toUpperCase() || 'unknown'}`).join(', ');
}

function presentNode(node: Node) {
  const cut = node.name.lastIndexOf('-') + 1;
  const full = node.volumeLimit > 0 && node.attachedVolumes >= node.volumeLimit;

  return {
    name: node.name,
    prefix: node.name.slice(0, cut),
    suffix: node.name.slice(cut),
    age: age(node.createdAt),
    volumes: Array.from({ length: node.volumeLimit }, (_, index) => index < node.attachedVolumes),
    volumeLimit: node.volumeLimit,
    volumeLabel: node.volumeLimit ? `${node.attachedVolumes} / ${node.volumeLimit}` : count(node.attachedVolumes),
    full,
    meters: [
      meter('Memory', node.memoryRequested, node.memoryAllocatable, node.memoryUsed, gibibytes),
      meter('CPU', node.cpuRequested, node.cpuAllocatable, node.cpuUsed, oneDecimal),
      meter('Pods', node.pods, node.podLimit, null, count),
    ],
  };
}

const columns = computed(() => (props.pools ?? []).map((pool) => ({
  name: pool.name || 'No pool',
  shape: shape(pool.nodes),
  summary: [
    meter('Volumes', sum(pool.nodes, (node) => node.attachedVolumes), sum(pool.nodes, (node) => node.volumeLimit), null, count),
    meter('Memory', sum(pool.nodes, (node) => node.memoryRequested), sum(pool.nodes, (node) => node.memoryAllocatable), sum(pool.nodes, (node) => node.memoryUsed), gibibytes),
    meter('CPU', sum(pool.nodes, (node) => node.cpuRequested), sum(pool.nodes, (node) => node.cpuAllocatable), sum(pool.nodes, (node) => node.cpuUsed), oneDecimal),
    meter('Pods', sum(pool.nodes, (node) => node.pods), sum(pool.nodes, (node) => node.podLimit), null, count),
  ],
  nodes: pool.nodes.map(presentNode),
})));

const nodeCount = computed(() => (props.pools ?? []).reduce((total, pool) => total + pool.nodes.length, 0));

const grid = computed(() => ({
  gridTemplateColumns: `repeat(${Math.max(columns.value.length, 1)}, minmax(260px, 1fr))`,
  minWidth: `${Math.max(columns.value.length, 1) * 260}px`,
}));
</script>

<template>
  <section aria-labelledby="pools-title" class="mt-4 border-t">
    <div class="flex min-h-11 flex-wrap items-center gap-x-4 gap-y-2 px-4">
      <h2 id="pools-title" class="font-semibold">
        <button type="button" :aria-expanded="open" class="-ml-1 flex items-center gap-1.5 rounded-lg py-0.5 pr-1.5 pl-1 font-semibold hover:bg-muted" @click="open = !open">
          <ChevronRight class="size-4 text-muted-foreground transition-transform" :class="open ? 'rotate-90' : ''" aria-hidden="true" />
          Node pools
        </button>
      </h2>
      <span v-if="pools" class="text-muted-foreground">{{ count(nodeCount) }} {{ nodeCount === 1 ? 'node' : 'nodes' }}</span>
      <span class="ml-auto flex items-center gap-1.5 text-muted-foreground">
        <span class="size-2.5 rounded-[3px] bg-foreground/20" aria-hidden="true" />Requested
      </span>
      <span class="flex items-center gap-1.5 text-muted-foreground">
        <span class="size-2.5 rounded-[3px] bg-foreground/62" aria-hidden="true" />In use
      </span>
    </div>

    <p v-if="error" class="border-t border-hairline px-4 py-3 text-status-danger">{{ error }}</p>
    <div v-else-if="!pools" class="h-40 animate-pulse border-t border-hairline bg-sidebar" />

    <div v-else class="overflow-x-auto">
      <div class="grid border-t border-hairline tabular-nums" :style="grid">
        <div
          v-for="(column, index) in columns"
          :key="column.name"
          class="flex flex-col gap-2.5 border-b border-hairline bg-sidebar px-4 py-3.5"
          :class="index < columns.length - 1 ? 'border-r' : ''"
        >
          <span class="flex items-baseline gap-2 whitespace-nowrap">
            <span class="font-semibold">{{ column.name }}</span>
            <span class="truncate text-muted-foreground">{{ column.shape }}</span>
          </span>
          <span v-for="item in column.summary" :key="item.name" class="grid grid-cols-[64px_minmax(0,1fr)_92px] items-center gap-2.5">
            <span class="whitespace-nowrap text-muted-foreground">{{ item.name }}</span>
            <span class="relative h-1.5 overflow-hidden rounded-full bg-muted shadow-[inset_0_1px_1px_oklch(0.30_0.04_55/0.12)]">
              <span class="absolute inset-y-0 left-0 rounded-full" :class="item.fill" :style="{ width: item.width }" />
              <span class="absolute inset-y-0 left-0 rounded-full bg-foreground/62" :style="{ width: item.used }" />
            </span>
            <span class="text-right font-semibold whitespace-nowrap" :class="item.color">{{ item.label }}</span>
          </span>
        </div>

        <div
          v-for="(column, index) in columns"
          v-show="open"
          :key="`${column.name}-nodes`"
          class="flex min-w-0 flex-col border-hairline"
          :class="index < columns.length - 1 ? 'border-r' : ''"
        >
          <div v-for="node in column.nodes" :key="node.name" class="flex flex-col gap-2.5 border-b border-hairline px-4 py-3.5">
            <span class="flex items-baseline gap-2 whitespace-nowrap">
              <span class="flex min-w-0 font-mono">
                <span class="truncate text-muted-foreground">{{ node.prefix }}</span>
                <span class="shrink-0 font-medium">{{ node.suffix }}</span>
              </span>
              <span class="ml-auto text-muted-foreground">{{ node.age }}</span>
            </span>
            <span class="grid grid-cols-[64px_minmax(0,1fr)_92px] items-center gap-2.5">
              <span class="whitespace-nowrap text-muted-foreground">Volumes</span>
              <span class="grid gap-px" :style="{ gridTemplateColumns: `repeat(${node.volumeLimit || 1}, minmax(0, 1fr))` }">
                <span
                  v-for="(attached, index) in node.volumes"
                  :key="index"
                  class="h-3 rounded-[3px]"
                  :class="!attached
                    ? 'bg-muted shadow-[inset_0_1px_1px_oklch(0.30_0.04_55/0.12)]'
                    : node.full
                      ? 'bg-linear-to-b from-[oklch(0.58_0.19_25)] to-[oklch(0.47_0.19_25)] shadow-[inset_0_1px_0_oklch(1_0_0/0.25)]'
                      : 'bg-linear-to-b from-foreground/45 to-foreground/60 shadow-[inset_0_1px_0_oklch(1_0_0/0.25)]'"
                />
              </span>
              <span class="text-right font-medium whitespace-nowrap" :class="node.full ? 'text-status-danger' : ''">{{ node.volumeLabel }}</span>
            </span>
            <span v-for="item in node.meters" :key="item.name" class="grid grid-cols-[64px_minmax(0,1fr)_92px] items-center gap-2.5">
              <span class="whitespace-nowrap text-muted-foreground">{{ item.name }}</span>
              <span class="relative h-1.5 overflow-hidden rounded-full bg-muted shadow-[inset_0_1px_1px_oklch(0.30_0.04_55/0.12)]">
                <span class="absolute inset-y-0 left-0 rounded-full" :class="item.fill" :style="{ width: item.width }" />
                <span class="absolute inset-y-0 left-0 rounded-full bg-foreground/62" :style="{ width: item.used }" />
              </span>
              <span class="text-right whitespace-nowrap" :class="item.color">{{ item.label }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
