<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <div :class="cn('-mx-6 -mb-6 mt-2 flex flex-col-reverse gap-2 border-t px-5 py-3.5 sm:flex-row sm:justify-end', props.class)">
    <slot />
  </div>
</template>
