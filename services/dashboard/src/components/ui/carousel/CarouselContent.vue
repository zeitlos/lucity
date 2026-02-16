<script setup lang="ts">
import { type HTMLAttributes, inject } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
const carousel = inject('carousel') as any;
</script>

<template>
  <div ref="carousel?.emblaNode" class="overflow-hidden">
    <div :class="cn('flex', carousel?.orientation?.value === 'vertical' && '-mt-4 flex-col', props.class)">
      <slot />
    </div>
  </div>
</template>
