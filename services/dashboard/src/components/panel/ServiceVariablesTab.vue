<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { useEnvironment } from '@/composables/useEnvironment';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import FormFooter from '@/components/FormFooter.vue';
import VariablesEditor from '@/components/VariablesEditor.vue';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import type { VariableReference, VariableRow, VariableSourceTypename } from '@/lib/variables';

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

const SetServiceVariablesDocument = graphql(`
  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {
    setServiceVariables(service: $service, variables: $variables)
  }
`);

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const { activeEnvironment } = useEnvironment();
const environmentId = computed(() => activeEnvironment.value?.id ?? '');

const rows = ref<VariableRow[]>([]);
const hasChanges = ref(false);

const { result, loading, refetch } = useQuery(
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

function resetRows() {
  const vars = result.value?.serviceVariables;
  if (!vars) return;
  rows.value = vars.map(v => ({ key: v.key, value: v.value ?? '', ref: v.ref ?? null }));
  hasChanges.value = false;
}

watch(() => result.value?.serviceVariables, () => {
  if (!hasChanges.value) resetRows();
}, { immediate: true });

useUnsavedChanges(hasChanges);

const { mutate: setVarsMutate, loading: saving } = useMutation(SetServiceVariablesDocument);

async function handleSave() {
  try {
    const variables = rows.value
      .filter(r => r.key.trim())
      .map(r => ({
        key: r.key.trim(),
        value: r.ref ? undefined : r.value,
        ref: r.ref ?? undefined,
      }));

    const res = await setVarsMutate({ service: props.serviceId, variables });

    if (res?.errors?.length) {
      errorToast('Failed to save variables', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    toast.success('Service variables saved');
    hasChanges.value = false;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to save variables', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-4">
    <VariablesEditor
      v-model:rows="rows"
      :references="references"
      :loading="loading"
      @change="hasChanges = true"
    >
      <template #empty>
        <p class="text-sm text-muted-foreground">No variables configured for this service.</p>
      </template>
    </VariablesEditor>

    <FormFooter
      class="sticky bottom-0 -mx-4 -mb-4 bg-card"
      :dirty="hasChanges"
      :saving="saving"
      @save="handleSave"
      @discard="resetRows"
    >
      <template #hint>
        Available at build time and runtime. Paste <code class="rounded bg-muted px-1 py-0.5 font-mono">KEY=value</code> lines into a key field to add several.
      </template>
    </FormFooter>
  </div>
</template>
