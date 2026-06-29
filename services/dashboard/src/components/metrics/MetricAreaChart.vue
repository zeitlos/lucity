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
    markers?: number[];
    stacked?: boolean;
    height?: number;
    formatValue?: (value: number) => string;
  }>(),
  {
    max: null,
    markers: () => [],
    stacked: false,
    height: 280,
    formatValue: (value: number) => String(value),
  },
);

const visibleMarkers = computed(() => props.markers.filter(m => m >= props.from && m <= props.to));

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
  if (props.max != null && props.max > 0) return props.max;
  const candidate = allValues.value.length ? Math.max(...allValues.value) : 0;
  return candidate > 0 ? candidate * 1.1 : 1;
});

function xFor(time: number): number {
  const span = props.to - props.from || 1;
  const fraction = Math.min(1, Math.max(0, (time - props.from) / span));
  return padding.left + fraction * innerWidth.value;
}

function yFor(value: number): number {
  const y = padding.top + innerHeight.value - (value / yMax.value) * innerHeight.value;
  return Math.min(padding.top + innerHeight.value, Math.max(padding.top, y));
}

const stackedBands = computed(() => {
  if (!props.stacked) return [];

  const times = new Set<number>();
  const lookups = props.series.map(s => {
    const m = new Map<number, number>();
    for (const p of s.points) {
      if (p.value == null) continue;
      const time = new Date(p.timestamp).getTime();
      m.set(time, p.value);
      times.add(time);
    }
    return m;
  });

  const sorted = [...times].sort((a, b) => a - b);
  const runningBottom = sorted.map(() => 0);

  return props.series.map((s, i) => {
    const lookup = lookups[i]!;
    const segments: { x: number; topY: number; bottomY: number }[][] = [];
    let current: { x: number; topY: number; bottomY: number }[] = [];

    sorted.forEach((time, ti) => {
      const bottom = runningBottom[ti]!;
      if (lookup.has(time)) {
        const top = bottom + lookup.get(time)!;
        current.push({ x: xFor(time), topY: yFor(top), bottomY: yFor(bottom) });
        runningBottom[ti] = top;
      } else if (current.length) {
        segments.push(current);
        current = [];
      }
    });
    if (current.length) segments.push(current);

    return {
      color: colorFor(i),
      areas: segments.map(seg => {
        const top = seg.map((p, j) => `${j === 0 ? 'M' : 'L'} ${p.x} ${p.topY}`).join(' ');
        const bottom = seg
          .slice()
          .reverse()
          .map(p => `L ${p.x} ${p.bottomY}`)
          .join(' ');
        return `${top} ${bottom} Z`;
      }),
      lines: segments.map(seg => seg.map((p, j) => `${j === 0 ? 'M' : 'L'} ${p.x} ${p.topY}`).join(' ')),
    };
  });
});

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

const hoverTime = defineModel<number | null>('cursor', { default: null });

const focusTime = computed(() => {
  if (hoverTime.value == null) return null;
  const cursor = hoverTime.value;

  let best = Infinity;
  let focus: number | null = null;
  for (const s of props.series) {
    for (const p of s.points) {
      if (p.value == null) continue;
      const time = new Date(p.timestamp).getTime();
      const distance = Math.abs(time - cursor);
      if (distance < best) {
        best = distance;
        focus = time;
      }
    }
  }
  return focus;
});

const hoverEntries = computed(() => {
  if (focusTime.value == null) return [];
  const focus = focusTime.value;

  return props.series
    .map((s, i) => {
      const point = s.points.find(p => p.value != null && new Date(p.timestamp).getTime() === focus);
      if (!point || point.value == null) return null;
      return { label: s.label, color: colorFor(i), time: focus, value: point.value };
    })
    .filter((e): e is NonNullable<typeof e> => e != null);
});

const crosshairX = computed(() => (focusTime.value == null ? 0 : xFor(focusTime.value)));

const tooltipTop = computed(() => {
  if (!hoverEntries.value.length) return 0;
  if (props.stacked) return yFor(0) + 8;
  return Math.max(...hoverEntries.value.map(e => yFor(e.value))) + 8;
});

const hoveredMarker = ref<number | null>(null);

function onMove(event: MouseEvent) {
  const rect = (event.currentTarget as SVGElement).getBoundingClientRect();
  if (rect.width === 0) return;
  const viewBoxX = ((event.clientX - rect.left) / rect.width) * width.value;
  const fraction = (viewBoxX - padding.left) / innerWidth.value;
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

      <g v-for="(marker, i) in visibleMarkers" :key="`marker-${i}`">
        <line
          :x1="xFor(marker)"
          :x2="xFor(marker)"
          :y1="padding.top"
          :y2="padding.top + innerHeight"
          class="stroke-muted-foreground/30"
          stroke-width="1"
          stroke-dasharray="3 3"
        />
        <path
          :d="`M ${xFor(marker) - 3} ${padding.top} L ${xFor(marker) + 3} ${padding.top} L ${xFor(marker)} ${padding.top + 5} Z`"
          class="fill-muted-foreground/60"
        />
      </g>

      <template v-if="stacked">
        <template v-for="(band, i) in stackedBands" :key="`band-${i}`">
          <path
            v-for="(area, j) in band.areas"
            :key="`band-area-${i}-${j}`"
            :d="area"
            :fill="band.color"
            fill-opacity="0.3"
            stroke="none"
          />
          <path
            v-for="(line, j) in band.lines"
            :key="`band-line-${i}-${j}`"
            :d="line"
            :stroke="band.color"
            fill="none"
            stroke-width="1.5"
            stroke-linejoin="round"
          />
        </template>
      </template>

      <template v-else>
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
          v-show="!stacked"
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

      <rect
        v-for="(marker, i) in visibleMarkers"
        :key="`marker-hit-${i}`"
        :x="xFor(marker) - 5"
        :y="padding.top"
        width="10"
        :height="innerHeight"
        fill="transparent"
        style="cursor: pointer"
        @mouseenter="hoveredMarker = marker"
        @mouseleave="hoveredMarker = null"
      />
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
      class="pointer-events-none absolute z-10 w-max -translate-x-1/2 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-xs shadow-sm"
      :style="{ left: `${(crosshairX / width) * 100}%`, top: `${tooltipTop}px` }"
    >
      <div class="mb-0.5 text-muted-foreground">{{ formatTime(focusTime!) }}</div>
      <div v-for="(entry, i) in hoverEntries" :key="`tip-${i}`" class="flex items-center gap-1.5">
        <span v-if="multi" class="inline-block h-2 w-2 rounded-full" :style="{ background: entry.color }"></span>
        <span v-if="entry.label" class="text-muted-foreground">{{ entry.label }}</span>
        <span class="ml-auto font-medium text-foreground">{{ formatValue(entry.value) }}</span>
      </div>
    </div>

    <div
      v-if="hoveredMarker != null"
      class="pointer-events-none absolute top-0 z-20 w-max -translate-x-1/2 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-xs text-muted-foreground shadow-sm"
      :style="{ left: `${(xFor(hoveredMarker) / width) * 100}%` }"
    >
      Deployed {{ formatTime(hoveredMarker) }}
    </div>
  </div>
</template>
