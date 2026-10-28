<script setup lang="ts" generic="T extends string">
import type { Component } from 'vue';
import { ToggleGroupItem, ToggleGroupRoot } from 'reka-ui';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';

defineProps<{
  label: string;
  options: { value: T; label: string; icon?: Component }[];
}>();

const model = defineModel<T>({ required: true });

const itemClass = 'flex h-6 min-w-7 items-center justify-center rounded-sm px-2 font-medium text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 aria-pressed:bg-card aria-pressed:text-foreground aria-pressed:shadow-xs [&_svg]:size-3.5';

function select(value: unknown) {
  if (typeof value === 'string' && value) model.value = value as T;
}
</script>

<template>
  <ToggleGroupRoot
    type="single"
    :model-value="model"
    :aria-label="label"
    class="inline-flex shrink-0 items-center gap-0.5 rounded-md bg-muted p-0.5 text-xs"
    @update:model-value="select"
  >
    <template v-for="option in options" :key="option.value">
      <Tooltip v-if="option.icon">
        <TooltipTrigger as-child>
          <ToggleGroupItem :value="option.value" :aria-label="option.label" :class="itemClass">
            <component :is="option.icon" />
          </ToggleGroupItem>
        </TooltipTrigger>
        <TooltipContent :class="darkTooltipClass">
          {{ option.label }}
        </TooltipContent>
      </Tooltip>
      <ToggleGroupItem v-else :value="option.value" :class="itemClass">
        {{ option.label }}
      </ToggleGroupItem>
    </template>
  </ToggleGroupRoot>
</template>
