<script setup lang="ts">
import { cn } from '@/lib/utils';
import { WORDMARK_CAP, WORDMARK_PATH, WORDMARK_WIDTH } from '@/lib/logo';

const props = withDefaults(defineProps<{
  size?: number;
  class?: string;
}>(), {
  size: 26,
});
</script>

<template>
  <svg
    :width="Math.round(props.size * WORDMARK_WIDTH / WORDMARK_CAP)"
    :height="props.size"
    :viewBox="`0 -${WORDMARK_CAP} ${WORDMARK_WIDTH} ${WORDMARK_CAP}`"
    :class="cn('inline-block shrink-0 overflow-visible fill-current', props.class)"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Lucity"
  >
    <path :d="WORDMARK_PATH" />
  </svg>
</template>
