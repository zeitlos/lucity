<script setup lang="ts">
import { type HTMLAttributes, inject } from 'vue';
import { ArrowRight } from 'lucide-vue-next';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
const carousel = inject('carousel') as any;
</script>

<template>
  <Button
    :disabled="!carousel?.canScrollNext?.value"
    :class="cn('absolute h-8 w-8 rounded-full', carousel?.orientation?.value === 'vertical' ? '-bottom-12 left-1/2 -translate-x-1/2 rotate-90' : '-right-12 top-1/2 -translate-y-1/2', props.class)"
    variant="outline"
    size="icon"
    @click="carousel?.scrollNext"
  >
    <ArrowRight class="h-4 w-4" />
    <span class="sr-only">Next slide</span>
  </Button>
</template>
