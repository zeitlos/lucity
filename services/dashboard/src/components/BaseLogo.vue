<script setup lang="ts">
import { cn } from '@/lib/utils';
import { LOGO_HEIGHT, LOGO_PATH, LOGO_WIDTH } from '@/lib/logo';

const props = withDefaults(defineProps<{
  size?: number;
  class?: string;
}>(), {
  size: 40,
});
</script>

<template>
  <svg
    :width="Math.round(props.size * LOGO_WIDTH / LOGO_HEIGHT)"
    :height="props.size"
    :viewBox="`0 0 ${LOGO_WIDTH} ${LOGO_HEIGHT}`"
    :class="cn('lucity-mark inline-block shrink-0', props.class)"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Lucity logo"
  >
    <path :d="LOGO_PATH" />
  </svg>
</template>

<style scoped>
.lucity-mark {
  fill: var(--brand-mark);
}
</style>
