<script lang="ts">
export interface LogEntry {
  text: string;
  pod?: string;
}
</script>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';
import { useLocalStorage } from '@vueuse/core';
import Anser from 'anser';
import {
  ArrowDownToLine,
  CircleAlert,
  CircleCheck,
  CircleX,
  Download,
  Eraser,
  Pause,
  Play,
  RotateCw,
  Search,
  WrapText,
  X,
} from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import IconButton from '@/components/IconButton.vue';
import LogLine from '@/components/LogLine.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { describeError } from '@/lib/errors';
import { formatCount } from '@/lib/format';

const props = withDefaults(defineProps<{
  title: string;
  lines: LogEntry[];
  state: 'connecting' | 'streaming' | 'finished' | 'error';
  error?: unknown;
  end?: { text: string; tone: 'ok' | 'danger' | 'neutral' };
  waitingHint?: string;
  fileName: string;
}>(), {
  error: undefined,
  end: undefined,
  waitingHint: undefined,
});

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'retry'): void;
}>();

const ALL = '__all__';
const podColors = ['var(--chart-1)', 'var(--chart-2)', 'var(--chart-4)', 'var(--chart-5)', 'var(--chart-3)'];
const quietButton = 'text-zinc-300 hover:bg-zinc-800 hover:text-zinc-100 aria-pressed:bg-zinc-800 aria-pressed:text-zinc-100';

const filter = ref('');
const pod = ref(ALL);
const wrap = useLocalStorage('lucity_logs_wrap', true);
const clearedAt = ref(0);
const frozenAt = ref<number | null>(null);
const behind = ref(false);
const container = ref<HTMLElement | null>(null);

const pods = computed(() => [...new Set(props.lines.map(line => line.pod).filter((name): name is string => !!name))]);

function shortPod(name?: string): string {
  return name?.split('-').pop() ?? '';
}

function podColor(name?: string): string {
  const index = name ? pods.value.indexOf(name) : -1;
  return index < 0 ? 'inherit' : podColors[index % podColors.length]!;
}

const paused = computed(() => frozenAt.value !== null);
const held = computed(() => (frozenAt.value === null ? 0 : props.lines.length - frozenAt.value));

const shown = computed(() => {
  const query = filter.value.trim().toLowerCase();
  const end = frozenAt.value ?? props.lines.length;
  const out: (LogEntry & { index: number })[] = [];
  for (let index = clearedAt.value; index < end; index++) {
    const line = props.lines[index]!;
    if (pod.value !== ALL && line.pod !== pod.value) continue;
    if (query && !Anser.ansiToText(line.text).toLowerCase().includes(query)) continue;
    out.push({ ...line, index });
  }
  return out;
});

const empty = computed(() => props.lines.length - clearedAt.value === 0);

function togglePause() {
  frozenAt.value = paused.value ? null : props.lines.length;
}

function clear() {
  clearedAt.value = props.lines.length;
  if (paused.value) frozenAt.value = props.lines.length;
}

function onScroll() {
  const el = container.value;
  if (!el) return;
  behind.value = el.scrollHeight - el.scrollTop - el.clientHeight > 40;
}

function jumpToLatest() {
  frozenAt.value = null;
  behind.value = false;
  nextTick(() => {
    if (container.value) container.value.scrollTop = container.value.scrollHeight;
  });
}

watch([() => shown.value.length, () => props.state], async () => {
  if (behind.value || paused.value) return;
  await nextTick();
  if (container.value) container.value.scrollTop = container.value.scrollHeight;
});

watch(() => props.lines.length, length => {
  if (length < clearedAt.value) {
    clearedAt.value = 0;
    frozenAt.value = null;
  }
});

