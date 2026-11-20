<script setup lang="ts">
import { ref, nextTick, useId } from 'vue';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useIdField } from '@/composables/useIdField';
import { ID_MAX_LENGTH, sanitizeId } from '@/lib/id';

const props = withDefaults(defineProps<{
  name: string;
  id: string;
  fallback: string;
  taken?: ReadonlySet<string>;
  scope?: string;
  reserved?: boolean;
  disabled?: boolean;
  namePlaceholder?: string;
  nameLabel?: string;
}>(), {
  taken: () => new Set<string>(),
  scope: undefined,
  reserved: false,
  disabled: false,
  namePlaceholder: 'e.g. My Project',
  nameLabel: 'Name',
});

const emit = defineEmits<{
  (e: 'update:name', value: string): void;
  (e: 'update:id', value: string): void;
}>();

const nameInputId = useId();
const idInputId = useId();
const nameInputRef = ref<InstanceType<typeof Input> | null>(null);
const idInputRef = ref<HTMLInputElement | null>(null);
const editing = ref(false);

const field = useIdField(props, id => emit('update:id', id));
const { base, reason, numbered, isValid, error } = field;

function openEdit() {
  editing.value = true;
  nextTick(() => {
    idInputRef.value?.focus();
    idInputRef.value?.select();
  });
}

function closeEdit() {
  editing.value = false;
  field.reset();
}

function onIdInput(event: Event) {
  const input = event.target as HTMLInputElement;
  const id = sanitizeId(input.value);
  input.value = id;
  field.input(id);
}

function focusName() {
  const el = nameInputRef.value?.$el;
  if (el instanceof HTMLInputElement) el.focus();
}

defineExpose({ isValid, focusName });
</script>

<template>
  <div>
    <Label :for="nameInputId" class="mb-2 block">{{ nameLabel }}</Label>
    <Input
      :id="nameInputId"
      ref="nameInputRef"
      :model-value="name"
      :placeholder="namePlaceholder"
      :disabled="disabled"
      autocomplete="off"
      data-1p-ignore
      @update:model-value="emit('update:name', String($event))"
    />

    <div v-if="editing" class="mt-4">
      <div class="mb-2 flex items-baseline justify-between">
        <Label :for="idInputId">ID</Label>
        <button
          type="button"
          class="text-xs font-medium text-foreground underline decoration-muted-foreground underline-offset-4 hover:decoration-foreground disabled:opacity-50"
          :disabled="disabled"
          @click="closeEdit"
        >
          Reset
        </button>
      </div>
      <input
        :id="idInputId"
        ref="idInputRef"
        :value="id"
        :maxlength="ID_MAX_LENGTH"
        :disabled="disabled"
        :aria-invalid="!!error"
        class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive aria-invalid:focus-visible:ring-destructive"
        autocomplete="off"
        data-1p-ignore
        spellcheck="false"
        @input="onIdInput"
      >
      <p v-if="error" class="mt-1.5 text-xs text-destructive">{{ error }}</p>
      <p v-else-if="numbered" class="mt-1.5 text-xs text-muted-foreground">"{{ base }}" is {{ reason }}, so a number was added.</p>
      <p v-else class="mt-1.5 text-xs text-muted-foreground">It can't be changed later.</p>
    </div>

    <p v-else class="mt-2 text-xs text-muted-foreground">
      ID <span class="font-mono text-foreground">{{ id }}</span><template v-if="numbered">, because "{{ base }}" is {{ reason }}</template>. It can't be changed later.
      <button
        type="button"
        class="font-medium text-foreground underline decoration-muted-foreground underline-offset-4 hover:decoration-foreground disabled:opacity-50"
        :disabled="disabled"
        @click="openEdit"
      >
        Edit
      </button>
    </p>
  </div>
</template>
