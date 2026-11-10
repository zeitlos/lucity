<script setup lang="ts">
defineProps<{
  tint: 'joy' | 'grow' | 'keep';
}>();
</script>

<template>
  <div class="stage" :class="`stage-${tint}`">
    <slot />
  </div>
</template>

<style scoped>
.stage {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  color: oklch(0.35 0.03 55);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .14) 1px, transparent 0);
  background-size: 22px 22px;
  box-shadow: inset 0 0 0 1px oklch(0.25 0.04 55 / .07);
}

.stage-joy {
  background-color: oklch(0.925 0.028 300);
}

.stage-grow {
  background-color: oklch(0.93 0.038 68);
}

.stage-keep {
  background-color: oklch(0.928 0.026 8);
}
</style>