function download() {
  const text = props.lines
    .slice(clearedAt.value)
    .map(line => (line.pod ? `[${shortPod(line.pod)}] ` : '') + Anser.ansiToText(line.text))
    .join('\n');
  const url = URL.createObjectURL(new Blob([text], { type: 'text/plain' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = props.fileName;
  link.click();
  URL.revokeObjectURL(url);
}
</script>

<template>
  <section class="flex h-full flex-col overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950 text-zinc-300 shadow-2xl">
    <header class="flex shrink-0 items-center gap-3 px-4 py-3">
      <h2 class="min-w-0 flex-1 truncate text-lg font-semibold text-zinc-100">
        {{ title }}
      </h2>
      <IconButton
        label="Close"
        shortcut="Esc"
        class="text-zinc-400 hover:bg-zinc-800 hover:text-zinc-100"
        @click="emit('close')"
      >
        <X />
      </IconButton>
    </header>

    <div class="flex shrink-0 items-center gap-2 border-b border-zinc-800 px-4 pb-3">
      <div class="relative min-w-0 flex-1">
        <Search :size="14" class="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-zinc-500" />
        <input
          v-model="filter"
          type="text"
          placeholder="Filter lines"
          aria-label="Filter lines"
          class="h-8 w-full rounded-md border border-zinc-800 bg-zinc-900 pr-2 pl-8 text-sm text-zinc-200 placeholder:text-zinc-500 focus-visible:ring-2 focus-visible:ring-zinc-700 focus-visible:outline-none"
        >
      </div>
      <Select v-if="pods.length > 1" v-model="pod">
        <SelectTrigger class="h-8 w-40 shrink-0 border-zinc-800 bg-zinc-900 text-zinc-200" aria-label="Replica">
          <SelectValue>{{ pod === ALL ? `All replicas (${pods.length})` : shortPod(pod) }}</SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem :value="ALL">All replicas ({{ pods.length }})</SelectItem>
          <SelectItem v-for="name in pods" :key="name" :value="name">
            {{ shortPod(name) }}
          </SelectItem>
        </SelectContent>
      </Select>
      <div class="flex shrink-0 items-center gap-0.5">
        <IconButton :label="wrap ? 'Don\'t wrap lines' : 'Wrap lines'" :class="quietButton" :aria-pressed="wrap" @click="wrap = !wrap">
          <WrapText />
        </IconButton>
        <IconButton
          v-if="state !== 'finished'"
          :label="paused ? (held ? `Resume (${formatCount(held)} new)` : 'Resume') : 'Pause'"
          :class="quietButton"
          :aria-pressed="paused"
          @click="togglePause"
        >
          <Play v-if="paused" />
          <Pause v-else />
        </IconButton>
        <IconButton label="Download" :class="quietButton" :disabled="empty" @click="download">
          <Download />
        </IconButton>
        <IconButton label="Clear" :class="quietButton" :disabled="empty" @click="clear">
          <Eraser />
        </IconButton>
      </div>
    </div>

    <div class="relative min-h-0 flex-1">
      <div
        ref="container"
        class="h-full overflow-auto py-2 font-mono text-xs leading-relaxed"
        @scroll="onScroll"
      >
        <div v-if="state === 'error'" class="flex flex-col items-start gap-1.5 px-4 py-2 font-sans">
          <p class="flex items-center gap-2 text-sm font-medium text-red-400">
            <CircleAlert :size="14" />
            Couldn't load logs
          </p>
          <p class="text-xs text-zinc-400">{{ describeError(error).message }}</p>
          <Button variant="ghost" size="xs" class="mt-1 bg-zinc-800 text-zinc-200 hover:bg-zinc-700 hover:text-zinc-100" @click="emit('retry')">
            <RotateCw />
            Try again
          </Button>
        </div>

        <div
          v-else-if="empty && state !== 'finished'"
          class="flex flex-col gap-1 px-4 py-2 font-sans"
        >
          <p class="flex items-center gap-2 text-sm text-zinc-300">
            <Spinner :size="12" class="text-zinc-500" />
            Waiting for output…
          </p>
          <p v-if="waitingHint" class="text-xs text-zinc-500">{{ waitingHint }}</p>
        </div>

        <p
          v-else-if="!empty && shown.length === 0 && filter.trim()"
          class="px-4 py-2 font-sans text-sm text-zinc-500"
        >
          No lines match "{{ filter.trim() }}".
        </p>

        <div
          v-for="line in shown"
          :key="line.index"
          :class="['flex gap-3 px-4 hover:bg-zinc-900', !wrap && 'w-max min-w-full']"
        >
          <span class="w-10 shrink-0 text-right text-zinc-600 tabular-nums select-none">{{ line.index + 1 }}</span>
          <span
            v-if="pods.length"
            class="w-12 shrink-0 truncate select-none"
            :style="{ color: podColor(line.pod) }"
          >{{ shortPod(line.pod) }}</span>
          <LogLine :text="line.text" :wrap="wrap" class="min-w-0 flex-1 text-zinc-300" />
        </div>

        <p
          v-if="state === 'finished'"
          class="mt-2 flex items-center gap-2 px-4 py-1 font-sans text-xs text-zinc-400"
        >
          <CircleCheck v-if="end?.tone === 'ok'" :size="14" class="text-status-ok" />
          <CircleX v-else-if="end?.tone === 'danger'" :size="14" class="text-red-400" />
          {{ empty ? `${end?.text ?? 'Finished'} without output.` : `${end?.text ?? 'Finished'}.` }}
        </p>
      </div>

      <Button
        v-if="behind || held"
        size="xs"
        class="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-zinc-800 text-zinc-100 shadow-lg hover:bg-zinc-700"
        @click="jumpToLatest"
      >
        <ArrowDownToLine />
        Jump to latest
      </Button>
    </div>
  </section>
</template>
