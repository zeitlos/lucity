<script setup lang="ts">
import { computed, ref } from 'vue';
import { useElementSize } from '@vueuse/core';

type Point = { timestamp: string; value?: number | null };
type Series = { label?: string; points: Point[] };
type Seg = { time: number; value: number }[];

const props = withDefaults(
  defineProps<{
    series: Series[];
    from: number;
    to: number;
    max?: number | null;
    height?: number;
    formatValue?: (value: number) => string;
  }>(),
  {
    max: null,
    height: 180,
    formatValue: (value: number) => String(value),
  },
);

const palette = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-3)', 'var(--chart-4)', 'var(--chart-5)'];

const padding = { top: 12, right: 12, bottom: 22, left: 56 };

const container = ref<HTMLElement | null>(null);
const { width: measured } = useElementSize(container);

const width = computed(() => Math.round(measured.value) || 600);
const innerWidth = computed(() => width.value - padding.left - padding.right);
const innerHeight = computed(() => props.height - padding.top - padding.bottom);

const multi = computed(() => props.series.length > 1);

function colorFor(index: number): string {
  return multi.value ? palette[index % palette.length]! : 'var(--primary)';
}

const allValues = computed(() =>
  props.series.flatMap(s => s.points.map(p => p.value).filter((v): v is number => v != null)),
);

const yMax = computed(() => {
  const candidate = Math.max(props.max ?? 0, allValues.value.length ? Math.max(...allValues.value) : 0);
  return candidate > 0 ? candidate * 1.1 : 1;
});

function xFor(time: number): number {
  const span = props.to - props.from || 1;
  const fraction = Math.min(1, Math.max(0, (time - props.from) / span));
  return padding.left + fraction * innerWidth.value;
}

function yFor(value: number): number {
  return padding.top + innerHeight.value - (value / yMax.value) * innerHeight.value;
}

function segmentsOf(points: Point[]): Seg[] {
  const result: Seg[] = [];
  let current: Seg = [];

  for (const p of points) {
    if (p.value == null) {
      if (current.length) result.push(current);
      current = [];
      continue;
    }
    current.push({ time: new Date(p.timestamp).getTime(), value: p.value });
  }

  if (current.length) result.push(current);
  return result;
}

const rendered = computed(() =>
  props.series.map((s, i) => {
    const segs = segmentsOf(s.points);
    const lineSegs = segs.filter(seg => seg.length >= 2);

    return {
      color: colorFor(i),
      lines: lineSegs.map(seg => seg.map((p, j) => `${j === 0 ? 'M' : 'L'} ${xFor(p.time)} ${yFor(p.value)}`).join(' ')),
      areas: multi.value
        ? []
        : lineSegs.map(seg => {
            const first = seg[0]!;
            const last = seg[seg.length - 1]!;
            const line = seg.map(p => `L ${xFor(p.time)} ${yFor(p.value)}`).join(' ');
            return `M ${xFor(first.time)} ${yFor(0)} ${line} L ${xFor(last.time)} ${yFor(0)} Z`;
          }),
      dots: segs.filter(seg => seg.length === 1).map(seg => seg[0]!),
    };
  }),
);

const gridLines = computed(() => {
  const lines: { y: number; label: string }[] = [];
  for (let i = 0; i <= 2; i++) {
    const value = (yMax.value / 2) * i;
    lines.push({ y: yFor(value), label: props.formatValue(value) });
  }
  return lines;
});

function formatTime(time: number): string {
  const spanHours = (props.to - props.from) / 3_600_000;
  return new Date(time).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: spanHours <= 48 ? '2-digit' : undefined,
    minute: spanHours <= 48 ? '2-digit' : undefined,
  });
}

const xLabels = computed(() => [
  { x: xFor(props.from), label: formatTime(props.from), anchor: 'start' },
  { x: xFor(props.to), label: formatTime(props.to), anchor: 'end' },
]);

const hoverTime = ref<number | null>(null);

const hoverEntries = computed(() => {
  if (hoverTime.value == null) return [];
  const cursor = hoverTime.value;

  return props.series
    .map((s, i) => {
      let nearest: { time: number; value: number } | null = null;
      let best = Infinity;
      for (const p of s.points) {
        if (p.value == null) continue;
        const time = new Date(p.timestamp).getTime();
        const distance = Math.abs(time - cursor);
        if (distance < best) {
          best = distance;
          nearest = { time, value: p.value };
        }
      }
      return nearest ? { label: s.label, color: colorFor(i), ...nearest } : null;
    })
    .filter((e): e is NonNullable<typeof e> => e != null);
});

