<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const alertVariants = cva(
  'relative w-full rounded-lg border p-3 text-sm [&>svg~*]:pl-7 [&>svg]:absolute [&>svg]:left-3 [&>svg]:top-3.5 [&>svg]:size-4',
  {
    variants: {
      variant: {
        default: 'bg-card text-foreground [&>svg]:text-muted-foreground',
        info: 'border-status-progress/30 bg-status-progress/5 [&>svg]:text-status-progress',
        ok: 'border-status-ok/30 bg-status-ok/5 [&>svg]:text-status-ok',
        warn: 'border-status-warn/30 bg-status-warn/5 [&>svg]:text-status-warn',
        danger: 'border-status-danger/30 bg-status-danger/5 [&>svg]:text-status-danger',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);

type AlertVariants = VariantProps<typeof alertVariants>;
</script>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  variant?: AlertVariants['variant'];
}>();
</script>

<template>
  <div :class="cn(alertVariants({ variant }), props.class)" role="alert">
    <slot />
  </div>
</template>
