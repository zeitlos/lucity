<script setup lang="ts">
import { computed } from 'vue';

import type { Bar, Layer, Tick } from '@/lib/trends';

const props = defineProps<{
  bars: Bar[];
  layers: Layer[];
  ticks: Tick[];
  label: string;
  hovered: number | null;
  tip: boolean;
  message?: string;
  failed?: boolean;
}>();

const emit = defineEmits<{
  hover: [index: number | null];
}>();

const WIDTH = 1000;
const HEIGHT = 160;
const scale = [
  2, 4, 6, 8, 10, 16, 20, 30, 40, 50, 60, 80, 100, 120, 160, 200, 300, 400, 600, 800,
  1000, 1200, 1500, 2000, 2500, 3000, 4000, 5000, 6000, 8000, 10000,
];

function f(value: number): string {
  return value.toFixed(1);
}

function segment(x: number, width: number, bottom: number, upper: number, round: boolean): string {
  const radius = round ? Math.min(2, width / 2, bottom - upper) : 0;
  const corners = radius
    ? `V${f(upper + radius)}Q${f(x)} ${f(upper)} ${f(x + radius)} ${f(upper)}H${f(x + width - radius)}Q${f(x + width)} ${f(upper)} ${f(x + width)} ${f(upper + radius)}`
    : `V${f(upper)}H${f(x + width)}`;

  return `M${f(x)} ${f(bottom)}${corners}V${f(bottom)}Z`;
}

const chart = computed(() => {
  const slot = WIDTH / props.bars.length;
  const width = Math.min(slot * 0.78, 56);
  const totals = props.bars.map((bar) => bar.values.reduce((sum, value) => sum + value, 0));
  const peak = Math.max(1, ...totals);
  const top = scale.find((step) => step >= peak) ?? peak;
  const y = (value: number) => HEIGHT - (value / top) * HEIGHT;
  const layers = props.layers.map((layer) => ({ ...layer, complete: '', partial: '' }));

  props.bars.forEach((bar, index) => {
    const x = index * slot + (slot - width) / 2;
    const last = bar.values.reduce((found, value, layer) => (value > 0 ? layer : found), -1);
    let base = 0;

    bar.values.forEach((value, layer) => {
      const target = layers[layer];

      if (value <= 0 || !target) return;

      const bottom = y(base) - (base > 0 ? 1.5 : 0);
      const path = segment(x, width, bottom, Math.min(y(base + value), bottom - 1), layer === last);

      if (bar.partial) target.partial += path;
      else target.complete += path;

      base += value;
    });
  });

  return { slot, top, y, totals, layers };
});

const band = computed(() => {
  if (props.hovered == null || props.hovered >= props.bars.length) return null;

  const x = chart.value.slot * props.hovered;
  return `M${f(x)} 0H${f(x + chart.value.slot)}V${HEIGHT}H${f(x)}Z`;
});

const tooltip = computed(() => {
  const index = props.hovered;
  const bar = index == null ? undefined : props.bars[index];

  if (!props.tip || index == null || !bar) return null;

  const center = ((index + 0.5) / props.bars.length) * 100;
  const height = 40 + 18 * props.layers.length;
  const barTop = chart.value.y(chart.value.totals[index] ?? 0);
  const beside = barTop < height;

  return {
    title: bar.title,
    rows: props.layers
      .map((layer, layerIndex) => ({ label: layer.label, swatch: layer.swatch, value: bar.values[layerIndex] ?? 0 }))
      .reverse(),
    style: beside
      ? { top: '8px', left: `${center}%`, transform: center > 70 ? 'translateX(calc(-100% - 12px))' : 'translateX(12px)' }
      : { top: `${barTop - height}px`, left: `${Math.min(Math.max(center, 10), 90)}%`, transform: 'translateX(-50%)' },
  };
});

function hover(event: MouseEvent) {
  const box = (event.currentTarget as HTMLElement).getBoundingClientRect();
  const index = Math.floor(((event.clientX - box.left) / box.width) * props.bars.length);

  emit('hover', Math.min(Math.max(index, 0), props.bars.length - 1));
}
</script>

<template>
  <div class="grid min-w-0 grid-cols-[36px_minmax(0,1fr)] gap-x-2 gap-y-1.5">
    <div class="relative h-40 text-[11px] leading-3.5 text-muted-foreground tabular-nums">
      <span class="absolute -top-[7px] right-0">{{ chart.top }}</span>
      <span class="absolute top-[73px] right-0">{{ chart.top / 2 }}</span>
      <span class="absolute right-0 -bottom-[7px]">0</span>
    </div>
    <div class="relative h-40" @mousemove="hover" @mouseleave="emit('hover', null)">
      <svg viewBox="0 0 1000 160" preserveAspectRatio="none" class="block h-40 w-full overflow-visible" role="img" :aria-label="label">
        <path v-if="band" :d="band" class="fill-foreground/5" />
        <line x1="0" y1="0.5" x2="1000" y2="0.5" class="stroke-hairline" vector-effect="non-scaling-stroke" />
        <line x1="0" y1="80" x2="1000" y2="80" class="stroke-hairline" vector-effect="non-scaling-stroke" />
        <template v-for="layer in chart.layers" :key="layer.label">
          <path :d="layer.complete" :class="layer.fill" />
          <path :d="layer.partial" :class="layer.partialFill" />
        </template>
        <line x1="0" y1="159.5" x2="1000" y2="159.5" class="stroke-border" vector-effect="non-scaling-stroke" />
      </svg>
      <p
        v-if="message"
        class="absolute inset-0 flex items-center justify-center"
        :class="failed ? 'text-status-danger' : 'text-muted-foreground'"
      >
        {{ message }}
      </p>
      <div
        v-if="tooltip"
        class="pointer-events-none absolute rounded-sm border bg-card px-2.5 py-1.5 text-xs whitespace-nowrap shadow-lg"
        :style="tooltip.style"
      >
        <p class="font-medium">{{ tooltip.title }}</p>
        <p v-for="row in tooltip.rows" :key="row.label" class="flex items-center gap-2">
          <span class="size-2 rounded-[2px]" :class="row.swatch" />
          <span class="text-muted-foreground">{{ row.label }}</span>
          <span class="ml-auto pl-4 tabular-nums">{{ row.value }}</span>
        </p>
      </div>
    </div>
    <span />
    <div class="relative h-4 text-xs text-muted-foreground">
      <span
        v-for="tick in ticks"
        :key="tick.left"
        class="absolute whitespace-nowrap"
        :style="{ left: tick.left, transform: tick.transform }"
      >{{ tick.text }}</span>
    </div>
  </div>
</template>
