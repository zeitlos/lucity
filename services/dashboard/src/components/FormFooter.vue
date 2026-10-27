<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

const props = withDefaults(defineProps<{
  dirty: boolean;
  saving?: boolean;
  disabled?: boolean;
  hint?: string;
  class?: HTMLAttributes['class'];
}>(), {
  saving: false,
  disabled: false,
  hint: undefined,
  class: undefined,
});

const emit = defineEmits<{
  (e: 'save'): void;
  (e: 'discard'): void;
}>();
</script>

<template>
  <div :class="cn('flex items-center gap-2 border-t px-4 py-3', props.class)">
    <p class="min-w-0 flex-1 text-xs text-muted-foreground">
      <slot name="hint">{{ hint }}</slot>
    </p>
    <Button
      v-if="dirty && !saving"
      variant="ghost"
      class="shrink-0"
      @click="emit('discard')"
    >
      Discard
    </Button>
    <Button
      class="shrink-0"
      :pending="saving"
      :disabled="!dirty || disabled"
      @click="emit('save')"
    >
      {{ saving ? 'Saving…' : 'Save' }}
    </Button>
  </div>
</template>
