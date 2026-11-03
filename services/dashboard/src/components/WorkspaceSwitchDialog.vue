<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useTemplateRef, watch } from 'vue';
import { useElementSize, usePreferredReducedMotion } from '@vueuse/core';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import type { WorkspaceSwitch } from '@/composables/useContextSwitch';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';

interface Flap {
  sequence: string[];
  index: number;
  targetIndex: number;
  character: string;
  previous: string;
  flips: number;
  final: boolean;
  settled: boolean;
  period: number;
  nextAt: number;
}

const props = defineProps<{ workspaceSwitch: WorkspaceSwitch | null }>();
const emit = defineEmits<{ finished: [] }>();

const { workspaceColor } = useWorkspaceColor();
const reducedMotion = usePreferredReducedMotion();

const blank = '\u00a0';
const drum = [blank, ...'abcdefghijklmnopqrstuvwxyz', '.', '-'];
const maxFlaps = 16;
const flapWidth = 26;
const flapGap = 4;
const tileGap = 10;
const narrowStage = 336;
const wideStage = 528;
const entranceMs = 500;
const entranceSpreadMs = 35;
const minimumRollMs = 1200;
const holdMs = 700;

const shown = ref<WorkspaceSwitch | null>(null);
const flaps = ref<Flap[]>([]);
const tileFlipped = ref(false);
const stage = useTemplateRef<HTMLElement>('stage');
const { width: stageWidth } = useElementSize(stage);

let frameRequest: number | undefined;
let finishTimer: ReturnType<typeof setTimeout> | undefined;
let startedAt = 0;
let settling = false;

function graphemes(text: string) {
  return Array.from(new Intl.Segmenter(undefined, { granularity: 'grapheme' }).segment(text), part => part.segment);
}

function fit(characters: string[]) {
  return characters.length <= maxFlaps ? characters : [...characters.slice(0, maxFlaps - 1), '…'];
}

function initial(name: string) {
  return (graphemes(name)[0] ?? '').toUpperCase();
}

function slot(sequence: string[], character: string) {
  const index = sequence.findIndex(entry => entry.toLowerCase() === character.toLowerCase());
  if (index >= 0) {
    sequence[index] = character;
    return index;
  }
  sequence.push(character);
  return sequence.length - 1;
}

function createFlap(start: string, target: string, position: number): Flap {
  const sequence = [...drum];
  const index = Math.max(0, sequence.findIndex(entry => entry.toLowerCase() === start.toLowerCase()));
  return {
    sequence,
    index,
    targetIndex: slot(sequence, target),
    character: start,
    previous: start,
    flips: 0,
    final: false,
    settled: start === target,
    period: 70 + Math.random() * 15,
    nextAt: entranceMs + position * entranceSpreadMs + Math.random() * 60,
  };
}

function board(from: string, to: string): Flap[] {
  const fromCharacters = fit(graphemes(from));
  const toCharacters = fit(graphemes(to));
  const count = Math.max(fromCharacters.length, toCharacters.length, 1);
  const fromOffset = Math.floor((count - fromCharacters.length) / 2);
  const toOffset = Math.floor((count - toCharacters.length) / 2);
  return Array.from({ length: count }, (_, position) =>
    createFlap(fromCharacters[position - fromOffset] ?? blank, toCharacters[position - toOffset] ?? blank, position),
  );
}

function boardWidth(count: number, width: number) {
  return count * width + (count - 1) * flapGap + tileGap + Math.round(width * 40 / 26);
}

const wide = computed(() => boardWidth(flaps.value.length, flapWidth) > narrowStage);

const size = computed(() => {
  const count = flaps.value.length;
  const available = stageWidth.value || (wide.value ? wideStage : narrowStage);
  let width = flapWidth;
  while (width > 12 && boardWidth(count, width) > available) {
    width--;
  }
  const height = Math.round(width * 40 / 26);
  return {
    flap: { width: `${width}px`, height: `${height}px` },
    glyph: { height: `${height}px`, lineHeight: `${height}px`, fontSize: `${Math.round(width * 0.74)}px` },
    tile: { width: `${height}px`, height: `${height}px` },
    face: { borderRadius: `${Math.round(height / 4)}px`, fontSize: `${Math.round(height * 0.45)}px` },
  };
});

function stop() {
  if (frameRequest !== undefined) cancelAnimationFrame(frameRequest);
  clearTimeout(finishTimer);
}

function land() {
  stop();
  tileFlipped.value = true;
  finishTimer = setTimeout(() => emit('finished'), holdMs);
}

function advance(flap: Flap) {
  flap.index = (flap.index + 1) % flap.sequence.length;
  flap.previous = flap.character;
  flap.character = flap.sequence[flap.index] ?? blank;
  flap.flips++;
  flap.final = settling && flap.index === flap.targetIndex;
  flap.settled = flap.final;
}

function frame(now: number) {
  const elapsed = now - startedAt;
  const arrived = props.workspaceSwitch?.arrived ?? false;

  if (reducedMotion.value === 'reduce') {
    if (arrived) {
      for (const flap of flaps.value) {
        flap.character = flap.sequence[flap.targetIndex] ?? blank;
        flap.previous = flap.character;
        flap.settled = true;
      }
      land();
      return;
    }
    frameRequest = requestAnimationFrame(frame);
    return;
  }

  if (!settling && arrived && elapsed >= minimumRollMs) {
    settling = true;
  }

  for (const flap of flaps.value) {
    if (flap.settled || elapsed < flap.nextAt) continue;
    advance(flap);
    flap.nextAt = elapsed + flap.period;
  }

  if (settling && flaps.value.every(flap => flap.settled)) {
    land();
    return;
  }
  frameRequest = requestAnimationFrame(frame);
}

