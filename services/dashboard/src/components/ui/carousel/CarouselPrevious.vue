<script setup lang="ts">
import { type HTMLAttributes, inject } from 'vue';
import { ArrowLeft } from 'lucide-vue-next';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
const carousel = inject('carousel') as any;
</script>

<template>
  <Button
    :disabled="!carousel?.canScrollPrev?.value"
    :class="cn('absolute h-8 w-8 rounded-full', carousel?.orientation?.value === 'vertical' ? '-top-12 left-1/2 -translate-x-1/2 rotate-90' : '-left-12 top-1/2 -translate-y-1/2', props.class)"
    variant="outline"
    size="icon"
    @click="carousel?.scrollPrev"
  >
    <ArrowLeft class="h-4 w-4" />
    <span class="sr-only">Previous slide</span>
  </Button>
</template>
