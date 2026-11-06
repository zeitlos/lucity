<script setup lang="ts">
import { computed, nextTick, ref, shallowRef, watch } from 'vue';
import { useLocalStorage } from '@vueuse/core';
import Anser from 'anser';
import {
  Activity,
  ArrowDownToLine,
  CalendarRange,
  CircleAlert,
  Download,
  FoldVertical,
  Pause,
  Play,
  RotateCcw,
  RotateCw,
  ScrollText,
  Search,
  SearchX,
  Square,
  UnfoldVertical,
  WrapText,
  X,
} from '@lucide/vue';
import CopyButton from '@/components/CopyButton.vue';
import IconButton from '@/components/IconButton.vue';
import LogLine from '@/components/LogLine.vue';
import LogTimeline from '@/components/LogTimeline.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useEscapeToClose } from '@/composables/useEscapeToClose';
import { useServiceLogs, type LogEntry } from '@/composables/useServiceLogs';
import { LogStream, ServiceLogEvent } from '@/gql/graphql';
import { formatCount, formatRelative } from '@/lib/format';
import { eventText, formatRangeEdge, logClock, logDate, shortPod, type LogRange } from '@/lib/logs';
import { describeError } from '@/lib/errors';
import { logTooltipClass } from '@/lib/tooltip';

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

useEscapeToClose(() => emit('close'));

const quietButton = 'text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 aria-pressed:bg-zinc-800 aria-pressed:text-zinc-100';
const pickerInput = 'h-8 border-zinc-800 bg-zinc-900 text-sm text-zinc-200 scheme-dark focus-visible:ring-zinc-600 focus-visible:ring-offset-zinc-950';

const serviceIdRef = computed(() => props.serviceId);
const logs = useServiceLogs(serviceIdRef);
const { range, search, pod, surrounding, bounds, live, searching, entries, instances, volume, loading, error, hasEarlier, hasLater, loadingEarlier, loadingLater } = logs;

const wrap = useLocalStorage('lucity_logs_wrap', true);

type RangeOption = Exclude<LogRange, 'custom'>;
const rangeOptions: { value: RangeOption; label: string }[] = [
  { value: 'live', label: 'Live' },
  { value: '1h', label: '1h' },
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
];

const rangeModel = computed<RangeOption>({
  get: () => (range.value === 'custom' ? ('' as RangeOption) : range.value),
  set: value => logs.setRange(value),
});

const pickerOpen = ref(false);
const pickFrom = ref('');
const pickTo = ref('');

function localInput(at: number): string {
  const date = new Date(at - new Date(at).getTimezoneOffset() * 60_000);
  return date.toISOString().slice(0, 19);
}

watch(pickerOpen, open => {
  if (!open) return;
  pickFrom.value = localInput(bounds.value.from);
  pickTo.value = localInput(bounds.value.to);
});

const pickedRange = computed(() => {
  const from = new Date(pickFrom.value).getTime();
  const to = new Date(pickTo.value).getTime();
  return { from, to, valid: Number.isFinite(from) && Number.isFinite(to) && from < to };
});

function applyPick() {
  if (!pickedRange.value.valid) return;
  logs.setRange('custom', { from: pickedRange.value.from, to: pickedRange.value.to });
  pickerOpen.value = false;
}

const rangeLabel = computed(() => `${formatRangeEdge(bounds.value.from, bounds.value.to)} to ${formatRangeEdge(bounds.value.to, bounds.value.from)}`);

const frozen = shallowRef<LogEntry[] | null>(null);
const paused = computed(() => frozen.value !== null);
const held = computed(() => (frozen.value ? Math.max(entries.value.length - frozen.value.length, 0) : 0));
const shown = computed(() => frozen.value ?? entries.value);

watch(live, () => { frozen.value = null; });

function togglePause() {
  frozen.value = paused.value ? null : entries.value;
}

