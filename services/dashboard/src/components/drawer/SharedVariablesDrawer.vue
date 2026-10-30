<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import { ref, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import FormFooter from '@/components/FormFooter.vue';
import VariablesEditor from '@/components/VariablesEditor.vue';
import { failureToast } from '@/components/ui/sonner';
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
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, error, refetch } = useQuery(SharedVariablesDocument, () => ({
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
    await setVarsMutate({
      environment: props.environmentId,
      variables: rows.value
        .filter(r => r.key.trim())
        .map(r => ({ key: r.key.trim(), value: r.value })),
    });

    hasChanges.value = false;
    refetch();
  } catch (e: unknown) {
    failureToast('Failed to save variables', e);
  }
}
</script>

<template>
  <ContextDrawer
    title="Shared variables"
    @close="emit('close')"
  >
    <QueryError
      v-if="error && !result"
      title="Couldn't load shared variables"
      :error="error"
      @retry="refetch()"
    />

    <VariablesEditor
      v-else
      v-model:rows="rows"
      :loading="loading"
      :disabled="saving"
      empty-title="No shared variables yet"
      empty-description="Services in this environment can reference them."
      @change="hasChanges = true"
    />

    <template v-if="result" #footer>
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
