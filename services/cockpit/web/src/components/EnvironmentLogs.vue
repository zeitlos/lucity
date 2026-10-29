<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable';

import { graphql } from '@/gql';

const props = defineProps<{
  environmentId: string;
}>();

const EnvironmentLogsDocument = graphql(`
  query EnvironmentLogs($id: ID!) {
    environmentLogs(id: $id)
  }
`);

const { result, loading, error } = useQuery(EnvironmentLogsDocument, () => ({ id: props.environmentId }), { fetchPolicy: 'network-only' });
</script>

<template>
  <div class="rounded-sm bg-foreground px-3 py-2 font-mono text-background">
    <p v-if="loading">Loading logs…</p>
    <p v-else-if="error" class="text-status-danger-background">{{ error.message }}</p>
    <p v-else-if="!result?.environmentLogs.length">No running pods.</p>
    <pre v-else class="max-h-72 overflow-auto whitespace-pre-wrap break-all">{{ result.environmentLogs.join('\n') }}</pre>
  </div>
</template>
