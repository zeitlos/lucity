<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  icon: Component;
}>();
</script>

<template>
  <span class="flex size-[26px] shrink-0 items-center justify-center rounded-[6px] bg-muted/70 text-muted-foreground ring-1 ring-inset ring-border/70">
    <component :is="icon" class="!size-3.5" :stroke-width="1.75" />
  </span>
</template>
