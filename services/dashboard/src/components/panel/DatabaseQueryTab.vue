<script setup lang="ts">
import { formatCount } from '@/lib/format';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { ref } from 'vue';
import { useApolloClient } from '@vue/apollo-composable';
import { Play, CircleAlert } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { graphql } from '@/gql';

const ExecuteQueryDocument = graphql(`
  mutation ExecuteQuery($database: DatabaseID!, $query: String!) {
    executeQuery(database: $database, query: $query) {
      columns
      rows
      affectedRows
    }
  }
`);
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';

const props = defineProps<{
  databaseId: string;
}>();

const { resolveClient } = useApolloClient();

const queryText = ref('');
const columns = ref<string[]>([]);
const rows = ref<(string | null)[][]>([]);
const affectedRows = ref(0);
const error = ref<string | null>(null);
const loading = ref(false);
const hasRun = ref(false);
const isMac = typeof window !== 'undefined' && window.navigator.platform?.includes('Mac');

async function executeQuery() {
  const sql = queryText.value.trim();
  if (!sql) return;

  loading.value = true;
  error.value = null;
  hasRun.value = true;

  try {
    const client = resolveClient();
    const { data } = await client.mutate({
      mutation: ExecuteQueryDocument,
      variables: {
        database: props.databaseId,
        query: sql,
      },
    });

    columns.value = data?.executeQuery.columns ?? [];
    rows.value = (data?.executeQuery.rows ?? []).filter((r): r is (string | null)[] => r !== null);
    affectedRows.value = data?.executeQuery.affectedRows ?? 0;
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    const gqlErrors = (e as { graphQLErrors?: { extensions?: { code?: string } }[] }).graphQLErrors;
    if (gqlErrors?.some(err => err.extensions?.code === 'DATABASE_PROVISIONING')) {
      error.value = 'Database is still provisioning. Please wait for PostgreSQL to become ready (~30–60s).';
    } else {
      error.value = msg;
    }
    columns.value = [];
    rows.value = [];
    affectedRows.value = 0;
  } finally {
    loading.value = false;
  }
}

function handleKeydown(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
    event.preventDefault();
    executeQuery();
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- Query input -->
    <div class="space-y-2">
      <Textarea
        v-model="queryText"
        placeholder="SELECT * FROM users LIMIT 10;"
        class="min-h-30 font-mono text-sm"
        @keydown="handleKeydown"
      />
      <div class="flex items-center justify-between">
        <span class="text-xs text-muted-foreground">
          {{ isMac ? '&#8984;' : 'Ctrl' }}+Enter to run
        </span>
        <Button
          :disabled="!queryText.trim() || loading"
          @click="executeQuery"
        >
          <Play v-if="!loading" />
          <Spinner v-else :size="14" class="animate-spin" />
          Run query
        </Button>
      </div>
    </div>

    <Alert v-if="error" variant="danger">
      <CircleAlert />
      <AlertTitle>Query failed</AlertTitle>
      <AlertDescription class="font-mono text-xs">{{ error }}</AlertDescription>
    </Alert>

    <!-- Results -->
    <template v-if="hasRun && !error && !loading">
      <!-- Stats -->
      <div class="flex items-center gap-2">
        <Badge v-if="columns.length > 0" variant="secondary" class="text-xs">
          {{ formatCount(rows.length) }} {{ rows.length === 1 ? 'row' : 'rows' }}
        </Badge>
        <Badge v-if="affectedRows > 0" variant="outline" class="text-xs">
          {{ formatCount(affectedRows) }} affected
        </Badge>
        <Badge v-if="columns.length === 0 && affectedRows === 0" variant="secondary" class="text-xs">
          Query executed
        </Badge>
      </div>

      <!-- Results table -->
      <ScrollArea v-if="columns.length > 0" class="-mx-4 max-h-100 border-y">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead
                v-for="col in columns"
                :key="col"
                class="whitespace-nowrap font-mono text-xs"
              >
                {{ col }}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="rows.length === 0" :colspan="columns.length">
              No rows returned
            </TableEmpty>
            <TableRow v-for="(row, i) in rows" :key="i">
              <TableCell
                v-for="(cell, j) in row"
                :key="j"
                class="max-w-75 truncate whitespace-nowrap font-mono text-xs"
              >
                <span v-if="cell === null" class="italic text-muted-foreground">NULL</span>
                <span v-else>{{ cell }}</span>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
        <ScrollBar orientation="horizontal" />
      </ScrollArea>
    </template>
  </div>
</template>
