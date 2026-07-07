<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import type { VariantProps } from 'class-variance-authority';
import { AlertDialogAction, type AlertDialogActionProps } from 'reka-ui';
import { cn } from '@/lib/utils';
import { buttonVariants } from '@/components/ui/button';

type Variant = VariantProps<typeof buttonVariants>['variant'];

const props = defineProps<
  AlertDialogActionProps & { class?: HTMLAttributes['class']; variant?: Variant }
>();
</script>

<template>
  <AlertDialogAction :class="cn(buttonVariants({ variant }), props.class)">
    <slot />
  </AlertDialogAction>
</template>