const crosshairX = computed(() => (hoverEntries.value.length ? xFor(hoverEntries.value[0]!.time) : 0));

function onMove(event: MouseEvent) {
  const rect = (event.currentTarget as SVGElement).getBoundingClientRect();
  const fraction = (event.clientX - rect.left) / rect.width;
  hoverTime.value = props.from + Math.min(1, Math.max(0, fraction)) * (props.to - props.from);
}

function onLeave() {
  hoverTime.value = null;
}
</script>

<template>
  <div ref="container" class="relative w-full">
    <svg
      :viewBox="`0 0 ${width} ${height}`"
      :width="width"
      :height="height"
      class="max-w-full"
      @mousemove="onMove"
      @mouseleave="onLeave"
    >
      <g class="stroke-border" stroke-width="1">
        <line
          v-for="line in gridLines"
          :key="`grid-${line.y}`"
          :x1="padding.left"
          :x2="width - padding.right"
          :y1="line.y"
          :y2="line.y"
        />
      </g>

      <text
        v-for="line in gridLines"
        :key="`ylabel-${line.y}`"
        :x="padding.left - 8"
        :y="line.y + 3"
        text-anchor="end"
        class="fill-muted-foreground"
        style="font-size: 10px"
      >
        {{ line.label }}
      </text>

      <template v-for="(s, i) in rendered" :key="`series-${i}`">
        <path
          v-for="(area, j) in s.areas"
          :key="`area-${i}-${j}`"
          :d="area"
          :fill="s.color"
          fill-opacity="0.1"
          stroke="none"
        />
        <path
          v-for="(line, j) in s.lines"
          :key="`line-${i}-${j}`"
          :d="line"
          :stroke="s.color"
          fill="none"
          stroke-width="2"
          stroke-linejoin="round"
        />
        <circle
          v-for="(dot, j) in s.dots"
          :key="`dot-${i}-${j}`"
          :cx="xFor(dot.time)"
          :cy="yFor(dot.value)"
          r="2.5"
          :fill="s.color"
        />
      </template>

      <g v-if="hoverEntries.length">
        <line
          :x1="crosshairX"
          :x2="crosshairX"
          :y1="padding.top"
          :y2="padding.top + innerHeight"
          class="stroke-muted-foreground/40"
          stroke-width="1"
        />
        <circle
          v-for="(entry, i) in hoverEntries"
          :key="`hover-${i}`"
          :cx="xFor(entry.time)"
          :cy="yFor(entry.value)"
          r="3"
          :fill="entry.color"
        />
      </g>

      <text
        v-for="label in xLabels"
        :key="`xlabel-${label.x}`"
        :x="label.x"
        :y="height - 6"
        :text-anchor="label.anchor"
        class="fill-muted-foreground"
        style="font-size: 10px"
      >
        {{ label.label }}
      </text>
    </svg>

    <div v-if="multi" class="mt-2 flex flex-wrap gap-x-3 gap-y-1 px-1">
      <div
        v-for="(s, i) in series"
        :key="`legend-${i}`"
        class="flex items-center gap-1.5 text-xs text-muted-foreground"
      >
        <span class="inline-block h-2 w-2 rounded-full" :style="{ background: colorFor(i) }"></span>
        {{ s.label }}
      </div>
    </div>

    <div
      v-if="hoverEntries.length"
      class="pointer-events-none absolute -translate-x-1/2 rounded-md border bg-popover px-2 py-1 text-xs shadow-sm"
      :style="{ left: `${(crosshairX / width) * 100}%`, top: '0' }"
    >
      <div class="mb-0.5 text-muted-foreground">{{ formatTime(hoverEntries[0]!.time) }}</div>
      <div v-for="(entry, i) in hoverEntries" :key="`tip-${i}`" class="flex items-center gap-1.5">
        <span class="inline-block h-2 w-2 rounded-full" :style="{ background: entry.color }"></span>
        <span v-if="entry.label" class="text-muted-foreground">{{ entry.label }}</span>
        <span class="ml-auto font-medium text-foreground">{{ formatValue(entry.value) }}</span>
      </div>
    </div>
  </div>
</template>
