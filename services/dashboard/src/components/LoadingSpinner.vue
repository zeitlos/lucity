<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(defineProps<{ size?: number }>(), { size: 14 });

const borderWidth = computed(() => `${Math.max(1.5, Math.round(props.size / 7))}px`);
</script>

<template>
  <span
    aria-hidden="true"
    class="inline-block shrink-0 animate-spin rounded-full border-current border-t-transparent"
    :style="{ width: `${size}px`, height: `${size}px`, borderWidth }"
  />
</template>
