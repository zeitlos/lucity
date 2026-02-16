<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ComboboxContent, type ComboboxContentProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ComboboxContentProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <ComboboxContent v-bind="props" :class="cn('max-h-[300px] overflow-y-auto overflow-x-hidden', props.class)">
    <slot />
  </ComboboxContent>
</template>
