<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { provide, ref } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();

const id = ref(`form-item-${Math.random().toString(36).slice(2, 9)}`);
provide('form-item-id', id);
</script>

<template>
  <div :class="cn('space-y-2', props.class)">
    <slot />
  </div>
</template>
