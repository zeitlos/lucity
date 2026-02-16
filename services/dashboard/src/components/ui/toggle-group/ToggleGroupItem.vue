<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { ToggleGroupItem, type ToggleGroupItemProps } from 'reka-ui';
import { type VariantProps } from 'class-variance-authority';
import { toggleVariants } from '@/components/ui/toggle';
import { cn } from '@/lib/utils';
import { inject } from 'vue';

type ToggleGroupVariants = VariantProps<typeof toggleVariants>;

const props = defineProps<ToggleGroupItemProps & {
  class?: HTMLAttributes['class'];
  variant?: ToggleGroupVariants['variant'];
  size?: ToggleGroupVariants['size'];
}>();

const context = inject<{ variant?: string; size?: string }>('toggleGroup', {});
</script>

<template>
  <ToggleGroupItem
    v-bind="props"
    :class="cn(toggleVariants({ variant: variant || context.variant as any, size: size || context.size as any }), props.class)"
  >
    <slot />
  </ToggleGroupItem>
</template>
