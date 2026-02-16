<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NumberFieldRoot, type NumberFieldRootEmits, type NumberFieldRootProps, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<NumberFieldRootProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<NumberFieldRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <NumberFieldRoot v-bind="forwarded" :class="cn('flex items-center', props.class)">
    <slot />
  </NumberFieldRoot>
</template>
