<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NumberFieldDecrement, type NumberFieldDecrementProps } from 'reka-ui';
import { ChevronDown } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

const props = defineProps<NumberFieldDecrementProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <NumberFieldDecrement
    v-bind="props"
    :class="cn('inline-flex items-center justify-center rounded-sm border px-2 py-1 text-sm font-medium ring-offset-background transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50', props.class)"
  >
    <slot>
      <ChevronDown class="h-4 w-4" />
    </slot>
  </NumberFieldDecrement>
</template>
