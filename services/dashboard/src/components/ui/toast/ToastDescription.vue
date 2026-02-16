<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ToastDescription, type ToastDescriptionProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ToastDescriptionProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <ToastDescription :class="cn('text-sm opacity-90', props.class)">
    <slot />
  </ToastDescription>
</template>
