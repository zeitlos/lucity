<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, useTemplateRef, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ExternalLink, FastForward, Pause, Play, X } from '@lucide/vue';
import Player from 'rrweb-player';
import 'rrweb-player/dist/style.css';

import { SessionReplayDocument } from '@/lib/users';
import { openExternal } from '@/lib/workspaces';

const props = defineProps<{
  sessionId: string;
  title: string;
  detail: string;
  analyticsUrl: string | null;
}>();

const emit = defineEmits<{
  close: [];
}>();

const speeds = [1, 2, 4, 8];
const idleDelay = 2500;

const dialog = useTemplateRef<HTMLDialogElement>('dialog');
const stage = useTemplateRef<HTMLDivElement>('stage');

const { result, loading, error } = useQuery(SessionReplayDocument, () => ({ sessionID: props.sessionId }), { fetchPolicy: 'no-cache' });

const failure = ref<string | null>(null);

const metaEvent = 4;
const snapshotEvent = 2;

type RecordedEvent = { type: number; timestamp: number };

function playable(events: RecordedEvent[]): RecordedEvent[] | null {
  const snapshot = events.findIndex((event) => event.type === snapshotEvent);

  if (snapshot < 0) return null;

  return events.slice(events[snapshot - 1]?.type === metaEvent ? snapshot - 1 : snapshot);
}

type MountedPlayer = Player & {
  $set(props: { width: number; height: number }): void;
  $destroy(): void;
};

type Update<T> = { payload: T };

let player: MountedPlayer | null = null;

const ready = ref(false);
const playing = ref(true);
const currentTime = ref(0);
const duration = ref(0);
const speed = ref(1);
const skipInactive = ref(true);
const skipping = ref(false);

function size() {
  return { width: stage.value!.clientWidth, height: stage.value!.clientHeight };
}

const resizer = new ResizeObserver(() => {
  if (!player) return;

  player.$set(size());
  player.triggerResize();
});

watch(result, (value) => {
  if (!value || !stage.value || player) return;

  const events = playable(JSON.parse(value.sessionReplay));

  if (!events) {
    failure.value = 'The recording has no snapshot of the page, only the changes made to it.';
    return;
  }

  try {
    player = new Player({
      target: stage.value,
      props: {
        events: events as ConstructorParameters<typeof Player>[0]['props']['events'],
        ...size(),
        autoPlay: true,
        skipInactive: skipInactive.value,
        showController: false,
      },
    }) as MountedPlayer;

    player.addEventListener('fullsnapshot-rebuilded', finishLoadingSnapshot);
    player.addEventListener('ui-update-current-time', (update) => (currentTime.value = (update as Update<number>).payload));
    player.addEventListener('ui-update-player-state', (update) => (playing.value = (update as Update<string>).payload === 'playing'));
    player.addEventListener('skip-start', () => (skipping.value = true));
    player.addEventListener('skip-end', () => (skipping.value = false));

    duration.value = player.getMetaData().totalTime;
    ready.value = true;
  } catch (e) {
    failure.value = e instanceof Error ? e.message : String(e);
  }
});

function finishLoadingSnapshot() {
  player?.getReplayer().iframe.contentDocument?.close();
}

function togglePlaying() {
  player?.toggle();
}

function seek(event: Event) {
  player?.goto(Number((event.target as HTMLInputElement).value));
}

function nextSpeed() {
  speed.value = speeds[(speeds.indexOf(speed.value) + 1) % speeds.length]!;
  player?.setSpeed(speed.value);
}

function toggleSkipInactive() {
  skipInactive.value = !skipInactive.value;
  player?.toggleSkipInactive();
}

function clock(milliseconds: number): string {
  const seconds = Math.floor(milliseconds / 1000);
  const minutes = Math.floor(seconds / 60);
  const pad = (value: number) => String(value).padStart(2, '0');

  if (minutes >= 60) return `${Math.floor(minutes / 60)}:${pad(minutes % 60)}:${pad(seconds % 60)}`;

  return `${pad(minutes)}:${pad(seconds % 60)}`;
}

const idle = ref(false);
const hovering = ref(false);
const chromeHidden = computed(() => ready.value && playing.value && idle.value && !hovering.value);

let idleTimer: ReturnType<typeof setTimeout> | undefined;

function wake() {
  idle.value = false;
  clearTimeout(idleTimer);
  idleTimer = setTimeout(() => (idle.value = true), idleDelay);
}

function onKeydown(event: KeyboardEvent) {
  wake();

  if (event.key !== ' ' || (event.target as HTMLElement).closest('button, input')) return;

  event.preventDefault();
  togglePlaying();
}

function close() {
  dialog.value?.close();
}

function closeOnBackdrop(event: MouseEvent) {
  if (event.target === dialog.value) close();
}

onMounted(() => {
  dialog.value!.showModal();
  resizer.observe(stage.value!);
  wake();
});

onBeforeUnmount(() => {
  clearTimeout(idleTimer);
  resizer.disconnect();
  player?.pause();
  player?.$destroy();
  player = null;
});
</script>