const matcher = computed(() => {
  const term = logs.filterSearch.value;
  if (!term) return null;
  return new RegExp(`(${term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
});

function highlight(text: string): { text: string; hit: boolean }[] {
  const plain = Anser.ansiToText(text);
  if (!matcher.value) return [{ text: plain, hit: false }];
  return plain
    .split(matcher.value)
    .map((part, index) => ({ text: part, hit: index % 2 === 1 }))
    .filter(part => part.text);
}

function matches(text: string): boolean {
  return !matcher.value || Anser.ansiToText(text).toLowerCase().includes(logs.filterSearch.value.toLowerCase());
}

const rows = computed(() => shown.value.map((entry, index) => ({
  key: `${entry.time}-${entry.pod}-${index}`,
  entry,
  clock: logClock(entry.time),
  short: shortPod(entry.pod),
  color: logs.colorOf(entry.pod),
  error: entry.stream === LogStream.Stderr,
  parts: !entry.event && matcher.value ? highlight(entry.line) : null,
  dimmed: !entry.event && surrounding.value && !matches(entry.line),
})));

const eventIcons = {
  [ServiceLogEvent.Started]: Play,
  [ServiceLogEvent.Restarted]: RotateCcw,
  [ServiceLogEvent.Stopping]: Square,
  [ServiceLogEvent.HealthCheckFailed]: Activity,
  [ServiceLogEvent.RestartDelayed]: CircleAlert,
};

function eventTone(event: ServiceLogEvent): string {
  if (event === ServiceLogEvent.Restarted || event === ServiceLogEvent.RestartDelayed) return 'text-red-400';
  if (event === ServiceLogEvent.HealthCheckFailed) return 'text-status-warn';
  return 'text-zinc-400';
}

const empty = computed(() => !loading.value && !error.value && shown.value.length === 0);

const emptyTitle = computed(() => {
  if (searching.value) return `No lines match "${logs.filterSearch.value}"`;
  return live.value ? 'No output yet' : 'No logs in this time range';
});

const container = ref<HTMLElement | null>(null);
const list = ref<HTMLElement | null>(null);
const behind = ref(false);
const band = ref<{ from: number; to: number } | null>(null);
let measuring: number | undefined;

function firstIndex(items: HTMLCollection, test: (row: HTMLElement) => boolean): number {
  let low = 0;
  let high = items.length;
  while (low < high) {
    const middle = (low + high) >> 1;
    if (test(items[middle] as HTMLElement)) high = middle;
    else low = middle + 1;
  }
  return low;
}

function measure() {
  measuring = undefined;
  const el = container.value;
  const items = list.value?.children;
  if (!el || !items || !items.length || !rows.value.length) {
    band.value = null;
    return;
  }
  const top = el.scrollTop;
  const bottom = top + el.clientHeight;
  const first = Math.min(firstIndex(items, row => row.offsetTop + row.offsetHeight > top), rows.value.length - 1);
  const last = Math.max(firstIndex(items, row => row.offsetTop >= bottom) - 1, first);
  band.value = {
    from: Date.parse(rows.value[first]!.entry.time),
    to: Date.parse(rows.value[Math.min(last, rows.value.length - 1)]!.entry.time),
  };
}

function scheduleMeasure() {
  measuring ??= requestAnimationFrame(measure);
}

const listArea = ref<HTMLElement | null>(null);
const dateTip = ref<{ date: string; time: string; ago: string; top: number; left: number; below: boolean } | null>(null);

function showDate(event: MouseEvent, time: string) {
  const area = listArea.value?.getBoundingClientRect();
  const cell = (event.currentTarget as HTMLElement).getBoundingClientRect();
  if (!area) return;
  const below = cell.top - area.top < 56;
  const clock = logClock(time);
  dateTip.value = {
    date: logDate(time),
    time: `${clock.time}${clock.fraction}`,
    ago: formatRelative(time),
    top: below ? cell.bottom - area.top + 4 : cell.top - area.top - 4,
    left: cell.left - area.left,
    below,
  };
}

function onScroll() {
  dateTip.value = null;
  const el = container.value;
  if (el) behind.value = el.scrollHeight - el.scrollTop - el.clientHeight > 40;
  scheduleMeasure();
}

function scrollToEnd() {
  nextTick(() => {
    if (container.value) container.value.scrollTop = container.value.scrollHeight;
    scheduleMeasure();
  });
}

function jumpToLatest() {
  frozen.value = null;
  behind.value = false;
  scrollToEnd();
}

function rowAt(at: number): number {
  let low = 0;
  let high = rows.value.length;
  while (low < high) {
    const middle = (low + high) >> 1;
    if (Date.parse(rows.value[middle]!.entry.time) >= at) high = middle;
    else low = middle + 1;
  }
  return Math.min(low, rows.value.length - 1);
}

function scrollToTime(at: number) {
  const el = container.value;
  const items = list.value?.children;
  if (!el || !items?.length) return;
  el.scrollTop = (items[rowAt(at)] as HTMLElement).offsetTop - 8;
  scheduleMeasure();
}

function covers(at: number): boolean {
  if (!shown.value.length) return false;
  const first = Date.parse(shown.value[0]!.time);
  const last = Date.parse(shown.value[shown.value.length - 1]!.time);
  if (live.value) return at >= first;
  return (at >= first || !hasEarlier.value) && (at <= last || !hasLater.value);
}

function seek(at: number, final: boolean) {
  if (covers(at)) scrollToTime(at);
  else if (final) logs.seek(at);
}

let restoreFrom: number | null = null;

async function loadEarlier() {
  restoreFrom = container.value ? container.value.scrollHeight - container.value.scrollTop : null;
  await logs.loadEarlier();
}

watch([() => rows.value.length, logs.version], async () => {
  await nextTick();
  if (live.value) {
    if (behind.value || paused.value) scheduleMeasure();
    else scrollToEnd();
    return;
  }
  const el = container.value;
  if (logs.change.value === 'earlier') {
    if (el && restoreFrom !== null) el.scrollTop = el.scrollHeight - restoreFrom;
    restoreFrom = null;
  } else if (logs.change.value === 'reset') {
    if (logs.target.value !== null) scrollToTime(logs.target.value);
    else scrollToEnd();
  }
  scheduleMeasure();
});

watch([range, bounds], () => {
  if (!live.value) behind.value = false;
});

function download() {
  const text = shown.value
    .map(entry => `${entry.time} ${shortPod(entry.pod)} ${entry.event ? eventText(entry.event, entry.line) : Anser.ansiToText(entry.line)}`)
    .join('\n');
  const url = URL.createObjectURL(new Blob([text], { type: 'text/plain' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = `${props.serviceName}-logs.txt`;
  link.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <section class="dark flex h-full flex-col overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950 text-zinc-300 shadow-2xl max-md:rounded-b-none max-md:border-x-0 max-md:border-b-0 max-md:pb-20">
    <header class="flex shrink-0 items-center gap-0.5 px-4 py-3">
      <h2 class="mr-2.5 min-w-0 flex-1 truncate text-lg font-semibold text-zinc-100">
        Runtime logs
      </h2>
      <IconButton
        v-if="live"
        :label="paused ? (held ? `Resume (${formatCount(held)} new)` : 'Resume') : 'Pause'"
        :class="quietButton"
        :tooltip-class="logTooltipClass"
        :aria-pressed="paused"
        @click="togglePause"
      >
        <Play v-if="paused" />
        <Pause v-else />
      </IconButton>
      <IconButton :label="wrap ? 'Don\'t wrap lines' : 'Wrap lines'" :class="quietButton" :tooltip-class="logTooltipClass" :aria-pressed="wrap" @click="wrap = !wrap">
        <WrapText />
      </IconButton>
      <IconButton label="Download" :class="quietButton" :tooltip-class="logTooltipClass" :disabled="!shown.length" @click="download">
        <Download />
      </IconButton>
      <IconButton
        label="Close"
        shortcut="Esc"
        class="text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
        :tooltip-class="logTooltipClass"
        @click="emit('close')"
      >
        <X />
      </IconButton>
    </header>

    <div class="flex shrink-0 flex-wrap items-center gap-2 px-4 pb-3">
      <div class="relative min-w-0 flex-1 max-md:basis-full">
        <Search :size="14" class="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-zinc-500" />
        <input
          v-model="search"
          type="text"
          placeholder="Search logs"
          aria-label="Search logs"
          class="h-8 w-full rounded-md border border-zinc-800 bg-zinc-900 pr-8 pl-8 text-sm text-zinc-200 placeholder:text-zinc-500 focus-visible:ring-2 focus-visible:ring-zinc-700 focus-visible:outline-none"
        >
        <IconButton
          v-if="search"
          label="Clear search"
          size="icon-xs"
          class="absolute top-1 right-1 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
          :tooltip-class="logTooltipClass"
          @click="search = ''"
        >
          <X />
        </IconButton>
      </div>
      <IconButton
        v-if="searching"
        :label="surrounding ? 'Hide surrounding lines' : 'Show surrounding lines'"
        :class="quietButton"
        :tooltip-class="logTooltipClass"
        :aria-pressed="surrounding"
        @click="surrounding = !surrounding"
      >
        <FoldVertical v-if="surrounding" />
        <UnfoldVertical v-else />
      </IconButton>
      <SegmentedControl
        v-model="rangeModel"
        label="Time range"
        :options="rangeOptions"
        class="bg-zinc-900"
        item-class="h-7 text-zinc-400 hover:text-zinc-100 focus-visible:ring-zinc-600 aria-pressed:bg-zinc-700 aria-pressed:text-zinc-100"
      />
      <Tooltip>
        <TooltipTrigger as-child>
          <span class="inline-flex shrink-0">
            <Popover v-model:open="pickerOpen">
              <PopoverTrigger as-child>
                <Button
                  v-if="range === 'custom'"
                  variant="ghost"
                  size="xs"
                  class="bg-zinc-800 text-zinc-100 hover:bg-zinc-700 hover:text-zinc-100"
                  aria-label="Pick a time range"
                >
                  <CalendarRange />
                  {{ rangeLabel }}
                </Button>
                <Button v-else variant="ghost" size="icon-sm" :class="quietButton" aria-label="Pick a time range">
                  <CalendarRange />
                </Button>
              </PopoverTrigger>
              <PopoverContent align="end" class="w-80 space-y-3 border-zinc-800 bg-zinc-950 text-zinc-300">
                <div class="space-y-1.5">
                  <Label for="logs-from" class="text-xs text-zinc-400">From</Label>
                  <Input id="logs-from" v-model="pickFrom" type="datetime-local" step="1" :class="pickerInput" />
                </div>
                <div class="space-y-1.5">
                  <Label for="logs-to" class="text-xs text-zinc-400">To</Label>
                  <Input id="logs-to" v-model="pickTo" type="datetime-local" step="1" :class="pickerInput" />
                </div>
                <div class="flex items-center justify-end gap-3 border-t border-zinc-800 pt-3">
                  <p v-if="!pickedRange.valid" class="mr-auto text-xs text-zinc-400">Pick a start before the end.</p>
                  <Button size="xs" :disabled="!pickedRange.valid" @click="applyPick">Apply</Button>
                </div>
              </PopoverContent>
            </Popover>
          </span>
        </TooltipTrigger>
        <TooltipContent :class="logTooltipClass">Pick a time range</TooltipContent>
      </Tooltip>
    </div>

    <LogTimeline
      v-if="!error"
      :from="bounds.from"
      :to="bounds.to"
      :volume="volume"
      :instances="instances"
      :selected="pod"
      :band="band"
      :total="searching"
      :loading="logs.timelineLoading.value"
      :color-of="logs.colorOf"
      @select="pod = $event"
      @range="logs.setRange('custom', $event)"
      @seek="seek"
    />
    <div v-else class="border-b border-zinc-800" />

    <div ref="listArea" class="relative min-h-0 flex-1">
      <div
        ref="container"
        class="relative h-full overflow-auto py-2 font-mono text-xs leading-relaxed"
        @scroll="onScroll"
      >
        <div v-if="error" class="flex flex-col items-start gap-1.5 px-4 py-2 font-sans">
          <p class="flex items-center gap-2 text-sm font-medium text-red-400">
            <CircleAlert :size="14" />
            Couldn't load logs
          </p>
          <p class="text-xs text-zinc-400">{{ describeError(error).message }}</p>
          <Button variant="ghost" size="xs" class="mt-1 bg-zinc-800 text-zinc-200 hover:bg-zinc-700 hover:text-zinc-100" @click="logs.retry">
            <RotateCw />
            Try again
          </Button>
        </div>

        <div v-else-if="loading && live" class="flex flex-col gap-1 px-4 py-2 font-sans">
          <p class="flex items-center gap-2 text-sm text-zinc-300">
            <Spinner :size="12" class="text-zinc-500" />
            Waiting for output…
          </p>
          <p class="text-xs text-zinc-500">Lines appear as the service writes them.</p>
        </div>

        <div v-else-if="loading" aria-hidden="true">
          <div v-for="row in 24" :key="row" class="flex h-5 items-center gap-3 px-4">
            <div class="h-2 w-20 rounded-md bg-zinc-900" />
            <div class="h-2 w-9 rounded-md bg-zinc-900" />
            <div class="h-2 rounded-md bg-zinc-900" :style="{ width: `${28 + ((row * 41) % 50)}%` }" />
          </div>
        </div>

        <div v-else-if="empty" class="px-4 py-2 font-sans">
          <div class="flex flex-col items-center gap-1 rounded-lg border border-dashed border-zinc-700 px-6 py-8 text-center">
            <span class="mb-2 flex size-10 items-center justify-center rounded-md border border-zinc-800 bg-zinc-900 text-zinc-400">
              <component :is="searching ? SearchX : ScrollText" :size="18" />
            </span>
            <p class="text-sm font-medium text-zinc-100">{{ emptyTitle }}</p>
            <Button
              v-if="range !== '7d'"
              variant="ghost"
              size="xs"
              class="mt-3 border border-zinc-700 text-zinc-200 hover:bg-zinc-800 hover:text-zinc-100"
              @click="logs.setRange('7d')"
            >
              {{ searching ? 'Search the last 7 days' : 'Show the last 7 days' }}
            </Button>
          </div>
        </div>

        <template v-else>
          <div v-if="hasEarlier" class="px-4 pb-2 font-sans">
            <Button variant="ghost" size="xs" class="bg-zinc-900 text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100" :pending="loadingEarlier" @click="loadEarlier">
              Show earlier lines
            </Button>
          </div>
          <div ref="list">
            <div
              v-for="row in rows"
              :key="row.key"
              :class="['group relative flex items-start gap-3 px-4 hover:bg-zinc-900', row.entry.event && 'bg-zinc-900', !wrap && 'w-max min-w-full']"
            >
              <span v-if="row.error" aria-hidden="true" class="absolute inset-y-1 left-1.5 w-0.5 bg-status-warn" />
              <span
                class="w-22 shrink-0 text-zinc-500 tabular-nums select-none hover:text-zinc-300 max-md:w-14"
                @mouseenter="showDate($event, row.entry.time)"
                @mouseleave="dateTip = null"
              >{{ row.clock.time }}<span class="max-md:hidden">{{ row.clock.fraction }}</span></span>
              <span class="w-10 shrink-0 truncate select-none" :style="{ color: row.color }">{{ row.short }}</span>
              <span
                v-if="row.entry.event"
                :class="['flex min-w-0 flex-1 items-center gap-1.5 font-sans', eventTone(row.entry.event)]"
              >
                <component :is="eventIcons[row.entry.event]" :size="12" class="shrink-0" />
                {{ eventText(row.entry.event, row.entry.line) }}
              </span>
              <span
                v-else-if="row.parts"
                :class="['min-w-0 flex-1', wrap ? 'whitespace-pre-wrap break-all' : 'whitespace-pre', row.dimmed ? 'text-zinc-500' : 'text-zinc-300']"
              ><template v-for="(part, index) in row.parts" :key="index"><mark v-if="part.hit" class="bg-yellow-400/25 text-yellow-100">{{ part.text }}</mark><template v-else>{{ part.text }}</template></template></span>
              <LogLine v-else :text="row.entry.line" :wrap="wrap" :class="['min-w-0 flex-1', row.dimmed ? 'text-zinc-500' : 'text-zinc-300']" />
              <CopyButton
                v-if="!row.entry.event"
                :value="Anser.ansiToText(row.entry.line)"
                label="Copy line"
                :tooltip-class="logTooltipClass"
                class="-my-0.5 size-5 shrink-0 text-zinc-400 opacity-0 group-hover:opacity-100 hover:bg-zinc-800 hover:text-zinc-100 focus-visible:opacity-100"
              />
            </div>
          </div>
          <div v-if="hasLater" class="px-4 pt-2 font-sans">
            <Button variant="ghost" size="xs" class="bg-zinc-900 text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100" :pending="loadingLater" @click="logs.loadLater()">
              Show later lines
            </Button>
          </div>
        </template>
      </div>

      <Button
        v-if="live && (behind || held)"
        size="xs"
        class="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-zinc-800 text-zinc-100 shadow-lg hover:bg-zinc-700"
        @click="jumpToLatest"
      >
        <ArrowDownToLine />
        Jump to latest
      </Button>

      <div
        v-if="dateTip"
        aria-hidden="true"
        class="pointer-events-none absolute z-10 rounded-md border border-zinc-700 bg-zinc-800 px-2 py-1 font-sans text-xs whitespace-nowrap text-zinc-100 shadow-lg"
        :class="!dateTip.below && '-translate-y-full'"
        :style="{ top: `${dateTip.top}px`, left: `${dateTip.left}px` }"
      >
        <p>{{ dateTip.date }}</p>
        <p class="font-mono tabular-nums">{{ dateTip.time }}</p>
        <p class="text-zinc-400">{{ dateTip.ago }}</p>
      </div>
    </div>
  </section>
</template>
