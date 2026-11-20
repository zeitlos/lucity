<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import { useVariableRows } from '@/composables/useVariableRows';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import FormFooter from '@/components/FormFooter.vue';
import VariablesEditor from '@/components/VariablesEditor.vue';
import { failureToast } from '@/components/ui/sonner';

const SharedVariablesDocument = graphql(`
  query SharedVariables($environment: EnvironmentID!) {
    sharedVariables(environment: $environment) {
      key
      value
    }
  }
`);

const UpdateSharedVariablesDocument = graphql(`
  mutation UpdateSharedVariables($environment: EnvironmentID!, $set: [VariableInput!]!, $unset: [String!]!) {
    updateSharedVariables(environment: $environment, set: $set, unset: $unset)
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

const { rows, changes, dirty, duplicate, reset, commit } = useVariableRows(() =>
  result.value?.sharedVariables.map(v => ({ key: v.key, value: v.value })),
);

useUnsavedChanges(dirty);

const { mutate: updateVarsMutate, loading: saving } = useMutation(UpdateSharedVariablesDocument);

async function handleSave() {
  try {
    await updateVarsMutate({
      environment: props.environmentId,
      set: changes.value.set.map(change => ({ key: change.key, value: change.value })),
      unset: changes.value.unset,
    });

    commit();
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
    />

    <template v-if="result" #footer>
      <FormFooter
        class="border-t-0 p-0"
        :dirty="dirty"
        :saving="saving"
        :disabled="!!duplicate"
        @save="handleSave"
        @discard="reset"
      >
        <template #hint>
          <template v-if="duplicate">Two variables are named "{{ duplicate }}".</template>
          <template v-else>
            Paste <code class="rounded bg-muted px-1 py-0.5 font-mono">KEY=value</code> lines into a key field to add several.
          </template>
        </template>
      </FormFooter>
    </template>
  </ContextDrawer>
</template>
