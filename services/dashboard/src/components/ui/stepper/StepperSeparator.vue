<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperSeparator, type StepperSeparatorProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperSeparatorProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <StepperSeparator :class="cn('h-px flex-1 bg-muted data-[state=completed]:bg-primary', props.class)" />
</template>
