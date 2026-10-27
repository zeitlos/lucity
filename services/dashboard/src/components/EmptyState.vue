<script setup lang="ts">
import type { Component } from 'vue';
import PatternTopo from '@/components/PatternTopo.vue';

withDefaults(defineProps<{
  icon?: Component;
  title: string;
  description?: string;
  pattern?: 'dots' | 'diagonal' | 'crosshatch' | 'iso-dots' | 'plus' | 'circles' | 'topo';
  size?: 'page' | 'compact';
  tone?: 'neutral' | 'ok';
}>(), {
  icon: undefined,
  description: undefined,
  pattern: undefined,
  size: 'page',
  tone: 'neutral',
});

const patternClass: Record<string, string> = {
  dots: 'pattern-dots',
  diagonal: 'pattern-diagonal',
  crosshatch: 'pattern-crosshatch',
  'iso-dots': 'pattern-iso-dots',
  plus: 'pattern-plus',
  circles: 'pattern-circles',
};
</script>

<template>
  <div
    v-if="size === 'compact'"
    class="flex flex-col items-center gap-1 rounded-lg border border-dashed px-6 py-8 text-center"
  >
    <span
      v-if="icon"
      class="mb-2 flex size-10 items-center justify-center rounded-full"
      :class="tone === 'ok' ? 'bg-status-ok/10 text-status-ok' : 'bg-muted text-muted-foreground'"
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
    :class="[
      'overflow-hidden rounded-lg border border-border',
      pattern ? 'bg-card' : '',
    ]"
  >
    <div
      :class="[
        'flex flex-col items-center justify-center px-8 py-18 text-center',
        pattern && pattern !== 'topo' ? ['relative', patternClass[pattern]] : '',
        pattern === 'topo' ? 'relative' : '',
      ]"
    >
      <PatternTopo v-if="pattern === 'topo'" />

      <div v-if="icon" class="relative mb-4 rounded-full bg-muted p-4">
        <component :is="icon" :size="32" class="text-muted-foreground" />
      </div>
      <h2 class="relative font-serif text-4xl leading-tight tracking-tight text-foreground">
        {{ title }}
      </h2>
      <p
        v-if="description"
        class="relative mt-6 mb-10 max-w-md text-base leading-relaxed text-muted-foreground"
      >
        {{ description }}
      </p>
      <div v-if="$slots.action" :class="['relative', description ? '' : 'mt-6']">
        <slot name="action" />
      </div>
    </div>
  </div>
</template>
