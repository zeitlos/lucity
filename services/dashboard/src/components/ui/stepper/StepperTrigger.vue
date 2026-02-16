<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperTrigger, type StepperTriggerProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperTriggerProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <StepperTrigger
    v-bind="props"
    :class="cn('inline-flex h-10 w-10 items-center justify-center rounded-full border-2 border-muted bg-background text-sm font-medium data-[state=active]:border-primary data-[state=completed]:border-primary data-[state=completed]:bg-primary data-[state=completed]:text-primary-foreground', props.class)"
  >
    <slot />
  </StepperTrigger>
</template>
