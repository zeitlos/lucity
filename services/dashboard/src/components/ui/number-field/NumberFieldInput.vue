<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { NumberFieldInput, type NumberFieldInputProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<NumberFieldInputProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <NumberFieldInput
    v-bind="props"
    :class="cn('flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-center tabular-nums ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50', props.class)"
  />
</template>
