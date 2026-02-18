<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { TabsTrigger, type TabsTriggerProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<TabsTriggerProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <TabsTrigger
    v-bind="props"
    :class="cn('inline-flex items-center justify-center whitespace-nowrap px-1 pb-2.5 pt-2 text-sm font-medium transition-all border-b-2 border-transparent -mb-px focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:border-primary data-[state=active]:text-foreground', props.class)"
  >
    <slot />
  </TabsTrigger>
</template>
