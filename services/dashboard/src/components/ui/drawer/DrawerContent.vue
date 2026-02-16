<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { type DrawerContentProps, DrawerContent, DrawerOverlay, DrawerPortal } from 'vaul-vue';
import { cn } from '@/lib/utils';

const props = defineProps<DrawerContentProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <DrawerPortal>
    <DrawerOverlay class="fixed inset-0 z-50 bg-black/80" />
    <DrawerContent
      v-bind="props"
      :class="cn('fixed inset-x-0 bottom-0 z-50 mt-24 flex h-auto flex-col rounded-t-[10px] border bg-background', props.class)"
    >
      <div class="mx-auto mt-4 h-2 w-[100px] rounded-full bg-muted" />
      <slot />
    </DrawerContent>
  </DrawerPortal>
</template>
