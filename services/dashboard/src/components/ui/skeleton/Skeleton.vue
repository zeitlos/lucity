<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <div aria-hidden="true" :class="cn('skeleton rounded-md', props.class)" />
</template>

<style scoped>
.skeleton {
  --skeleton-shine: var(--card);
  background-color: var(--muted);
  background-image: linear-gradient(
    105deg,
    transparent 32%,
    var(--skeleton-shine) 50%,
    transparent 68%
  );
  background-size: 100vw 100%;
  background-repeat: no-repeat;
  background-attachment: fixed;
  animation:
    skeleton-in 300ms ease-out 100ms both,
    skeleton-shine 2.4s ease-in-out infinite;
}

:global(.dark .skeleton) {
  --skeleton-shine: color-mix(in oklab, var(--foreground) 12%, var(--muted));
  background-color: color-mix(in oklab, var(--foreground) 4%, var(--muted));
}

@keyframes skeleton-in {
  from { opacity: 0; }
}

@keyframes skeleton-shine {
  from { background-position: -70vw 0; }
  65%, to { background-position: 70vw 0; }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton {
    background-image: none;
    animation: none;
  }
}
</style>
