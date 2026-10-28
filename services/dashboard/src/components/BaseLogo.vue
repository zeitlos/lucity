<script setup lang="ts">
import { computed } from 'vue';
import { cn } from '@/lib/utils';
import { LOGO_HEIGHT, LOGO_PATH, LOGO_WIDTH, logoPath } from '@/lib/logo';

const props = withDefaults(defineProps<{
  size?: number;
  flex?: number;
  class?: string;
}>(), {
  size: 40,
  flex: 0,
});

const path = computed(() => (props.flex ? logoPath(props.flex) : LOGO_PATH));
</script>

<template>
  <svg
    :width="Math.round(props.size * LOGO_WIDTH / LOGO_HEIGHT)"
    :height="props.size"
    :viewBox="`0 0 ${LOGO_WIDTH} ${LOGO_HEIGHT}`"
    :class="cn('lucity-mark inline-block shrink-0 overflow-visible', props.class)"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="Lucity logo"
  >
    <path :d="path" />
  </svg>
</template>

<style scoped>
.lucity-mark {
  fill: var(--brand-mark);
}
</style>
