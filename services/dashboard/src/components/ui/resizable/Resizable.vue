<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { SplitterGroup, type SplitterGroupProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<SplitterGroupProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <SplitterGroup v-bind="props" :class="cn('flex h-full w-full data-[panel-group-direction=vertical]:flex-col', props.class)">
    <slot />
  </SplitterGroup>
</template>
