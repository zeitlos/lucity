<script setup lang="ts">
import { computed } from 'vue';

type Point = { timestamp: string; value?: number | null };
type Series = { color: string; points: Point[] };
type Sample = { time: number; value: number };

const props = withDefaults(
  defineProps<{
    series: Series[];
    from: number;
    to: number;
    min?: number;
    max?: number | null;
    height?: number;
    strokeWidth?: number;
    area?: boolean;
    limit?: boolean;
    grid?: boolean;
  }>(),
  {
    min: 0,
    max: null,
    height: 32,
    strokeWidth: 1.5,
    area: false,
    limit: false,
    grid: false,
  },
);

const cursor = defineModel<number | null>('cursor', { default: null });

const width = 1000;

const samples = computed(() =>
  props.series.map(s =>
    s.points
      .filter((p): p is Point & { value: number } => p.value != null)
      .map(p => ({ time: new Date(p.timestamp).getTime(), value: p.value })),
  ),
);

const top = computed(() => {
  if (props.max != null && props.max > props.min) return props.max;
  const values = samples.value.flat().map(s => s.value);
  const peak = values.length ? Math.max(...values) : props.min;
  return peak > props.min ? props.min + (peak - props.min) * 1.2 : props.min + 1;
});

function x(time: number): number {
  const span = props.to - props.from || 1;
  return ((time - props.from) / span) * width;
}

function y(value: number): number {
  const fraction = (value - props.min) / (top.value - props.min);
  return 100 - Math.min(1, Math.max(0, fraction)) * 100;
}

function segments(points: Sample[]): Sample[][] {
  let step = Infinity;
  for (let i = 1; i < points.length; i++) {
    step = Math.min(step, points[i]!.time - points[i - 1]!.time);
  }

  const result: Sample[][] = [];
  let current: Sample[] = [];

  for (const point of points) {
    const previous = current[current.length - 1];
    if (previous && point.time - previous.time > step * 2.5) {
      result.push(current);
      current = [];
    }
    current.push(point);
  }

  if (current.length) result.push(current);
  return result;
}

function linePath(segment: Sample[]): string {
  const path = segment.map((p, i) => `${i ? 'L' : 'M'}${x(p.time).toFixed(1)} ${y(p.value).toFixed(2)}`).join(' ');
  if (segment.length > 1) return path;
  const only = segment[0]!;
  return `${path} L${x(only.time).toFixed(1)} ${y(only.value).toFixed(2)}`;
}

const paths = computed(() =>
  samples.value.map((points, i) => {
    const parts = segments(points);
    return {
      color: props.series[i]!.color,
      lines: parts.map(linePath),
      areas: props.area
        ? parts
            .filter(part => part.length > 1)
            .map(part => {
              const first = part[0]!;
              const last = part[part.length - 1]!;
              return `${linePath(part)} L${x(last.time).toFixed(1)} 100 L${x(first.time).toFixed(1)} 100 Z`;
            })
        : [],
    };
  }),
);

const focus = computed(() => {
  if (cursor.value == null) return null;
  const target = cursor.value;

  let best = Infinity;
  let time: number | null = null;
  for (const point of samples.value.flat()) {
    const distance = Math.abs(point.time - target);
    if (distance < best) {
      best = distance;
      time = point.time;
    }
  }

  if (time == null) return null;

  const dots = samples.value
    .map((points, i) => {
      const point = points.find(p => p.time === time);
      return point ? { x: x(point.time), y: y(point.value), color: props.series[i]!.color } : null;
    })
    .filter((dot): dot is NonNullable<typeof dot> => dot != null);

  return { x: x(time), dots };
});

function onMove(event: MouseEvent) {
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
  if (rect.width === 0) return;
  const fraction = Math.min(1, Math.max(0, (event.clientX - rect.left) / rect.width));
  cursor.value = props.from + fraction * (props.to - props.from);
}

function onLeave() {
  cursor.value = null;
}
</script>

<template>
  <div
    class="relative w-full"
    :style="{ height: `${height}px` }"
    @mousemove="onMove"
    @mouseleave="onLeave"
  >
    <div v-if="limit" class="pointer-events-none absolute inset-x-0 top-0 border-t border-dashed border-border" />
    <div v-if="grid" class="pointer-events-none absolute inset-x-0 top-1/2 border-t border-border/50" />
    <div v-if="grid" class="pointer-events-none absolute inset-x-0 bottom-0 border-t border-border" />

    <svg
      :viewBox="`0 0 ${width} 100`"
      preserveAspectRatio="none"
      class="absolute inset-0 h-full w-full overflow-visible"
      aria-hidden="true"
    >
      <template v-for="(path, i) in paths" :key="i">
        <path
          v-for="(d, j) in path.areas"
          :key="`area-${j}`"
          :d="d"
          :fill="path.color"
          fill-opacity="0.12"
          stroke="none"
        />
        <path
          v-for="(d, j) in path.lines"
          :key="`line-${j}`"
          :d="d"
          fill="none"
          :stroke="path.color"
          :stroke-width="strokeWidth"
          stroke-linejoin="round"
          stroke-linecap="round"
          vector-effect="non-scaling-stroke"
        />
      </template>

      <template v-if="focus">
        <line
          :x1="focus.x"
          :x2="focus.x"
          y1="0"
          y2="100"
          class="stroke-muted-foreground/50"
          stroke-width="1"
          vector-effect="non-scaling-stroke"
        />
        <path
          v-for="(dot, i) in focus.dots"
          :key="i"
          :d="`M${dot.x} ${dot.y} L${dot.x} ${dot.y}`"
          :stroke="dot.color"
          :stroke-width="strokeWidth * 3.5"
          stroke-linecap="round"
          vector-effect="non-scaling-stroke"
        />
      </template>
    </svg>
  </div>
</template>
