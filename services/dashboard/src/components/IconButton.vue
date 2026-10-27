<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { Button, type ButtonVariants } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';

defineOptions({ inheritAttrs: false });

const props = withDefaults(defineProps<{
  label: string;
  shortcut?: string;
  variant?: ButtonVariants['variant'];
  size?: ButtonVariants['size'];
  side?: 'top' | 'right' | 'bottom' | 'left';
  class?: HTMLAttributes['class'];
}>(), {
  shortcut: undefined,
  variant: 'ghost',
  size: 'icon-sm',
  side: 'top',
  class: undefined,
});
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Button
        v-bind="$attrs"
        :variant="variant"
        :size="size"
        :class="props.class"
        :aria-label="label"
      >
        <slot />
      </Button>
    </TooltipTrigger>
    <TooltipContent :side="side" :class="darkTooltipClass">
      {{ label }}
      <kbd
        v-if="shortcut"
        class="ml-1.5 rounded border border-background/30 px-1 font-mono [.dark_&]:border-border"
      >{{ shortcut }}</kbd>
    </TooltipContent>
  </Tooltip>
</template>
