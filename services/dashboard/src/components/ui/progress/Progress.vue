<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ProgressIndicator, ProgressRoot, type ProgressRootEmits, type ProgressRootProps, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ProgressRootProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<ProgressRootEmits>();

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <ProgressRoot
    v-bind="forwarded"
    :class="cn('relative h-4 w-full overflow-hidden rounded-full bg-secondary', props.class)"
  >
    <ProgressIndicator
      class="h-full w-full flex-1 bg-primary transition-all"
      :style="`transform: translateX(-${100 - (props.modelValue ?? 0)}%)`"
    />
  </ProgressRoot>
</template>
