<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <nav :class="cn('', props.class)" aria-label="breadcrumb">
    <slot />
  </nav>
</template>
