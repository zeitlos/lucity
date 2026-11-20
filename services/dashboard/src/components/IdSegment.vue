<script setup lang="ts">
import { computed, useId } from 'vue';
import { useIdField } from '@/composables/useIdField';
import { ID_MAX_LENGTH, sanitizeId } from '@/lib/id';

const props = withDefaults(defineProps<{
  name: string;
  id: string;
  fallback: string;
  taken?: ReadonlySet<string>;
  scope?: string;
  disabled?: boolean;
}>(), {
  taken: () => new Set<string>(),
  scope: undefined,
  disabled: false,
});

const emit = defineEmits<{
  (e: 'update:id', value: string): void;
}>();

const inputId = useId();
const field = useIdField(props, id => emit('update:id', id));
const { edited, derived } = field;

const message = computed(() => {
  if (field.error.value) return { text: field.error.value, error: true };
  if (field.numbered.value) {
    return { text: `"${field.base.value}" is ${field.reason.value}, so a number was added.`, error: false };
  }
  return null;
});

function onInput(event: Event) {
  const input = event.target as HTMLInputElement;
  const id = sanitizeId(input.value);
  input.value = id;
  if (id) field.input(id);
  else field.reset();
}

defineExpose({ message });
</script>

<template>
  <label
    :for="inputId"
    class="group flex h-8 shrink-0 items-center gap-1.5 border-l px-2 transition-colors focus-within:rounded-md focus-within:border-transparent focus-within:bg-accent"
  >
    <span class="text-xs text-muted-foreground">ID</span>
    <input
      :id="inputId"
      :value="edited ? id : ''"
      :placeholder="derived"
      :maxlength="ID_MAX_LENGTH"
      :disabled="disabled"
      class="w-36 bg-transparent font-mono text-sm outline-none placeholder:text-muted-foreground disabled:opacity-50"
      :class="{ 'text-destructive': message?.error }"
      autocomplete="off"
      data-1p-ignore
      spellcheck="false"
      @input="onInput"
    >
    <kbd class="rounded border bg-card px-1.5 font-sans text-xs leading-4.5 text-muted-foreground group-focus-within:hidden max-md:hidden">tab</kbd>
  </label>
</template>
