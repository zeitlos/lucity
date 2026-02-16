<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperIndicator, type StepperIndicatorProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperIndicatorProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <StepperIndicator :class="cn(props.class)">
    <slot />
  </StepperIndicator>
</template>
