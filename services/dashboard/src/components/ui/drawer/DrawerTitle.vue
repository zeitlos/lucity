<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { type DrawerTitleProps, DrawerTitle } from 'vaul-vue';
import { cn } from '@/lib/utils';

const props = defineProps<DrawerTitleProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <DrawerTitle :class="cn('text-lg font-semibold leading-none tracking-tight', props.class)">
    <slot />
  </DrawerTitle>
</template>
