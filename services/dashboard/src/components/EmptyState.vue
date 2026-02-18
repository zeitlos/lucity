<script setup lang="ts">
import type { Component } from 'vue';

defineProps<{
  icon: Component;
  title: string;
  description?: string;
}>();
</script>

<template>
  <div class="flex flex-col items-center justify-center rounded-lg border border-dashed px-8 py-16">
    <div class="mb-4 rounded-full bg-muted p-4">
      <component :is="icon" :size="32" class="text-muted-foreground" />
    </div>
    <h2 class="text-lg font-medium text-foreground">{{ title }}</h2>
    <p v-if="description" class="mt-1 mb-6 max-w-sm text-center text-sm text-muted-foreground">
      {{ description }}
    </p>
    <div v-if="$slots.action" :class="description ? '' : 'mt-4'">
      <slot name="action" />
    </div>
  </div>
</template>
