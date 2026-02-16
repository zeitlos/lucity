<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ComboboxGroup, type ComboboxGroupProps, ComboboxLabel } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<ComboboxGroupProps & { class?: HTMLAttributes['class']; heading?: string }>();
</script>

<template>
  <ComboboxGroup v-bind="props" :class="cn('overflow-hidden p-1 text-foreground', props.class)">
    <ComboboxLabel v-if="heading" class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
      {{ heading }}
    </ComboboxLabel>
    <slot />
  </ComboboxGroup>
</template>
