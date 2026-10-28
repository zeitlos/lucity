<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ScrollText } from '@lucide/vue';
import { graphql } from '@/gql';
import { AgentAuditKind } from '@/gql/graphql';
import { auditResultStatus } from '@/lib/status';
import { formatCompact, formatDateTime, formatTime } from '@/lib/format';
import BrandTile from '@/components/BrandTile.vue';
import EmptyState from '@/components/EmptyState.vue';
import QueryError from '@/components/QueryError.vue';
import SegmentedControl from '@/components/SegmentedControl.vue';
import { Skeleton } from '@/components/ui/skeleton';
import { Status } from '@/components/ui/status';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';

const AgentAuditDocument = graphql(`
  query AgentAudit($agent: AgentID!, $kinds: [AgentAuditKind!]) {
    agentAudit(agent: $agent, kinds: $kinds) {
      time
      kind
      modelProvider
      connector
      target
      inputTokens
      outputTokens
      result
    }
  }
`);

const props = defineProps<{
  agentId: string;
}>();

type Filter = 'all' | 'models' | 'tools';

const filter = ref<Filter>('all');
const filters: { value: Filter; label: string }[] = [
  { value: 'all', label: 'All' },
  { value: 'models', label: 'Models' },
  { value: 'tools', label: 'Tools' },
];

const kinds = computed(() => {
  if (filter.value === 'models') return [AgentAuditKind.Model];
  if (filter.value === 'tools') return [AgentAuditKind.Tool];
  return null;
});

const { result, loading, error, refetch } = useQuery(
  AgentAuditDocument,
  () => ({ agent: props.agentId, kinds: kinds.value }),
  { pollInterval: 15000 },
);

const entries = computed(() => result.value?.agentAudit ?? []);

function when(time: string) {
  return new Date(time).toDateString() === new Date().toDateString() ? formatTime(time) : formatDateTime(time);
}

function tokens(entry: { inputTokens?: number | null; outputTokens?: number | null }) {
  if (entry.inputTokens == null && entry.outputTokens == null) return '';
  return `${formatCompact(entry.inputTokens ?? 0)} in · ${formatCompact(entry.outputTokens ?? 0)} out`;
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between gap-4">
      <SegmentedControl v-model="filter" label="Calls" :options="filters" />
      <span class="text-xs text-muted-foreground">Last 24 hours</span>
    </div>

    <div v-if="loading && !result" class="space-y-2">
      <Skeleton v-for="i in 6" :key="i" class="h-9 w-full" />
    </div>
    <QueryError
      v-else-if="error && !result"
      title="Couldn't load the audit log"
      :error="error"
      @retry="refetch()"
    />
    <EmptyState
      v-else-if="entries.length === 0"
      size="compact"
      :icon="ScrollText"
      title="No calls yet"
      description="Every call to a model or a connector shows up here."
    />
    <div v-else class="-mx-4 border-y">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="pl-4">Time</TableHead>
            <TableHead>Kind</TableHead>
            <TableHead>Source</TableHead>
            <TableHead>Call</TableHead>
            <TableHead class="pr-4">Result</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(entry, index) in entries" :key="`${entry.time}-${index}`">
            <TableCell class="pl-4 text-muted-foreground tabular-nums whitespace-nowrap">{{ when(entry.time) }}</TableCell>
            <TableCell>{{ entry.kind === AgentAuditKind.Model ? 'Model' : 'Tool' }}</TableCell>
            <TableCell>
              <span class="flex items-center gap-2">
                <BrandTile :name="entry.modelProvider ?? entry.connector ?? '?'" size="sm" />
                <span class="truncate">{{ entry.modelProvider ?? entry.connector }}</span>
              </span>
            </TableCell>
            <TableCell class="max-w-64">
              <span class="block truncate">
                <span class="font-mono">{{ entry.target }}</span>
                <span v-if="tokens(entry)" class="text-muted-foreground"> · {{ tokens(entry) }}</span>
              </span>
            </TableCell>
            <TableCell class="pr-4">
              <Status variant="text" :tone="auditResultStatus(entry.result).tone">
                {{ auditResultStatus(entry.result).label }}
              </Status>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
