<script setup lang="ts">
import { computed, ref } from 'vue';
import { useEventListener } from '@vueuse/core';
import { formatTime } from '@/lib/format';
import { darkTooltipClass } from '@/lib/tooltip';
import { dayDate, formatCents, formatDay, formatDayTime, sum } from '@/lib/usage';
import { cn } from '@/lib/utils';

defineOptions({ inheritAttrs: false });

const props = defineProps<{
  label: string;
  sources: { key: string; name: string; color?: string; daily: number[] }[];
  periodStart: string;
  measuredUntil: string;
  days: number;
  lastDay: number;
  running: boolean;
  peak: number;
}>();

const root = ref<HTMLElement>();
const active = ref<{ index: number; left: number; top: number } | null>(null);

const columns = computed(() => Array.from({ length: props.days }, (_, i) => ({
  key: i,
  future: i > props.lastDay,
  running: props.running && i === props.lastDay,
  segments: i > props.lastDay
    ? []
    : props.sources
        .filter(source => (source.daily[i] ?? 0) > 0)
        .map(source => ({
          key: source.key,
          height: `${(((source.daily[i] ?? 0) / (props.peak || 1)) * 100).toFixed(2)}%`,
          color: source.color,
        })),
})));

const detail = computed(() => {
  if (!active.value) return null;

  const i = active.value.index;
  const running = props.running && i === props.lastDay;
  const start = dayDate(props.periodStart, i);
  const end = running ? new Date(props.measuredUntil) : dayDate(props.periodStart, i + 1);
  const midnight = start.getHours() === 0 && start.getMinutes() === 0;
  const sameDay = start.toDateString() === end.toDateString();
  const lines = props.sources
    .map(source => ({ key: source.key, name: source.name, color: source.color, cents: source.daily[i] ?? 0 }))
    .filter(line => line.cents > 0);

  return {
    range: midnight ? formatDay(start) : `${formatDayTime(start)} to ${sameDay ? formatTime(end) : formatDayTime(end)}`,
    total: formatCents(sum(lines.map(line => line.cents))),
    running,
    lines: props.sources.length > 1 ? lines : [],
  };
});

function point(event: PointerEvent) {
  const element = root.value;
  if (!element || !props.days) return;

  const box = element.getBoundingClientRect();
  const index = Math.min(props.days - 1, Math.max(0, Math.floor(((event.clientX - box.left) / box.width) * props.days)));

  if (index > props.lastDay) {
    active.value = null;
    return;
  }

  active.value = { index, left: box.left + ((index + 0.5) / props.days) * box.width, top: box.top };
}

function leave(event: PointerEvent) {
  if (event.pointerType === 'mouse') active.value = null;
}

useEventListener(window, 'scroll', () => {
  active.value = null;
}, { capture: true, passive: true });

useEventListener(document, 'pointerdown', (event) => {
  if (!root.value?.contains(event.target as Node)) active.value = null;
});
</script>

<template>
  <div
    v-bind="$attrs"
    ref="root"
    role="img"
    :aria-label="label"
    class="flex h-8 items-end gap-0.5 border-b"
    @pointermove="point"
    @pointerdown="point"
    @pointerleave="leave"
    @click.stop
  >
    <div
      v-for="column in columns"
      :key="column.key"
      :class="cn(
        'flex h-full min-w-0 flex-1 flex-col-reverse overflow-hidden rounded-t-xs',
        column.running && 'opacity-50',
        active?.index === column.key && 'bg-muted',
      )"
    >
      <div v-if="column.future" class="h-0.5 bg-muted" />
      <div
        v-for="segment in column.segments"
        :key="segment.key"
        class="min-h-px shrink-0"
        :style="{ height: segment.height, background: segment.color }"
      />
    </div>
  </div>

  <Teleport to="body">
    <div
      v-if="active && detail"
      :class="cn(darkTooltipClass, 'pointer-events-none fixed flex -translate-x-1/2 -translate-y-full flex-col gap-0.5 border tabular-nums shadow-md')"
      :style="{ left: `${active.left}px`, top: `${active.top - 6}px` }"
    >
      <span class="opacity-70">{{ detail.range }}</span>
      <span v-for="line in detail.lines" :key="line.key" class="flex items-center gap-1.5">
        <span class="size-2 shrink-0 rounded-xs" :style="{ background: line.color }" />
        <span>{{ line.name }}</span>
        <span class="ml-auto pl-3">{{ formatCents(line.cents) }}</span>
      </span>
      <span class="font-medium">
        {{ detail.total }}<template v-if="detail.running"> so far</template>
      </span>
    </div>
  </Teleport>
</template>
