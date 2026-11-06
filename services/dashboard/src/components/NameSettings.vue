<script setup lang="ts">
import { ref, computed, watch, useId } from 'vue';
import FormFooter from '@/components/FormFooter.vue';
import CopyButton from '@/components/CopyButton.vue';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';

const props = withDefaults(defineProps<{
  name: string;
  id: string;
  saving?: boolean;
  disabled?: boolean;
}>(), {
  saving: false,
  disabled: false,
});

const emit = defineEmits<{
  (e: 'save', name: string): void;
}>();

const inputId = useId();
const draft = ref(props.name);
const changed = computed(() => draft.value.trim() !== props.name);

useUnsavedChanges(changed);

watch(() => props.name, (name) => {
  draft.value = name;
});

function save() {
  const name = draft.value.trim();
  if (!name || name === props.name) return;
  emit('save', name);
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border">
    <fieldset :disabled="saving" class="min-w-0">
      <div class="space-y-3.5 p-3">
        <form class="space-y-1.5" @submit.prevent="save">
          <Label :for="inputId" class="text-xs text-muted-foreground">Name</Label>
          <Input
            :id="inputId"
            v-model="draft"
            class="h-9 min-w-0"
            :disabled="disabled || saving"
          />
          <p class="flex min-w-0 items-center gap-1 text-xs text-muted-foreground">
            ID <span class="truncate font-mono text-foreground">{{ id }}</span>
            <CopyButton :value="id" label="Copy ID" />
          </p>
        </form>
      </div>
    </fieldset>
    <FormFooter
      v-if="!disabled"
      :dirty="changed"
      :saving="saving"
      :disabled="!draft.trim()"
      @save="save"
      @discard="draft = name"
    />
  </div>
</template>
