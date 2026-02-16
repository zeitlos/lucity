<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ToastTitle, type ToastTitleProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ToastTitleProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <ToastTitle :class="cn('text-sm font-semibold', props.class)">
    <slot />
  </ToastTitle>
</template>
