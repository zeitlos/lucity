<script setup lang="ts">
import { ref, reactive, computed, watch, nextTick, onMounted } from 'vue';
import type { Component } from 'vue';
import { useApolloClient } from '@vue/apollo-composable';
import { onKeyStroke } from '@vueuse/core';
import { ArrowLeft, Loader2, Check, ChevronRight, CornerDownLeft } from '@lucide/vue';
import type { CommandFlowConfig } from '@/lib/commandFlow';
import { errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{ flow: CommandFlowConfig }>();
const emit = defineEmits<{
  (e: 'back'): void;
  (e: 'created'): void;
}>();

const { resolveClient } = useApolloClient();

const stepIndex = ref(0);
const values = reactive<Record<string, string>>({});
const focusedIndex = ref(0);
const submitting = ref(false);
const inputRef = ref<HTMLInputElement>();

props.flow.steps.forEach((step) => { values[step.id] = step.initial ?? ''; });

const currentStep = computed(() => props.flow.steps[stepIndex.value]!);
const isLastStep = computed(() => stepIndex.value === props.flow.steps.length - 1);

const inputModel = computed<string>({
  get: () => values[currentStep.value.id] ?? '',
  set: (value) => { values[currentStep.value.id] = value; },
});

const stepValid = computed(() => {
  const step = currentStep.value;
  const value = values[step.id] ?? '';
  if (step.validate) return step.validate(value);
  if (step.required === false) return true;
  return value.trim().length > 0;
});

const showInvalidHint = computed(() =>
  !stepValid.value && (values[currentStep.value.id] ?? '').length > 0 && !!currentStep.value.invalidHint,
);

type Row = { id: string; label: string; icon: Component; actionLabel: string; disabled?: boolean; action: () => void };

const rows = computed<Row[]>(() => {
  if (!isLastStep.value) {
    return [{ id: 'next', label: 'Continue', icon: ChevronRight, actionLabel: 'next', disabled: !stepValid.value, action: goNext }];
  }
  return [{
    id: 'submit',
    label: submitting.value ? 'Creating...' : (props.flow.submitLabel ?? 'Create'),
    icon: Check,
    actionLabel: 'create',
    disabled: !stepValid.value || submitting.value,
    action: submit,
  }];
});

function goNext() {
  if (stepValid.value) stepIndex.value++;
}

function back() {
  if (stepIndex.value > 0) stepIndex.value--;
  else emit('back');
}

async function submit() {
  if (!stepValid.value || submitting.value) return;
  submitting.value = true;
  try {
    const res = await resolveClient().mutate({
      mutation: props.flow.mutation,
      variables: props.flow.variables({ ...values }),
      errorPolicy: 'all',
    });
    if (res.errors?.length) {
      errorToast(`Failed to create ${props.flow.title.toLowerCase()}`, {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }
    emit('created');
  } catch (e: unknown) {
    errorToast(`Failed to create ${props.flow.title.toLowerCase()}`, { description: errorMessage(e) });
  } finally {
    submitting.value = false;
  }
}

onMounted(() => { focusedIndex.value = 0; nextTick(() => inputRef.value?.focus()); });
watch(stepIndex, () => { focusedIndex.value = 0; nextTick(() => inputRef.value?.focus()); });

onKeyStroke('ArrowDown', (e) => {
  if (rows.value.length === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value + 1) % rows.value.length;
});
onKeyStroke('ArrowUp', (e) => {
  if (rows.value.length === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value - 1 + rows.value.length) % rows.value.length;
});
onKeyStroke('Enter', (e) => {
  const row = rows.value[focusedIndex.value];
  if (!row || row.disabled) return;
  e.preventDefault();
  row.action();
});
</script>

<template>
  <div>
    <div class="flex items-center border-b px-3">
      <button
        class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
        @click="back"
      >
        <ArrowLeft :size="16" />
      </button>
      <img v-if="flow.iconSrc" :src="flow.iconSrc" :width="18" :height="18" class="shrink-0" alt="" />
      <component :is="flow.icon" v-else-if="flow.icon" :size="18" class="shrink-0 text-muted-foreground" />
      <input
        ref="inputRef"
        v-model="inputModel"
        :placeholder="currentStep.placeholder"
        class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
        autocomplete="off"
        data-1p-ignore
        spellcheck="false"
      />
      <Loader2 v-if="submitting" :size="14" class="shrink-0 animate-spin text-muted-foreground" />
    </div>

    <div class="max-h-[50vh] overflow-y-auto">
      <p v-if="showInvalidHint" class="px-3 py-1.5 text-xs text-destructive">{{ currentStep.invalidHint }}</p>

      <button
        v-for="(row, index) in rows"
        :key="row.id"
        :data-focused="focusedIndex === index"
        :disabled="row.disabled"
        class="group flex h-12 w-full items-center gap-2 px-3 text-sm text-popover-foreground transition-colors disabled:opacity-50"
        :class="[
          focusedIndex === index ? 'bg-accent' : 'hover:bg-accent/40',
          index === rows.length - 1 ? 'rounded-b-xl' : 'rounded-none',
        ]"
        @click="row.action()"
      >
        <component :is="row.icon" :size="16" class="shrink-0 text-muted-foreground" />
        <span class="min-w-0 flex-1 truncate text-left">{{ row.label }}</span>
        <span
          v-if="focusedIndex === index && !row.disabled"
          class="flex shrink-0 items-center gap-1 text-xs font-medium text-green-500"
        >
          {{ row.actionLabel }}
          <CornerDownLeft :size="14" />
        </span>
        <span
          v-else-if="!row.disabled"
          class="shrink-0 text-xs text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
        >
          {{ row.actionLabel }}
        </span>
      </button>
    </div>
  </div>
</template>
