<script setup lang="ts">
import { type HTMLAttributes, inject } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
const carousel = inject('carousel') as any;
</script>

<template>
  <div
    :class="cn('min-w-0 shrink-0 grow-0 basis-full', carousel?.orientation?.value === 'vertical' ? 'pt-4' : 'pl-4', props.class)"
    role="group"
    aria-roledescription="slide"
  >
    <slot />
  </div>
</template>
