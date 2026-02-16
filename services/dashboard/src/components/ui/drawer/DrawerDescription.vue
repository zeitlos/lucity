<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { type DrawerDescriptionProps, DrawerDescription } from 'vaul-vue';
import { cn } from '@/lib/utils';

const props = defineProps<DrawerDescriptionProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <DrawerDescription :class="cn('text-sm text-muted-foreground', props.class)">
    <slot />
  </DrawerDescription>
</template>
