<script setup lang="ts">
import { computed } from 'vue';
import type { StatusGlyph } from '@/lib/status';

const props = defineProps<{
  glyph: StatusGlyph;
  glow?: boolean;
}>();

interface Cell {
  lit: boolean;
  motion?: string;
  delay?: string;
}

const patterns: Partial<Record<StatusGlyph, string>> = {
  online: '111111111',
  degraded: '111110100',
  failed: '101010101',
  stopped: '101101101',
  idle: '000111000',
  dot: '000010000',
  unknown: '010101010',
};

const perimeter = [0, 1, 2, 5, 8, 7, 6, 3];

function snake(): Cell[] {
  const cells: Cell[] = Array.from({ length: 9 }, () => ({ lit: false }));
  const step = 1.2 / perimeter.length;

  perimeter.forEach((cell, position) => {
    cells[cell] = {
      lit: position === 0 || position === perimeter.length - 1,
      motion: 'trail',
      delay: `${-((perimeter.length - position) % perimeter.length) * step}s`,
    };
  });

  return cells;
}

function ripple(): Cell[] {
  return Array.from({ length: 9 }, (_, cell) => {
    const center = cell === 4;

    return { lit: center, motion: 'ripple', delay: center ? '0s' : '-0.75s' };
  });
}

function build(): Cell[] {
  return Array.from({ length: 9 }, (_, cell) => {
    const level = 2 - Math.floor(cell / 3);

    return { lit: level < 2, motion: `build-${level}` };
  });
}

const cells = computed<Cell[]>(() => {
  switch (props.glyph) {
    case 'snake':
      return snake();
    case 'ripple':
      return ripple();
    case 'build':
      return build();
    default:
      return [...(patterns[props.glyph] ?? patterns.dot!)].map(bit => ({ lit: bit === '1' }));
  }
});
</script>

<template>
  <span aria-hidden="true" :class="['status-glyph grid shrink-0 grid-cols-3 gap-px', glow && 'status-glyph-glow']">
    <span
      v-for="(cell, index) in cells"
      :key="`${glyph}-${index}`"
      :class="['size-0.5', cell.lit && 'lit', cell.motion]"
      :style="cell.delay ? { animationDelay: cell.delay } : undefined"
    />
  </span>
</template>

<style scoped>
.status-glyph > span {
  background-color: color-mix(in oklab, currentColor 22%, transparent);
}

.status-glyph > span.lit {
  background-color: currentColor;
}

.status-glyph-glow {
  filter: drop-shadow(0 0 2px color-mix(in oklab, currentColor 60%, transparent));
}

.status-glyph > .trail {
  animation: status-trail 1.2s linear infinite;
}

.status-glyph > .ripple {
  animation: status-ripple 1.5s linear infinite;
}

.status-glyph > .build-0 {
  animation: status-build-0 2.4s linear infinite;
}

.status-glyph > .build-1 {
  animation: status-build-1 2.4s linear infinite;
}

.status-glyph > .build-2 {
  animation: status-build-2 2.4s linear infinite;
}

@keyframes status-trail {
  0% { background-color: currentColor; }
  35%, 100% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
}

@keyframes status-ripple {
  0% { background-color: currentColor; }
  40%, 100% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
}

@keyframes status-build-0 {
  0%, 2% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
  8%, 85% { background-color: currentColor; }
  92%, 100% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
}

@keyframes status-build-1 {
  0%, 32% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
  38%, 85% { background-color: currentColor; }
  92%, 100% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
}

@keyframes status-build-2 {
  0%, 62% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
  68%, 85% { background-color: currentColor; }
  92%, 100% { background-color: color-mix(in oklab, currentColor 22%, transparent); }
}

@media (prefers-reduced-motion: reduce) {
  .status-glyph > span {
    animation: none;
  }
}
</style>
