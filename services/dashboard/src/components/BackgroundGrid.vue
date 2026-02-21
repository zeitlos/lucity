<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';

const cursorEl = ref<HTMLElement | null>(null);
let cursorVisible = false;
let rafId: number | null = null;

function onMouseMove(e: MouseEvent) {
  if (!cursorEl.value) return;
  if (!cursorVisible) {
    cursorEl.value.style.opacity = '1';
    cursorVisible = true;
  }
  if (rafId) cancelAnimationFrame(rafId);
  rafId = requestAnimationFrame(() => {
    cursorEl.value?.style.setProperty('--cursor-x', `${e.clientX}px`);
    cursorEl.value?.style.setProperty('--cursor-y', `${e.clientY}px`);
  });
}

function onMouseLeave() {
  if (cursorEl.value) {
    cursorEl.value.style.opacity = '0';
    cursorVisible = false;
  }
}

onMounted(() => {
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseleave', onMouseLeave);
});

onUnmounted(() => {
  document.removeEventListener('mousemove', onMouseMove);
  document.removeEventListener('mouseleave', onMouseLeave);
  if (rafId) cancelAnimationFrame(rafId);
});
</script>

<template>
  <div class="bg-grid" aria-hidden="true" />
  <div ref="cursorEl" class="bg-grid-cursor" aria-hidden="true" />
</template>

<style scoped>
.bg-grid {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.55 0.03 80 / 0.25) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.60 0.02 80 / 0.12) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.12) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.5) 30%, transparent 55%);
  -webkit-mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.5) 30%, transparent 55%);
}

/* Sub-grid */
.bg-grid::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(oklch(0.60 0.02 80 / 0.06) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.06) 1px, transparent 1px);
  background-size: 12px 12px;
  mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.3) 20%, transparent 40%);
  -webkit-mask-image: linear-gradient(180deg, black 0%, rgb(0 0 0 / 0.3) 20%, transparent 40%);
}

/* Cursor spotlight — major grid */
.bg-grid-cursor {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.4s ease;
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.55 0.03 80 / 0.30) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.60 0.02 80 / 0.14) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.14) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(circle 500px at var(--cursor-x, -999px) var(--cursor-y, -999px), rgb(0 0 0 / 0.5) 0%, rgb(0 0 0 / 0.2) 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(circle 500px at var(--cursor-x, -999px) var(--cursor-y, -999px), rgb(0 0 0 / 0.5) 0%, rgb(0 0 0 / 0.2) 40%, transparent 100%);
}

/* Cursor spotlight — sub-grid */
.bg-grid-cursor::after {
  content: '';
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(oklch(0.60 0.02 80 / 0.07) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.60 0.02 80 / 0.07) 1px, transparent 1px);
  background-size: 12px 12px;
  mask-image: radial-gradient(circle 350px at var(--cursor-x, -999px) var(--cursor-y, -999px), rgb(0 0 0 / 0.4) 0%, rgb(0 0 0 / 0.15) 40%, transparent 100%);
  -webkit-mask-image: radial-gradient(circle 350px at var(--cursor-x, -999px) var(--cursor-y, -999px), rgb(0 0 0 / 0.4) 0%, rgb(0 0 0 / 0.15) 40%, transparent 100%);
}

/* Dark mode — brighter grid lines on dark backgrounds */
:global(.dark) .bg-grid {
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.80 0.03 80 / 0.20) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.80 0.02 80 / 0.08) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.08) 1px, transparent 1px);
}
:global(.dark) .bg-grid::after {
  background-image:
    linear-gradient(oklch(0.80 0.02 80 / 0.04) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.04) 1px, transparent 1px);
}
:global(.dark) .bg-grid-cursor {
  background-image:
    radial-gradient(circle 1.5px at 0 0, oklch(0.80 0.03 80 / 0.25) 1.5px, transparent 1.5px),
    linear-gradient(oklch(0.80 0.02 80 / 0.10) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.10) 1px, transparent 1px);
}
:global(.dark) .bg-grid-cursor::after {
  background-image:
    linear-gradient(oklch(0.80 0.02 80 / 0.05) 1px, transparent 1px),
    linear-gradient(90deg, oklch(0.80 0.02 80 / 0.05) 1px, transparent 1px);
}
</style>