watch(() => props.workspaceSwitch, (workspaceSwitch) => {
  stop();
  if (!workspaceSwitch) return;
  shown.value = workspaceSwitch;
  flaps.value = board(workspaceSwitch.from.name, workspaceSwitch.to.name);
  tileFlipped.value = false;
  settling = false;
  startedAt = performance.now();
  frameRequest = requestAnimationFrame(frame);
}, { immediate: true });

onBeforeUnmount(stop);
</script>

<template>
  <AlertDialog :open="!!workspaceSwitch">
    <AlertDialogContent class="flex flex-col items-center gap-4" :class="wide ? 'sm:max-w-xl' : 'sm:max-w-sm'">
      <AlertDialogTitle class="sr-only">Switching workspace</AlertDialogTitle>
      <div
        v-if="shown"
        ref="stage"
        aria-hidden="true"
        class="flex h-24 w-full items-center justify-center gap-2.5"
      >
        <div class="relative shrink-0 perspective-near" :style="size.tile">
          <div
            class="absolute inset-0 flex items-center justify-center font-semibold backface-hidden transition-transform duration-300 ease-out"
            :class="tileFlipped ? 'rotate-x-180' : 'rotate-x-0'"
            :style="[size.face, workspaceColor(shown.from.id)]"
          >
            {{ initial(shown.from.name) }}
          </div>
          <div
            class="absolute inset-0 flex items-center justify-center font-semibold backface-hidden transition-transform duration-300 ease-out"
            :class="tileFlipped ? 'rotate-x-0' : '-rotate-x-180'"
            :style="[size.face, workspaceColor(shown.to.id)]"
          >
            {{ initial(shown.to.name) }}
          </div>
        </div>
        <div class="board flex gap-1">
          <div
            v-for="(flap, position) in flaps"
            :key="position"
            class="unit relative shrink-0 rounded font-sans font-semibold shadow-sm"
            :style="size.flap"
          >
            <span class="half top"><span class="glyph" :style="size.glyph">{{ flap.character }}</span></span>
            <span class="half bottom"><span class="glyph" :style="size.glyph">{{ flap.previous }}</span></span>
            <template v-if="flap.flips">
              <span :key="`shade-${flap.flips}`" class="half bottom shade" />
              <span :key="`fall-${flap.flips}`" class="half top fall">
                <span class="glyph" :style="size.glyph">{{ flap.previous }}</span>
              </span>
              <span :key="`land-${flap.flips}`" class="half bottom land" :class="{ final: flap.final }">
                <span class="glyph" :style="size.glyph">{{ flap.character }}</span>
              </span>
            </template>
          </div>
        </div>
      </div>
      <AlertDialogDescription class="text-center text-sm wrap-anywhere">
        Switching to <span class="font-medium text-foreground">{{ shown?.to.name }}</span>…
      </AlertDialogDescription>
    </AlertDialogContent>
  </AlertDialog>
</template>

<style scoped>
.board {
  --flap: var(--foreground);
  --glyph: var(--background);
}

.dark .board {
  --flap: var(--background);
  --glyph: var(--foreground);
}

.unit {
  perspective: 240px;
  background: color-mix(in oklab, var(--flap) 55%, black);
}

.half {
  position: absolute;
  right: 0;
  left: 0;
  height: calc(50% - 0.5px);
  overflow: hidden;
  color: var(--glyph);
  backface-visibility: hidden;
}

.top {
  top: 0;
  border-radius: 0.25rem 0.25rem 1px 1px;
  background: linear-gradient(to bottom, color-mix(in oklab, var(--flap) 86%, white), var(--flap));
}

.bottom {
  bottom: 0;
  border-radius: 1px 1px 0.25rem 0.25rem;
  background: linear-gradient(to bottom, var(--flap), color-mix(in oklab, var(--flap) 88%, black));
}

.glyph {
  position: absolute;
  right: 0;
  left: 0;
  text-align: center;
  white-space: nowrap;
}

.top .glyph {
  top: 0;
}

.bottom .glyph {
  bottom: 0;
}

.fall::after,
.land::after {
  position: absolute;
  inset: 0;
  background: black;
  content: '';
}

.shade {
  z-index: 1;
  background: black;
  opacity: 0;
  animation: shade 70ms ease-in-out forwards;
}

.fall {
  z-index: 2;
  transform-origin: 50% 100%;
  animation: fall 34ms ease-in forwards;
}

.fall::after {
  opacity: 0;
  animation: darken 34ms ease-in forwards;
}

.land {
  z-index: 2;
  transform: rotateX(90deg);
  transform-origin: 50% 0;
  animation: land 34ms ease-out 34ms forwards;
}

.land::after {
  opacity: 0.35;
  animation: lighten 34ms ease-out 34ms forwards;
}

.land.final {
  animation: settle 260ms cubic-bezier(0.3, 0, 0.4, 1) 34ms forwards;
}

.land.final::after {
  animation-duration: 140ms;
}

@keyframes fall {
  99% {
    opacity: 1;
  }

  to {
    opacity: 0;
    transform: rotateX(-90deg);
  }
}

@keyframes darken {
  to {
    opacity: 0.45;
  }
}

@keyframes land {
  to {
    transform: rotateX(0deg);
  }
}

@keyframes settle {
  45% {
    transform: rotateX(0deg);
  }

  65% {
    transform: rotateX(16deg);
  }

  82% {
    transform: rotateX(-3deg);
  }

  to {
    transform: rotateX(0deg);
  }
}

@keyframes lighten {
  to {
    opacity: 0;
  }
}

@keyframes shade {
  50% {
    opacity: 0.3;
  }

  to {
    opacity: 0;
  }
}
</style>
