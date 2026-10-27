<script setup lang="ts">
import type { Component } from 'vue';
import { cn } from '@/lib/utils';

withDefaults(defineProps<{
  icon?: Component;
  title: string;
  description?: string;
  size?: 'page' | 'compact';
  tone?: 'neutral' | 'ok';
}>(), {
  icon: undefined,
  description: undefined,
  size: 'page',
  tone: 'neutral',
});
</script>

<template>
  <div
    v-if="size === 'compact'"
    class="flex flex-col items-center gap-1 rounded-lg border border-dashed px-6 py-8 text-center"
  >
    <span
      v-if="icon"
      :class="cn(
        'mb-2 flex size-10 items-center justify-center rounded-md border bg-card shadow-sm',
        tone === 'ok' ? 'border-status-ok/20 text-status-ok' : 'text-muted-foreground',
      )"
    >
      <component :is="icon" :size="18" />
    </span>
    <p class="text-sm font-medium text-foreground">{{ title }}</p>
    <p v-if="description || $slots.description" class="max-w-sm text-sm text-muted-foreground">
      <slot name="description">{{ description }}</slot>
    </p>
    <div v-if="$slots.action" class="mt-3">
      <slot name="action" />
    </div>
  </div>

  <div
    v-else
    class="pattern-crosshatch flex flex-col items-center justify-center rounded-lg border bg-card px-8 py-18 text-center"
  >
    <h2 class="font-serif text-4xl leading-tight tracking-tight text-foreground">
      {{ title }}
    </h2>
    <p
      v-if="description"
      class="mt-6 mb-10 max-w-md text-base leading-relaxed text-muted-foreground"
    >
      {{ description }}
    </p>
    <div v-if="$slots.action" :class="description ? '' : 'mt-6'">
      <slot name="action" />
    </div>
  </div>
</template>