<template>
  <dialog
    ref="dialog"
    class="fixed inset-4 m-0 h-auto max-h-none w-auto max-w-none overflow-hidden rounded-xl bg-neutral-950 p-0 text-white shadow-lg ring-1 ring-white/10 transition-[opacity,scale] duration-200 backdrop:bg-black/60 starting:open:scale-[0.98] starting:open:opacity-0 motion-reduce:transition-none"
    :class="chromeHidden ? 'cursor-none' : ''"
    :aria-label="`Replay of ${title}`"
    @click="closeOnBackdrop"
    @close="emit('close')"
    @keydown="onKeydown"
    @mousemove="wake"
  >
    <div ref="stage" class="replay-stage absolute inset-0" />

    <p v-if="loading && !result" class="absolute inset-0 flex items-center justify-center text-white/60">Loading replay…</p>
    <div v-else-if="error || failure" class="absolute inset-0 flex items-center justify-center p-6">
      <div class="max-w-md rounded-lg bg-white/5 px-4 py-3 ring-1 ring-white/10">
        <p class="font-medium text-red-300">The replay could not be played</p>
        <p class="mt-1 text-white/60">{{ error?.message ?? failure }}</p>
      </div>
    </div>

    <div
      class="pointer-events-none absolute inset-x-0 top-0 flex items-start gap-3 p-3 pl-20 transition-opacity duration-300"
      :class="chromeHidden ? 'opacity-0' : 'opacity-100'"
    >
      <div
        class="pointer-events-auto flex min-w-0 flex-col rounded-lg bg-black/60 px-3 py-1.5 ring-1 ring-white/10 backdrop-blur-md"
        @mouseenter="hovering = true"
        @mouseleave="hovering = false"
      >
        <h2 class="truncate font-semibold">{{ title }}</h2>
        <p class="truncate text-white/60">{{ detail }}</p>
      </div>
      <div class="pointer-events-auto ml-auto flex shrink-0 gap-2" @mouseenter="hovering = true" @mouseleave="hovering = false">
        <button
          v-if="analyticsUrl"
          type="button"
          class="flex h-9 items-center gap-2 rounded-full bg-black/60 px-3.5 ring-1 ring-white/10 backdrop-blur-md hover:bg-black/80"
          @click="openExternal(analyticsUrl)"
        >
          <ExternalLink class="size-3.5" aria-hidden="true" />
          Rybbit
        </button>
        <button
          type="button"
          class="flex size-9 items-center justify-center rounded-full bg-black/60 ring-1 ring-white/10 backdrop-blur-md hover:bg-black/80"
          aria-label="Close"
          @click="close"
        >
          <X class="size-4" aria-hidden="true" />
        </button>
      </div>
    </div>

    <div
      v-if="ready"
      class="pointer-events-none absolute inset-x-0 bottom-0 flex justify-center p-4 transition-opacity duration-300"
      :class="chromeHidden ? 'opacity-0' : 'opacity-100'"
    >
      <div
        class="pointer-events-auto flex w-full max-w-3xl items-center gap-3 rounded-full bg-black/60 py-1.5 pr-2 pl-1.5 ring-1 ring-white/10 backdrop-blur-md"
        @mouseenter="hovering = true"
        @mouseleave="hovering = false"
      >
        <button
          type="button"
          class="flex size-9 shrink-0 items-center justify-center rounded-full bg-white text-neutral-950 hover:bg-white/85"
          :aria-label="playing ? 'Pause' : 'Play'"
          @click="togglePlaying"
        >
          <component :is="playing ? Pause : Play" class="size-4 fill-current" aria-hidden="true" />
        </button>
        <span class="w-14 shrink-0 text-right text-white/80 tabular-nums">{{ clock(currentTime) }}</span>
        <input
          type="range"
          class="min-w-0 flex-1 cursor-pointer accent-white"
          min="0"
          :max="duration"
          step="100"
          :value="currentTime"
          aria-label="Position"
          @input="seek"
        >
        <span class="w-14 shrink-0 text-white/60 tabular-nums">{{ clock(duration) }}</span>
        <button
          type="button"
          class="h-8 w-11 shrink-0 rounded-full text-white/80 tabular-nums hover:bg-white/10"
          aria-label="Playback speed"
          @click="nextSpeed"
        >{{ speed }}×</button>
        <button
          type="button"
          class="flex h-8 shrink-0 items-center gap-1.5 rounded-full px-3 hover:bg-white/10"
          :class="skipInactive ? 'text-white' : 'text-white/50'"
          :aria-pressed="skipInactive"
          @click="toggleSkipInactive"
        >
          <FastForward class="size-4" aria-hidden="true" />
          {{ skipping ? 'Skipping…' : 'Skip idle' }}
        </button>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.replay-stage :deep(.rr-player) {
  float: none;
  background: transparent;
  border-radius: 0;
  box-shadow: none;
}

.replay-stage :deep(.replayer-wrapper > iframe) {
  background: #fff;
}
</style>
