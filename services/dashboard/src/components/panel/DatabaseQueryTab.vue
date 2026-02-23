<script setup lang="ts">
import { ref } from 'vue';
import { useApolloClient } from '@vue/apollo-composable';
import { Play, Loader2 } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { ExecuteQueryMutation } from '@/graphql/databases';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';

const props = defineProps<{
  projectId: string;
  database: {
    name: string;
    version: string;
    instances: number;
    size: string;
  };
}>();

const { activeEnvironment } = useEnvironment();
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
  if (!sql || !activeEnvironment.value) return;

  loading.value = true;
  error.value = null;
  hasRun.value = true;

  try {
    const client = resolveClient();
    const { data } = await client.mutate({
      mutation: ExecuteQueryMutation,
      variables: {
        input: {
          projectId: props.projectId,
          environment: activeEnvironment.value.name,
          database: props.database.name,
          query: sql,
        },
      },
    });

    columns.value = data.executeQuery.columns ?? [];
    rows.value = data.executeQuery.rows ?? [];
    affectedRows.value = data.executeQuery.affectedRows ?? 0;
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : String(e);
    // Detect provisioning error from GraphQL extension
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
    <!-- No environment selected -->
    <div
      v-if="!activeEnvironment"
      class="flex flex-col items-center justify-center gap-2 py-12 text-center"
    >
      <p class="text-sm text-muted-foreground">Select an environment to run queries.</p>
    </div>

    <template v-else>
      <!-- Query input -->
      <div class="space-y-2">
        <Textarea
          v-model="queryText"
          placeholder="SELECT * FROM users LIMIT 10;"
          class="min-h-[120px] font-mono text-sm"
          @keydown="handleKeydown"
        />
        <div class="flex items-center justify-between">
          <span class="text-xs text-muted-foreground">
            {{ isMac ? '&#8984;' : 'Ctrl' }}+Enter to run
          </span>
          <Button
            size="sm"
            :disabled="!queryText.trim() || loading"
            @click="executeQuery"
          >
            <Play v-if="!loading" :size="14" class="mr-1" />
            <Loader2 v-else :size="14" class="mr-1 animate-spin" />
            Run Query
          </Button>
        </div>
      </div>

      <!-- Error -->
      <div
        v-if="error"
        class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
      >
        <p class="font-mono text-xs text-destructive">{{ error }}</p>
      </div>

      <!-- Results -->
      <template v-if="hasRun && !error && !loading">
        <!-- Stats -->
        <div class="flex items-center gap-2">
          <Badge v-if="columns.length > 0" variant="secondary" class="text-xs">
            {{ rows.length }} {{ rows.length === 1 ? 'row' : 'rows' }}
          </Badge>
          <Badge v-if="affectedRows > 0" variant="outline" class="text-xs">
            {{ affectedRows }} affected
          </Badge>
          <Badge v-if="columns.length === 0 && affectedRows === 0" variant="secondary" class="text-xs">
            Query executed
          </Badge>
        </div>

        <!-- Results table -->
        <ScrollArea v-if="columns.length > 0" class="max-h-[400px] rounded-lg border">
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
              <TableRow v-if="rows.length === 0">
                <TableCell
                  :colspan="columns.length"
                  class="py-8 text-center text-sm text-muted-foreground"
                >
                  No rows returned
                </TableCell>
              </TableRow>
              <TableRow v-for="(row, i) in rows" :key="i">
                <TableCell
                  v-for="(cell, j) in row"
                  :key="j"
                  class="max-w-[300px] truncate whitespace-nowrap font-mono text-xs"
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
    </template>
  </div>
</template>
