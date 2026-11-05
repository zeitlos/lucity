<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import { computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { useEnvironment } from '@/composables/useEnvironment';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import { useVariableRows } from '@/composables/useVariableRows';
import FormFooter from '@/components/FormFooter.vue';
import VariablesEditor from '@/components/VariablesEditor.vue';
import { failureToast } from '@/components/ui/sonner';
import type { VariableReference, VariableSourceTypename } from '@/lib/variables';

const ServiceVariablesDocument = graphql(`
  query ServiceVariables($service: ServiceID!) {
    serviceVariables(service: $service) {
      key
      value
      ref
    }
  }
`);

const AvailableVariablesDocument = graphql(`
  query AvailableVariables($environment: EnvironmentID!) {
    availableVariables(environment: $environment) {
      id
      key
      source {
        __typename
        ... on DatabaseSource { databaseId: id name }
        ... on KeyValueStoreSource { keyValueStoreId: id name }
        ... on BucketSource { bucketId: id name }
        ... on SharedSource { name }
      }
    }
  }
`);

const UpdateServiceVariablesDocument = graphql(`
  mutation UpdateServiceVariables($service: ServiceID!, $set: [ServiceVariableInput!]!, $unset: [String!]!) {
    updateServiceVariables(service: $service, set: $set, unset: $unset)
  }
`);

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const { activeEnvironment } = useEnvironment();
const environmentId = computed(() => activeEnvironment.value?.id ?? '');

const { result, loading, error, refetch } = useQuery(
  ServiceVariablesDocument,
  () => ({ service: props.serviceId }),
  () => ({ enabled: !!props.serviceId }),
);

const { result: availableResult } = useQuery(
  AvailableVariablesDocument,
  () => ({ environment: environmentId.value }),
  () => ({ enabled: !!environmentId.value }),
);

const references = computed<VariableReference[]>(() =>
  (availableResult.value?.availableVariables ?? []).map(v => ({
    id: v.id,
    key: v.key,
    typename: v.source.__typename as VariableSourceTypename,
    sourceName: v.source.name,
  })),
);

const { rows, changes, dirty, duplicate, reset, commit } = useVariableRows(() =>
  result.value?.serviceVariables.map(v => ({ key: v.key, value: v.value ?? '', ref: v.ref ?? null })),
);

useUnsavedChanges(dirty);

const { mutate: updateVarsMutate, loading: saving } = useMutation(UpdateServiceVariablesDocument);

async function handleSave() {
  try {
    await updateVarsMutate({
      service: props.serviceId,
      set: changes.value.set.map(change => ({
        key: change.key,
        value: change.ref ? undefined : change.value,
        ref: change.ref ?? undefined,
      })),
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
  <QueryError
    v-if="error && !result"
    title="Couldn't load variables"
    :error="error"
    @retry="refetch()"
  />

  <div v-else class="space-y-4">
    <VariablesEditor
      v-model:rows="rows"
      :references="references"
      :loading="loading"
      :disabled="saving"
      empty-title="No variables yet"
      empty-description="Add one, or paste KEY=value lines into a key field."
    />

    <FormFooter
      class="sticky bottom-0 -mx-4 -mb-4 bg-card"
      :dirty="dirty"
      :saving="saving"
      :disabled="!!duplicate"
      @save="handleSave"
      @discard="reset"
    >
      <template #hint>
        <template v-if="duplicate">Two variables are named "{{ duplicate }}".</template>
        <template v-else>
          Available at build time and runtime. Paste <code class="rounded bg-muted px-1 py-0.5 font-mono">KEY=value</code> lines into a key field to add several.
        </template>
      </template>
    </FormFooter>
  </div>
</template>
