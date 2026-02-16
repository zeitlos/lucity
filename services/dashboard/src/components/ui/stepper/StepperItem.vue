<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperItem, type StepperItemProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperItemProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <StepperItem
    v-bind="props"
    :class="cn('flex items-center gap-2 data-[disabled]:pointer-events-none data-[disabled]:opacity-50', props.class)"
  >
    <slot />
  </StepperItem>
</template>
