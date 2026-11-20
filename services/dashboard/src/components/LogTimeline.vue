<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ChevronRight, X } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import { Status } from '@/components/ui/status';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { ServiceInstanceState } from '@/gql/graphql';
import { formatCount, formatDateTime } from '@/lib/format';
import { axisTicks, logClock, shortPod } from '@/lib/logs';
import { instanceStatus } from '@/lib/status';
import { logTooltipClass } from '@/lib/tooltip';

interface TimelineInstance {
  pod: string;
  firstSeen: string;
  lastSeen: string;
  restarts: string[];
  state: ServiceInstanceState;
}

interface TimelineSlot {
  time: string;
  stdout: number;
  stderr: number;
}

const props = defineProps<{
  from: number;
  to: number;
  volume: TimelineSlot[];
  instances: TimelineInstance[];
  selected: string | null;
  band: { from: number; to: number } | null;
  total: boolean;
  loading: boolean;
  colorOf: (pod: string) => string;
}>();

const emit = defineEmits<{
  (e: 'select', pod: string | null): void;
  (e: 'range', range: { from: number; to: number }): void;
  (e: 'seek', at: number, final: boolean): void;
}>();

const COLLAPSE_AFTER = 6;

const expanded = ref(false);
const chart = ref<HTMLElement | null>(null);
const drag = ref<{ start: number; end: number } | null>(null);
const ghost = ref<number | null>(null);
const grabbing = ref(false);
let grabOffset = 0;
let seekFrame: number | undefined;
let settle: ReturnType<typeof setTimeout> | undefined;

const span = computed(() => Math.max(props.to - props.from, 1));

function fraction(at: number): number {
  return Math.min(Math.max((at - props.from) / span.value, 0), 1);
}

const peak = computed(() => Math.max(1, ...props.volume.map(slot => slot.stdout + slot.stderr)));
const lines = computed(() => props.volume.reduce((sum, slot) => sum + slot.stdout + slot.stderr, 0));

const bars = computed(() => {
  const starts = props.volume.map(slot => Date.parse(slot.time));
  const step = starts.length > 1 ? starts[1]! - starts[0]! : span.value;
  return props.volume.map((slot, index) => {
    const left = fraction(starts[index]!);
    return {
      key: slot.time,
      left: left * 100,
      width: (fraction(starts[index]! + step) - left) * 100,
      stdout: (slot.stdout / peak.value) * 100,
      stderr: (slot.stderr / peak.value) * 100,
    };
  });
});

const ticks = computed(() => axisTicks(props.from, props.to));

const stopped = computed(() => props.instances.filter(instance => instance.state === ServiceInstanceState.Stopped && instance.pod !== props.selected));
const collapsible = computed(() => props.instances.length > COLLAPSE_AFTER && stopped.value.length > 0);

function when(time: string): string {
  return new Date(time).toDateString() === new Date().toDateString() ? logClock(time).time : formatDateTime(time);
}

const lanes = computed(() => {
  const hidden = collapsible.value && !expanded.value ? new Set(stopped.value.map(instance => instance.pod)) : new Set<string>();
  return props.instances
    .filter(instance => !hidden.has(instance.pod))
    .map(instance => {
      const start = fraction(Date.parse(instance.firstSeen));
      const end = instance.state === ServiceInstanceState.Stopped ? fraction(Date.parse(instance.lastSeen)) : 1;
      return {
        ...instance,
        short: shortPod(instance.pod),
        left: start * 100,
        width: Math.max(end - start, 0.004) * 100,
        restartsAt: instance.restarts.map(time => fraction(Date.parse(time)) * 100),
        status: instanceStatus(instance.state),
        dimmed: !!props.selected && props.selected !== instance.pod,
      };
    });
});

const band = computed(() => {
  if (!props.band) return null;
  const left = ghost.value ?? fraction(props.band.from);
  const width = Math.max(fraction(props.band.to) - fraction(props.band.from), 0.006);
  return { left: left * 100, width: width * 100, at: props.from + left * span.value };
});

const bandLabel = computed(() => (band.value ? when(new Date(band.value.at).toISOString()) : ''));

watch(() => props.band, () => {
  if (grabbing.value) return;
  clearTimeout(settle);
  ghost.value = null;
});

onBeforeUnmount(() => {
  clearTimeout(settle);
  if (seekFrame !== undefined) cancelAnimationFrame(seekFrame);
});

const selection = computed(() => {
  if (!drag.value) return null;
  const left = Math.min(drag.value.start, drag.value.end);
  return { left: left * 100, width: Math.abs(drag.value.end - drag.value.start) * 100 };
});

function position(event: PointerEvent): number {
  const rect = chart.value!.getBoundingClientRect();
  return Math.min(Math.max((event.clientX - rect.left) / rect.width, 0), 1);
}

