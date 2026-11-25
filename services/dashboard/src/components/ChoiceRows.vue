<script setup lang="ts" generic="O extends { value: string; label: string; hint?: string }">
import type { HTMLAttributes } from 'vue';
import { nextTick } from 'vue';
import { RadioGroupIndicator, RadioGroupItem } from 'reka-ui';
import { Check } from '@lucide/vue';
import { cn } from '@/lib/utils';
import { RadioGroup } from '@/components/ui/radio-group';

const props = withDefaults(defineProps<{
  options: O[];
  columns?: 1 | 2;
  class?: HTMLAttributes['class'];
}>(), {
  columns: 2,
  class: undefined,
});

const model = defineModel<O['value']>({ required: true });

const emit = defineEmits<{
  (e: 'pointer'): void;
}>();

defineSlots<{
  visual(props: { option: O }): unknown;
}>();

function select(value: unknown) {
  if (typeof value === 'string' && value) model.value = value as O['value'];
}

async function onClick(event: MouseEvent, value: O['value']) {
  if (!event.detail) return;
  model.value = value;
  await nextTick();
  emit('pointer');
}
</script>

<template>
  <RadioGroup :model-value="model" :class="cn('gap-1', columns === 2 && 'grid-cols-2', props.class)" @update:model-value="select">
    <RadioGroupItem
      v-for="option in options"
      :key="option.value"
      :value="option.value"
      class="flex min-w-0 items-center gap-2.5 rounded-md border border-transparent p-1 text-left text-sm text-foreground transition-colors hover:bg-accent/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none data-[state=checked]:border-border data-[state=checked]:bg-accent data-[state=checked]:font-medium"
      @click="onClick($event, option.value)"
    >
      <slot name="visual" :option="option" />
      <span class="flex min-w-0 flex-1 items-baseline gap-2">
        <span class="truncate">{{ option.label }}</span>
        <span v-if="option.hint" class="truncate text-xs font-normal text-muted-foreground">{{ option.hint }}</span>
      </span>
      <RadioGroupIndicator class="mr-1 shrink-0">
        <Check :size="16" />
      </RadioGroupIndicator>
    </RadioGroupItem>
  </RadioGroup>
</template>
