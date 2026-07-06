<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const statusVariants = cva(
  'inline-flex items-center gap-1.5 rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs font-medium text-muted-foreground transition-colors',
  {
    variants: {
      tone: {
        ok: '[&_.status-dot]:bg-[var(--status-ok)] [&_.status-dot]:shadow-[0_0_6px_var(--status-ok)]',
        warn: '[&_.status-dot]:bg-[var(--status-warn)] [&_.status-dot]:shadow-[0_0_6px_var(--status-warn)]',
        danger: '[&_.status-dot]:bg-[var(--status-danger)] [&_.status-dot]:shadow-[0_0_6px_var(--status-danger)]',
        progress: '[&_.status-dot]:bg-[var(--status-progress)] [&_.status-dot]:shadow-[0_0_6px_var(--status-progress)]',
        neutral: '[&_.status-dot]:bg-[var(--status-neutral)]',
      },
    },
    defaultVariants: {
      tone: 'neutral',
    },
  },
);

type StatusVariants = VariantProps<typeof statusVariants>;
</script>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  tone?: StatusVariants['tone'];
}>();
</script>

<template>
  <div :class="cn(statusVariants({ tone }), props.class)">
    <span class="status-dot h-[7px] w-[7px] shrink-0 rounded-full" />
    <slot />
  </div>
</template>
