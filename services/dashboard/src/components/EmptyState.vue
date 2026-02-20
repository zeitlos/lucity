<script setup lang="ts">
import type { Component } from 'vue';
import PatternTopo from '@/components/PatternTopo.vue';

defineProps<{
  icon?: Component;
  title: string;
  description?: string;
  pattern?: 'dots' | 'diagonal' | 'crosshatch' | 'iso-dots' | 'plus' | 'circles' | 'topo';
}>();

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
    :class="[
      'flex flex-col items-center justify-center px-8 py-16',
      pattern
        ? ['relative overflow-hidden rounded-lg', pattern !== 'topo' ? patternClass[pattern] : '']
        : 'rounded-lg border border-dashed',
    ]"
  >
    <PatternTopo v-if="pattern === 'topo'" />

    <div v-if="icon" class="relative mb-4 rounded-full bg-muted p-4">
      <component :is="icon" :size="32" class="text-muted-foreground" />
    </div>
    <h2 class="relative font-serif text-3xl text-foreground">{{ title }}</h2>
    <p v-if="description" class="relative mt-1 mb-6 max-w-sm text-center text-sm text-muted-foreground">
      {{ description }}
    </p>
    <div v-if="$slots.action" :class="['relative', description ? '' : 'mt-4']">
      <slot name="action" />
    </div>
  </div>
</template>
