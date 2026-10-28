<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const statusVariants = cva(
  'inline-flex shrink-0 items-center gap-1.5 whitespace-nowrap tabular-nums',
  {
    variants: {
      tone: {
        ok: '[&_.status-dot]:text-status-ok',
        warn: '[&_.status-dot]:text-status-warn',
        danger: '[&_.status-dot]:text-status-danger',
        progress: '[&_.status-dot]:text-status-progress',
        neutral: '[&_.status-dot]:text-status-neutral [&_.status-dot]:shadow-none',
      },
      variant: {
        pill: 'rounded-full border border-border bg-muted font-medium text-muted-foreground',
        text: '',
      },
      size: {
        default: '',
        sm: '',
      },
    },
    compoundVariants: [
      { variant: 'pill', size: 'default', class: 'px-2.5 py-0.5 text-xs' },
      { variant: 'pill', size: 'sm', class: 'px-2 py-0.5 text-xs' },
      { variant: 'text', tone: 'progress', class: 'text-status-progress' },
      { variant: 'text', tone: 'warn', class: 'text-status-warn' },
      { variant: 'text', tone: 'danger', class: 'text-status-danger' },
    ],
    defaultVariants: {
      tone: 'neutral',
      variant: 'pill',
      size: 'default',
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
  variant?: StatusVariants['variant'];
  size?: StatusVariants['size'];
}>();
</script>

<template>
  <div :class="cn(statusVariants({ tone, variant, size }), props.class)">
    <span v-if="variant !== 'text'" class="status-dot size-1.5 shrink-0 rounded-full bg-current shadow-glow" />
    <slot />
  </div>
</template>
