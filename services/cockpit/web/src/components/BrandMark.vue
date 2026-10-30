<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';

import { markPath } from '@/lib/mark';

type Spring = { x: number; v: number; target: number };

const props = defineProps<{
  loading: boolean;
}>();

const SWING = 1.1;

const rest = markPath({});
const path = ref(rest);
const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const sway: Spring = { x: 1, v: 0, target: 1 };
let since = SWING;
let last = 0;
let frame = 0;

function spring(state: Spring, dt: number, stiffness: number, damping: number) {
  state.v += (stiffness * (state.target - state.x) - 2 * damping * Math.sqrt(stiffness) * state.v) * dt;
  state.x += state.v * dt;
}

function settled(state: Spring): boolean {
  return Math.abs(state.target - state.x) < 0.001 && Math.abs(state.v) < 0.001;
}

function tick(now: number) {
  const dt = Math.min(0.032, (now - last) / 1000);

  last = now;
  since += dt;

  if (props.loading && since >= SWING) {
    since = 0;
    sway.target = -sway.target;
  }

  if (!props.loading) sway.target = 1;

  spring(sway, dt, 90, 0.58);

  if (!props.loading && settled(sway)) {
    path.value = rest;
    frame = 0;
    since = SWING;
    return;
  }

  const bend = Math.abs(sway.x);

  path.value = markPath({
    bw: 60 * bend,
    flat: 10 * bend,
    dT: 144.5 - 24.5 * bend,
    dB: 144.5 + 24.5 * bend,
  }, sway.x < 0);

  frame = requestAnimationFrame(tick);
}

watch(() => props.loading, (loading) => {
  if (!loading || frame || reducedMotion) return;

  last = performance.now();
  frame = requestAnimationFrame(tick);
}, { immediate: true });

onBeforeUnmount(() => cancelAnimationFrame(frame));
</script>

<template>
  <svg viewBox="-20 0 290 181" class="-mx-0.5 h-[22px] w-[35px] shrink-0" aria-hidden="true">
    <path :d="path" class="fill-brand-mark" :class="loading && reducedMotion ? 'opacity-50' : ''" />
  </svg>
</template>
