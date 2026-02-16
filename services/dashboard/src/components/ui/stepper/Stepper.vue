<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { StepperRoot, type StepperRootEmits, type StepperRootProps, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<StepperRootProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<StepperRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <StepperRoot v-bind="forwarded" :class="cn('flex gap-2', props.class)">
    <slot />
  </StepperRoot>
</template>
