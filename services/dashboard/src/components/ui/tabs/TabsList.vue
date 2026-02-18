<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { TabsList, type TabsListProps } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<TabsListProps & { class?: HTMLAttributes['class'] }>();
</script>

<template>
  <TabsList
    v-bind="props"
    :class="cn('inline-flex h-10 items-center gap-4 border-b text-muted-foreground', props.class)"
  >
    <slot />
  </TabsList>
</template>
