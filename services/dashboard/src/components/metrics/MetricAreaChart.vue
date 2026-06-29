<script setup lang="ts">
import { computed, ref } from 'vue';
import { useElementSize } from '@vueuse/core';

type Point = { timestamp: string; value?: number | null };

const props = withDefaults(
  defineProps<{
    points: Point[];
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

const padding = { top: 12, right: 12, bottom: 22, left: 56 };

const container = ref<HTMLElement | null>(null);
const { width: measured } = useElementSize(container);

const width = computed(() => Math.round(measured.value) || 600);
const innerWidth = computed(() => width.value - padding.left - padding.right);
const innerHeight = computed(() => props.height - padding.top - padding.bottom);

const samples = computed(() =>
  props.points.map(p => ({ time: new Date(p.timestamp).getTime(), value: p.value })),
);

const values = computed(() =>
  samples.value.map(s => s.value).filter((v): v is number => v != null),
);

const dataMax = computed(() => (values.value.length ? Math.max(...values.value) : 0));

const yMax = computed(() => {
  const candidate = Math.max(props.max ?? 0, dataMax.value);
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

const segments = computed(() => {
  const result: { time: number; value: number }[][] = [];
  let current: { time: number; value: number }[] = [];

  samples.value.forEach(sample => {
    if (sample.value == null) {
      if (current.length) result.push(current);
      current = [];
      return;
    }
    current.push({ time: sample.time, value: sample.value });
  });

  if (current.length) result.push(current);
  return result;
});

const lineSegments = computed(() => segments.value.filter(s => s.length >= 2));
const dots = computed(() => segments.value.filter(s => s.length === 1).map(s => s[0]!));

const linePaths = computed(() =>
  lineSegments.value.map(segment =>
    segment.map((p, i) => `${i === 0 ? 'M' : 'L'} ${xFor(p.time)} ${yFor(p.value)}`).join(' '),
  ),
);

const areaPaths = computed(() =>
  lineSegments.value.map(segment => {
    const baseline = yFor(0);
    const first = segment[0]!;
    const last = segment[segment.length - 1]!;
    const line = segment.map(p => `L ${xFor(p.time)} ${yFor(p.value)}`).join(' ');
    return `M ${xFor(first.time)} ${baseline} ${line} L ${xFor(last.time)} ${baseline} Z`;
  }),
);

const ceilingY = computed(() => (props.max && props.max > 0 ? yFor(props.max) : null));

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

const hoverPoint = computed(() => {
  if (hoverTime.value == null) return null;
  let nearest: { time: number; value: number } | null = null;
  let best = Infinity;
  for (const segment of segments.value) {
    for (const p of segment) {
      const distance = Math.abs(p.time - hoverTime.value);
      if (distance < best) {
        best = distance;
        nearest = p;
      }
    }
  }
  return nearest;
});

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

      <path
        v-for="(area, i) in areaPaths"
        :key="`area-${i}`"
        :d="area"
        class="fill-primary/10"
        stroke="none"
      />
      <path
        v-for="(line, i) in linePaths"
        :key="`line-${i}`"
        :d="line"
        class="stroke-primary"
        fill="none"
        stroke-width="2"
        stroke-linejoin="round"
      />
      <circle
        v-for="(dot, i) in dots"
        :key="`dot-${i}`"
        :cx="xFor(dot.time)"
        :cy="yFor(dot.value)"
        r="2.5"
        class="fill-primary"
      />

      <line
        v-if="ceilingY != null"
        :x1="padding.left"
        :x2="width - padding.right"
        :y1="ceilingY"
        :y2="ceilingY"
        class="stroke-destructive/60"
        stroke-width="1"
        stroke-dasharray="4 4"
      />

      <g v-if="hoverPoint">
        <line
          :x1="xFor(hoverPoint.time)"
          :x2="xFor(hoverPoint.time)"
          :y1="padding.top"
          :y2="padding.top + innerHeight"
          class="stroke-muted-foreground/40"
          stroke-width="1"
        />
        <circle
          :cx="xFor(hoverPoint.time)"
          :cy="yFor(hoverPoint.value)"
          r="3"
          class="fill-primary"
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

    <div
      v-if="hoverPoint"
      class="pointer-events-none absolute -translate-x-1/2 rounded-md border bg-popover px-2 py-1 text-xs shadow-sm"
      :style="{ left: `${(xFor(hoverPoint.time) / width) * 100}%`, top: '0' }"
    >
      <div class="font-medium text-foreground">{{ formatValue(hoverPoint.value) }}</div>
      <div class="text-muted-foreground">{{ formatTime(hoverPoint.time) }}</div>
    </div>
  </div>
</template>
