<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const avatarVariants = cva(
  'relative flex shrink-0 overflow-hidden rounded-full',
  {
    variants: {
      size: {
        sm: 'h-8 w-8',
        base: 'h-10 w-10',
        lg: 'h-14 w-14',
      },
    },
    defaultVariants: {
      size: 'base',
    },
  },
);

type AvatarVariants = VariantProps<typeof avatarVariants>;
</script>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { AvatarRoot } from 'reka-ui';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  size?: AvatarVariants['size'];
}>();
</script>

<template>
  <AvatarRoot :class="cn(avatarVariants({ size }), props.class)">
    <slot />
  </AvatarRoot>
</template>
