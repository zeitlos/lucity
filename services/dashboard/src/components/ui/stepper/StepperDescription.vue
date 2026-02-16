<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperDescription, type StepperDescriptionProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperDescriptionProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <StepperDescription :class="cn('text-xs text-muted-foreground', props.class)">
    <slot />
  </StepperDescription>
</template>
