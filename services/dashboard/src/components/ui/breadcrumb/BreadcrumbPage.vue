<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <span :class="cn('font-normal text-foreground', props.class)" role="link" aria-disabled="true" aria-current="page">
    <slot />
  </span>
</template>
