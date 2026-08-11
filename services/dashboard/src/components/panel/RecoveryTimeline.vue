<script setup lang="ts">
import { computed, ref } from 'vue';

const props = defineProps<{
  from: number;
  to: number;
  markers: number[];
  disabled?: boolean;
}>();

const selected = defineModel<number | null>({ default: null });

const track = ref<HTMLElement | null>(null);
const dragging = ref(false);

const span = computed(() => Math.max(1, props.to - props.from));

function percentFor(time: number): number {
  return Math.min(100, Math.max(0, ((time - props.from) / span.value) * 100));
}

const visibleMarkers = computed(() =>
  props.markers.filter(marker => marker >= props.from && marker <= props.to).map(percentFor),
);

const cursorPercent = computed(() =>
  selected.value === null ? null : percentFor(selected.value),
);

function timeAt(clientX: number): number | null {
  const rect = track.value?.getBoundingClientRect();
  if (!rect || rect.width === 0) return null;
  const fraction = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
  return Math.round(props.from + fraction * span.value);
}

function onPointerDown(event: PointerEvent) {
  if (props.disabled) return;
  dragging.value = true;
  (event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
  selected.value = timeAt(event.clientX);
}

function onPointerMove(event: PointerEvent) {
  if (!dragging.value) return;
  const time = timeAt(event.clientX);
  if (time !== null) selected.value = time;
}

function onPointerUp(event: PointerEvent) {
  if (!dragging.value) return;
  dragging.value = false;
  (event.currentTarget as HTMLElement).releasePointerCapture(event.pointerId);
}

const MINUTE = 60_000;

function nudge(direction: number, event: KeyboardEvent) {
  if (props.disabled) return;
  event.preventDefault();
  const step = event.shiftKey ? MINUTE * 60 : MINUTE;
  const base = selected.value ?? props.to;
  selected.value = Math.min(props.to, Math.max(props.from, base + direction * step));
}

const stripes = `repeating-linear-gradient(-45deg,
  color-mix(in oklch, var(--primary) 22%, transparent) 0 5px,
  color-mix(in oklch, var(--primary) 10%, transparent) 5px 10px)`;

function formatEdge(time: number): string {
  return new Date(time).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}
</script>

<template>
  <div class="space-y-1.5">
    <div class="flex justify-between text-xs text-muted-foreground">
      <span>Earliest</span>
      <span>Latest</span>
    </div>

    <div
      ref="track"
      role="slider"
      tabindex="0"
      :aria-disabled="disabled"
      :aria-valuemin="from"
      :aria-valuemax="to"
      :aria-valuenow="selected ?? undefined"
      aria-label="Restore point"
      class="relative h-8 rounded-md border outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      :class="disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'"
      :style="{
        background: stripes,
        borderColor: 'color-mix(in oklch, var(--primary) 32%, transparent)',
      }"
      @pointerdown="onPointerDown"
      @pointermove="onPointerMove"
      @pointerup="onPointerUp"
      @pointercancel="onPointerUp"
      @keydown.left="nudge(-1, $event)"
      @keydown.right="nudge(1, $event)"
    >
      <div
        v-for="(percent, index) in visibleMarkers"
        :key="`marker-${index}`"
        class="pointer-events-none absolute -top-1 -bottom-1 w-0.5 rounded-full"
        :style="{
          left: `${percent}%`,
          background: 'color-mix(in oklch, var(--primary) 90%, var(--foreground))',
        }"
      />

      <div
        v-if="cursorPercent !== null"
        class="pointer-events-none absolute -top-1.5 -bottom-1.5 w-0.5 rounded-full"
        :style="{ left: `${cursorPercent}%`, background: 'var(--accent-pop)' }"
      >
        <div
          class="absolute -bottom-1 left-1/2 size-2.5 -translate-x-1/2 rounded-full"
          :style="{
            background: 'var(--accent-pop)',
            boxShadow: '0 0 0 3px color-mix(in oklch, var(--accent-pop) 22%, transparent)',
          }"
        />
      </div>
    </div>

    <div class="flex justify-between text-xs tabular-nums text-muted-foreground">
      <span>{{ formatEdge(from) }}</span>
      <span>{{ formatEdge(to) }}</span>
    </div>
  </div>
</template>
