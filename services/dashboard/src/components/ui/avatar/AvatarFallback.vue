<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { AvatarFallback, type AvatarFallbackProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<AvatarFallbackProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <AvatarFallback :class="cn('flex h-full w-full items-center justify-center rounded-full bg-muted', props.class)">
    <slot />
  </AvatarFallback>
</template>
