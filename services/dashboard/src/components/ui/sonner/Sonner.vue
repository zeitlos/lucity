<script setup lang="ts">
import { Toaster as Sonner } from 'vue-sonner';

const props = defineProps<{
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center';
  expand?: boolean;
  theme?: 'light' | 'dark' | 'system';
}>();
</script>

<template>
  <Sonner
    class="toaster group"
    v-bind="props"
    :toast-options="{
      class: 'w-(--width) rounded-xl border bg-card p-1.5 text-card-foreground shadow-lg',
    }"
  />
</template>

<style>
[data-sonner-toast][data-expanded='false'][data-front='false'] > * {
  opacity: 0;
}
</style>
