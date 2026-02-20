<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const badgeVariants = cva(
  'inline-flex items-center gap-1.5 rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
  {
    variants: {
      variant: {
        default: '[&_.badge-dot]:bg-[var(--status-ok)] [&_.badge-dot]:shadow-[0_0_6px_var(--status-ok)]',
        secondary: '[&_.badge-dot]:bg-[var(--status-neutral)]',
        destructive: '[&_.badge-dot]:bg-[var(--status-danger)] [&_.badge-dot]:shadow-[0_0_6px_var(--status-danger)]',
        outline: '[&_.badge-dot]:bg-[var(--status-neutral)]',
        warning: '[&_.badge-dot]:bg-[var(--status-warn)] [&_.badge-dot]:shadow-[0_0_6px_var(--status-warn)]',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);

type BadgeVariants = VariantProps<typeof badgeVariants>;
</script>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  variant?: BadgeVariants['variant'];
  hideDot?: boolean;
}>();
</script>

<template>
  <div :class="cn(badgeVariants({ variant }), props.class)">
    <span v-if="!hideDot" class="badge-dot h-[7px] w-[7px] shrink-0 rounded-full" />
    <slot />
  </div>
</template>
