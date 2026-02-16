<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { PinInputInput, type PinInputInputProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<PinInputInputProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <PinInputInput
    v-bind="props"
    :class="cn('relative h-10 w-10 rounded-md border border-input bg-background text-center text-sm shadow-sm transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50', props.class)"
  />
</template>
