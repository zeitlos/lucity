<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <thead :class="cn('bg-muted/40 [&_tr]:border-b [&_tr]:hover:bg-transparent', props.class)">
    <slot />
  </thead>
</template>
