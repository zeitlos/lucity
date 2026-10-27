<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <div :class="cn('mt-0.5 text-sm text-muted-foreground [&_p]:leading-relaxed', props.class)">
    <slot />
  </div>
</template>