function startDrag(event: PointerEvent) {
  chart.value?.setPointerCapture(event.pointerId);
  const at = position(event);
  drag.value = { start: at, end: at };
}

function moveDrag(event: PointerEvent) {
  if (drag.value) drag.value = { ...drag.value, end: position(event) };
}

function endDrag() {
  const current = drag.value;
  drag.value = null;
  if (!current) return;
  const start = Math.min(current.start, current.end);
  const end = Math.max(current.start, current.end);
  if (end - start < 0.01) {
    emit('seek', props.from + current.start * span.value, true);
    return;
  }
  emit('range', { from: props.from + start * span.value, to: props.from + end * span.value });
}

function grabBand(event: PointerEvent) {
  if (!props.band) return;
  event.stopPropagation();
  (event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
  clearTimeout(settle);
  grabbing.value = true;
  ghost.value = fraction(props.band.from);
  grabOffset = position(event) - ghost.value;
}

function moveBand(event: PointerEvent) {
  if (!grabbing.value) return;
  ghost.value = Math.min(Math.max(position(event) - grabOffset, 0), 1);
  seekFrame ??= requestAnimationFrame(() => {
    seekFrame = undefined;
    if (ghost.value !== null) emit('seek', props.from + ghost.value * span.value, false);
  });
}

function dropBand() {
  if (!grabbing.value || ghost.value === null) return;
  grabbing.value = false;
  emit('seek', props.from + ghost.value * span.value, true);
  settle = setTimeout(() => { ghost.value = null; }, 4000);
}

function cancelBand() {
  grabbing.value = false;
  ghost.value = null;
}

function stepBand(event: KeyboardEvent) {
  if (!props.band) return;
  const step = (span.value / 50) * (event.shiftKey ? 5 : 1);
  if (event.key === 'ArrowLeft') emit('seek', Math.max(props.band.from - step, props.from), true);
  else if (event.key === 'ArrowRight') emit('seek', Math.min(props.band.from + step, props.to), true);
  else return;
  event.preventDefault();
}

function toggle(pod: string) {
  emit('select', props.selected === pod ? null : pod);
}
</script>

<template>
  <div class="flex shrink-0 flex-col border-b border-zinc-800 px-4 pt-1 pb-3">
    <div class="flex gap-2 overflow-hidden [scrollbar-gutter:stable]">
      <div class="w-12 shrink-0" />
      <div
        ref="chart"
        role="img"
        aria-label="Lines over time. Click a time to jump there, or drag across it to look at a shorter range."
        class="relative flex h-10 min-w-0 flex-1 cursor-crosshair touch-none items-end gap-px select-none"
        @pointerdown="startDrag"
        @pointermove="moveDrag"
        @pointerup="endDrag"
        @pointercancel="drag = null"
      >
        <template v-if="loading && !volume.length">
          <div v-for="slot in 60" :key="slot" class="flex-1 bg-zinc-800" :style="{ height: `${20 + ((slot * 29) % 9) * 5}%` }" />
        </template>
        <div v-else-if="!lines" class="h-px flex-1 bg-zinc-800" />
        <div
          v-for="bar in lines ? bars : []"
          :key="bar.key"
          class="absolute inset-y-0 flex flex-col justify-end pr-px"
          :style="{ left: `${bar.left}%`, width: `${bar.width}%` }"
        >
          <div class="bg-status-warn" :style="{ height: `${bar.stderr}%` }" />
          <div class="bg-zinc-600" :style="{ height: `${bar.stdout}%` }" />
        </div>
        <div
          v-if="band"
          role="slider"
          tabindex="0"
          aria-label="Position in the logs"
          :aria-valuemin="from"
          :aria-valuemax="to"
          :aria-valuenow="Math.round(band.at)"
          :aria-valuetext="bandLabel"
          class="absolute inset-y-0 cursor-grab touch-none border-x border-zinc-500 bg-zinc-100/5 hover:border-zinc-300 hover:bg-zinc-100/10 focus-visible:ring-2 focus-visible:ring-zinc-500 focus-visible:outline-none"
          :class="grabbing && 'cursor-grabbing border-zinc-300 bg-zinc-100/15'"
          :style="{ left: `${band.left}%`, width: `max(${band.width}%, 0.5rem)` }"
          @pointerdown="grabBand"
          @pointermove="moveBand"
          @pointerup="dropBand"
          @pointercancel="cancelBand"
          @keydown="stepBand"
        />
        <div
          v-if="selection"
          class="pointer-events-none absolute inset-y-0 border-x border-zinc-400 bg-zinc-100/15"
          :style="{ left: `${selection.left}%`, width: `${selection.width}%` }"
        />
      </div>
      <div class="flex w-18 shrink-0 items-end text-xs text-zinc-400 max-md:hidden">
        <span v-if="total">{{ formatCount(lines) }} {{ lines === 1 ? 'line' : 'lines' }}</span>
      </div>
    </div>

    <div class="flex gap-2 overflow-hidden [scrollbar-gutter:stable]" aria-hidden="true">
      <div class="w-12 shrink-0" />
      <div class="relative mb-2 h-4 min-w-0 flex-1 text-xs leading-4 text-zinc-500">
        <span
          v-for="tick in ticks"
          :key="tick.at"
          class="absolute top-0.5 -translate-x-1/2 whitespace-nowrap"
          :style="{ left: `${tick.at * 100}%` }"
        >{{ tick.label }}</span>
      </div>
      <div class="w-18 shrink-0 max-md:hidden" />
    </div>

    <div class="flex max-h-36 flex-col gap-1 overflow-y-auto [scrollbar-gutter:stable]">
      <template v-if="loading && !instances.length">
        <div v-for="row in 2" :key="row" class="flex h-4 shrink-0 items-center gap-2">
          <div class="w-12 shrink-0"><div class="h-2 w-9 rounded-md bg-zinc-800" /></div>
          <div class="h-1.5 min-w-0 flex-1 rounded-full bg-zinc-800" />
          <div class="w-18 shrink-0 max-md:hidden"><div class="h-2 w-12 rounded-md bg-zinc-800" /></div>
        </div>
      </template>
      <div v-for="lane in lanes" :key="lane.pod" class="flex h-4 shrink-0 items-center gap-2">
        <span
          class="w-12 shrink-0 truncate font-mono text-xs leading-4"
          :class="[lane.dimmed && 'opacity-35', selected === lane.pod && 'font-medium']"
          :style="{ color: colorOf(lane.pod) }"
        >{{ lane.short }}</span>
        <Tooltip>
          <TooltipTrigger as-child>
            <button
              type="button"
              class="relative block h-4 min-w-0 flex-1"
              :class="lane.dimmed && 'opacity-35'"
              :aria-label="`Show only ${lane.short}`"
              :aria-pressed="selected === lane.pod"
              @click="toggle(lane.pod)"
            >
              <span class="absolute inset-x-0 top-1/2 h-0.5 -translate-y-1/2 bg-zinc-900" />
              <span
                class="absolute top-1/2 h-1.5 -translate-y-1/2 rounded-full opacity-60"
                :style="{ left: `${lane.left}%`, width: `${lane.width}%`, background: colorOf(lane.pod) }"
              />
              <span
                v-for="(at, index) in lane.restartsAt"
                :key="index"
                class="absolute inset-y-0.5 w-0.5 bg-red-400"
                :style="{ left: `${at}%` }"
              />
              <span
                v-if="band"
                class="pointer-events-none absolute inset-y-0 border-x border-zinc-500 bg-zinc-100/5"
                :style="{ left: `${band.left}%`, width: `max(${band.width}%, 0.5rem)` }"
              />
            </button>
          </TooltipTrigger>
          <TooltipContent :class="logTooltipClass">
            <p class="font-mono">{{ lane.short }}</p>
            <p>First line {{ when(lane.firstSeen) }}</p>
            <p v-for="time in lane.restarts" :key="time" class="text-red-300">Restarted {{ when(time) }}</p>
            <p v-if="lane.state === ServiceInstanceState.Stopped">Last line {{ when(lane.lastSeen) }}</p>
          </TooltipContent>
        </Tooltip>
        <div class="flex w-18 shrink-0 items-center justify-between gap-1 text-xs max-md:hidden">
          <Status variant="text" :tone="lane.status.tone" :class="[lane.status.tone === 'neutral' && 'text-zinc-400', lane.dimmed && 'opacity-35']">
            {{ lane.status.label }}
          </Status>
          <IconButton
            v-if="selected === lane.pod"
            label="Show all instances"
            size="icon-xs"
            class="size-4 bg-zinc-800 text-zinc-200 hover:bg-zinc-700 hover:text-zinc-100 [&_svg]:size-3"
            :tooltip-class="logTooltipClass"
            @click="emit('select', null)"
          >
            <X />
          </IconButton>
        </div>
      </div>
    </div>

    <button
      v-if="collapsible"
      type="button"
      class="mt-1.5 flex items-center gap-1.5 self-start text-xs text-zinc-400 hover:text-zinc-200"
      :aria-expanded="expanded"
      @click="expanded = !expanded"
    >
      <ChevronRight class="size-3 transition-transform" :class="expanded && 'rotate-90'" />
      {{ expanded ? 'Hide stopped instances' : `${stopped.length} stopped instances` }}
    </button>
  </div>
</template>
