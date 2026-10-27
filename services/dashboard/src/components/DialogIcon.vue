<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  icon: Component;
  destructive?: boolean;
}>();
</script>

<template>
  <span
    class="mb-2 flex size-10 items-center justify-center self-center rounded-[11px] border bg-card shadow-sm sm:self-start"
    :class="destructive ? 'text-destructive' : 'text-foreground'"
  >
    <component :is="icon" :size="18" />
  </span>
</template>
