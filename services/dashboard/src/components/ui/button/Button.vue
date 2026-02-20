<script lang="ts">
import { type VariantProps, cva } from 'class-variance-authority';

export const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-none text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground shadow-[var(--shadow-button)] hover:brightness-[0.92] hover:-translate-y-px hover:shadow-[0_6px_28px_-4px_oklch(0.75_0.18_160/0.35),0_12px_50px_-8px_oklch(0.75_0.18_160/0.18)] active:translate-y-0 active:shadow-[inset_0_2px_4px_oklch(0/0.15)]',
        destructive: 'bg-destructive text-destructive-foreground shadow-[var(--shadow-destructive-button)] hover:brightness-[0.92] hover:-translate-y-px active:translate-y-0',
        outline: 'border border-input bg-background hover:border-primary hover:text-primary',
        secondary: 'bg-secondary text-secondary-foreground border border-border shadow-[0_2px_12px_-2px_oklch(0.50_0.02_55/0.06)] hover:border-muted-foreground',
        ghost: 'hover:bg-muted hover:text-foreground',
        link: 'text-primary underline-offset-4 hover:underline',
        accent: 'bg-accent-pop text-accent-pop-foreground shadow-[var(--shadow-accent-button)] hover:brightness-[0.92] hover:-translate-y-px active:translate-y-0',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm: 'h-9 px-3',
        lg: 'h-11 px-8',
        icon: 'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);

type ButtonVariants = VariantProps<typeof buttonVariants>;
</script>

<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { Primitive, type PrimitiveProps } from 'reka-ui';
import { cn } from '@/lib/utils';

interface Props extends PrimitiveProps {
  variant?: ButtonVariants['variant'];
  size?: ButtonVariants['size'];
  class?: HTMLAttributes['class'];
}

const props = withDefaults(defineProps<Props>(), {
  as: 'button',
});
</script>

<template>
  <Primitive :as="as" :as-child="asChild" :class="cn(buttonVariants({ variant, size }), props.class)">
    <slot />
  </Primitive>
</template>
