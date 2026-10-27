<script setup lang="ts">
import { ref, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import FormFooter from '@/components/FormFooter.vue';
import VariablesEditor from '@/components/VariablesEditor.vue';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import type { VariableRow } from '@/lib/variables';

const SharedVariablesDocument = graphql(`
  query SharedVariables($environment: EnvironmentID!) {
    sharedVariables(environment: $environment) {
      key
      value
    }
  }
`);

const SetSharedVariablesDocument = graphql(`
  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {
    setSharedVariables(environment: $environment, variables: $variables)
  }
`);

const props = defineProps<{
  environmentId: string;
  environmentName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, refetch } = useQuery(SharedVariablesDocument, () => ({
  environment: props.environmentId,
}));

const rows = ref<VariableRow[]>([]);
const hasChanges = ref(false);

function resetRows() {
  const vars = result.value?.sharedVariables;
  if (!vars) return;
  rows.value = vars.map(v => ({ key: v.key, value: v.value }));
  hasChanges.value = false;
}

watch(() => result.value?.sharedVariables, () => {
  if (!hasChanges.value) resetRows();
}, { immediate: true });

useUnsavedChanges(hasChanges);

const { mutate: setVarsMutate, loading: saving } = useMutation(SetSharedVariablesDocument);

async function handleSave() {
  try {
    const res = await setVarsMutate({
      environment: props.environmentId,
      variables: rows.value
        .filter(r => r.key.trim())
        .map(r => ({ key: r.key.trim(), value: r.value })),
    });

    if (res?.errors?.length) {
      errorToast('Failed to save variables', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Shared variables saved');
    hasChanges.value = false;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to save variables', { description: errorMessage(e) });
  }
}
</script>

<template>
  <ContextDrawer
    title="Shared variables"
    @close="emit('close')"
  >
    <VariablesEditor
      v-model:rows="rows"
      :loading="loading"
      :disabled="saving"
      @change="hasChanges = true"
    >
      <template #empty>
        <p class="text-sm text-muted-foreground">No shared variables yet.</p>
        <p class="mt-1 text-xs text-muted-foreground">Services in {{ environmentName }} can reference them.</p>
      </template>
    </VariablesEditor>

    <template #footer>
      <FormFooter
        class="border-t-0 p-0"
        :dirty="hasChanges"
        :saving="saving"
        @save="handleSave"
        @discard="resetRows"
      >
        <template #hint>
          Paste <code class="rounded bg-muted px-1 py-0.5 font-mono">KEY=value</code> lines into a key field to add several.
        </template>
      </FormFooter>
    </template>
  </ContextDrawer>
</template>
