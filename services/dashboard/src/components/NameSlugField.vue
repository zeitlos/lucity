<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { CircleHelp, Pencil } from 'lucide-vue-next';
import { deriveSlug, isValidSlug } from '@/lib/slug';

const props = withDefaults(defineProps<{
  name: string;
  slug: string;
  disabled?: boolean;
  namePlaceholder?: string;
  nameLabel?: string;
  slugDescription?: string;
}>(), {
  namePlaceholder: 'e.g. My Project',
  nameLabel: 'Name',
  slugDescription: 'Used in URLs and infrastructure. Auto-derived from the name.',
});

const emit = defineEmits<{
  (e: 'update:name', value: string): void;
  (e: 'update:slug', value: string): void;
}>();

const nameInputRef = ref<InstanceType<typeof Input> | null>(null);
const slugInputRef = ref<HTMLInputElement | null>(null);
const slugEditing = ref(false);
const slugManuallyEdited = ref(false);

const isValid = computed(() =>
  props.name.trim().length > 0 && isValidSlug(props.slug),
);

watch(() => props.name, (newName) => {
  if (!slugManuallyEdited.value) {
    emit('update:slug', deriveSlug(newName));
  }
  if (newName === '') {
    slugManuallyEdited.value = false;
  }
}, { immediate: true });

function onNameInput(value: string | number) {
  emit('update:name', String(value));
}

function startSlugEdit() {
  if (props.disabled) return;
  slugEditing.value = true;
  nextTick(() => {
    slugInputRef.value?.focus();
  });
}

function onSlugInput(e: Event) {
  const value = (e.target as HTMLInputElement).value;
  slugManuallyEdited.value = true;
  emit('update:slug', value);
}

function confirmSlugEdit() {
  slugEditing.value = false;
}

function cancelSlugEdit() {
  slugEditing.value = false;
  slugManuallyEdited.value = false;
  emit('update:slug', deriveSlug(props.name));
}

function focusName() {
  const el = nameInputRef.value?.$el;
  if (el instanceof HTMLInputElement) {
    el.focus();
  } else {
    (nameInputRef.value as unknown as HTMLElement)?.querySelector?.('input')?.focus();
  }
}

defineExpose({ isValid, focusName });
</script>

<template>
  <div>
    <Label class="mb-2 block">{{ nameLabel }}</Label>
    <Input
      ref="nameInputRef"
      :model-value="name"
      :placeholder="namePlaceholder"
      :disabled="disabled"
      autocomplete="off"
      data-1p-ignore
      @update:model-value="onNameInput"
    />
    <div class="mx-3 flex items-center gap-1.5 rounded-b-md border border-t-0 border-input bg-muted/50 px-2 py-1.5">
      <template v-if="!slugEditing">
        <button
          type="button"
          class="group inline-flex cursor-pointer items-center gap-1 rounded px-1.5 py-0.5 font-mono text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          :class="{ 'pointer-events-none opacity-50': disabled }"
          :disabled="disabled"
          @click="startSlugEdit"
        >
          <span class="truncate">{{ slug || 'auto-generated-slug' }}</span>
          <Pencil
            :size="12"
            class="shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
          />
        </button>
      </template>
      <template v-else>
        <input
          ref="slugInputRef"
          :value="slug"
          class="rounded bg-muted px-1.5 py-0.5 font-mono text-xs text-foreground outline-none"
          autocomplete="off"
          data-1p-ignore
          @input="onSlugInput"
          @blur="confirmSlugEdit"
          @keydown.enter.stop.prevent="confirmSlugEdit"
          @keydown.escape="cancelSlugEdit"
        />
      </template>
      <div class="flex-1" />
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger as-child>
            <CircleHelp :size="12" class="shrink-0 cursor-help text-muted-foreground" />
          </TooltipTrigger>
          <TooltipContent
            side="top"
            class="max-w-60 text-xs"
          >
            {{ slugDescription }}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  </div>
</template>
