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
      'overflow-hidden rounded-lg border border-border',
      pattern ? 'bg-card' : '',
    ]"
  >
    <div
      :class="[
        'flex flex-col items-center justify-center px-8 py-[4.5rem] text-center',
        pattern && pattern !== 'topo' ? ['relative', patternClass[pattern]] : '',
        pattern === 'topo' ? 'relative' : '',
      ]"
    >
      <PatternTopo v-if="pattern === 'topo'" />

      <div v-if="icon" class="relative mb-4 rounded-full bg-muted p-4">
        <component :is="icon" :size="32" class="text-muted-foreground" />
      </div>
      <h2 class="relative font-serif text-[3.2rem] leading-[1.1] tracking-[-0.02em] text-foreground">
        {{ title }}
      </h2>
      <p
        v-if="description"
        class="relative mt-6 mb-10 max-w-[460px] text-base leading-[1.75] text-muted-foreground"
      >
        {{ description }}
      </p>
      <div v-if="$slots.action" :class="['relative', description ? '' : 'mt-6']">
        <slot name="action" />
      </div>
    </div>
  </div>
</template>
