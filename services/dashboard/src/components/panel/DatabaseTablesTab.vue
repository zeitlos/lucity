<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { useApolloClient } from '@vue/apollo-composable';
import { ArrowLeft, Table2, Key, ChevronLeft, ChevronRight, Loader2 } from 'lucide-vue-next';
import { useEnvironment } from '@/composables/useEnvironment';
import { DatabaseTablesQuery, DatabaseTableDataQuery } from '@/graphql/databases';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Skeleton } from '@/components/ui/skeleton';

const PAGE_SIZE = 50;

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

// Table list query
const queryEnabled = computed(() => !!activeEnvironment.value);
const queryVars = computed(() => ({
  projectId: props.projectId,
  environment: activeEnvironment.value?.name ?? '',
  database: props.database.name,
}));

const { result: tablesResult, loading: tablesLoading, error: tablesError } = useQuery(
  DatabaseTablesQuery,
  queryVars,
  () => ({ enabled: queryEnabled.value }),
);

const tables = computed(() => tablesResult.value?.databaseTables ?? []);

// Selected table for data view
const selectedTable = ref<string | null>(null);
const selectedSchema = ref('public');
const offset = ref(0);
const dataLoading = ref(false);
const dataColumns = ref<string[]>([]);
const dataRows = ref<(string | null)[][]>([]);
const totalEstimatedRows = ref(0);
const dataError = ref<string | null>(null);

function openTable(tableName: string, schema: string) {
  selectedTable.value = tableName;
  selectedSchema.value = schema;
  offset.value = 0;
  fetchData();
}

function closeTable() {
  selectedTable.value = null;
  dataColumns.value = [];
  dataRows.value = [];
  dataError.value = null;
}

async function fetchData() {
  if (!activeEnvironment.value || !selectedTable.value) return;

  dataLoading.value = true;
  dataError.value = null;

  try {
    const client = resolveClient();
    const { data } = await client.query({
      query: DatabaseTableDataQuery,
      variables: {
        projectId: props.projectId,
        environment: activeEnvironment.value.name,
        database: props.database.name,
        table: selectedTable.value,
        schema: selectedSchema.value,
        limit: PAGE_SIZE,
        offset: offset.value,
      },
      fetchPolicy: 'network-only',
    });

    dataColumns.value = data.databaseTableData.columns;
    dataRows.value = data.databaseTableData.rows;
    totalEstimatedRows.value = data.databaseTableData.totalEstimatedRows;
  } catch (e: unknown) {
    dataError.value = e instanceof Error ? e.message : String(e);
  } finally {
    dataLoading.value = false;
  }
}

function nextPage() {
  offset.value += PAGE_SIZE;
  fetchData();
}

function prevPage() {
  offset.value = Math.max(0, offset.value - PAGE_SIZE);
  fetchData();
}

const currentPage = computed(() => Math.floor(offset.value / PAGE_SIZE) + 1);
const hasMore = computed(() => dataRows.value.length === PAGE_SIZE);

// Reset when environment changes
watch(() => activeEnvironment.value?.name, () => {
  if (selectedTable.value) {
    offset.value = 0;
    fetchData();
  }
});
</script>

<template>
  <div class="space-y-4">
    <!-- No environment selected -->
    <div
      v-if="!activeEnvironment"
      class="flex flex-col items-center justify-center gap-2 py-12 text-center"
    >
      <p class="text-sm text-muted-foreground">Select an environment to browse tables.</p>
    </div>

    <!-- Data view (selected table) -->
    <template v-else-if="selectedTable">
      <!-- Header -->
      <div class="flex items-center gap-2">
        <Button variant="ghost" size="icon" class="h-7 w-7" @click="closeTable">
          <ArrowLeft :size="14" />
        </Button>
        <Table2 :size="16" class="text-muted-foreground" />
        <span class="text-sm font-medium">{{ selectedSchema }}.{{ selectedTable }}</span>
        <Badge variant="outline" class="text-xs">~{{ totalEstimatedRows }} rows</Badge>
      </div>

      <!-- Loading -->
      <div v-if="dataLoading" class="space-y-2">
        <Skeleton v-for="i in 5" :key="i" class="h-8 w-full" />
      </div>

      <!-- Error -->
      <div
        v-else-if="dataError"
        class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
      >
        <p class="font-mono text-xs text-destructive">{{ dataError }}</p>
      </div>

      <!-- Data table -->
      <template v-else>
        <ScrollArea class="max-h-[500px] rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead
                  v-for="col in dataColumns"
                  :key="col"
                  class="whitespace-nowrap font-mono text-xs"
                >
                  {{ col }}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="dataRows.length === 0">
                <TableCell :colspan="dataColumns.length" class="py-8 text-center text-sm text-muted-foreground">
                  No rows
                </TableCell>
              </TableRow>
              <TableRow v-for="(row, i) in dataRows" :key="i">
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

        <!-- Pagination -->
        <div class="flex items-center justify-between">
          <span class="text-xs text-muted-foreground">
            Showing {{ offset + 1 }}&ndash;{{ offset + dataRows.length }}
          </span>
          <div class="flex items-center gap-1">
            <Button
              variant="outline"
              size="icon"
              class="h-7 w-7"
              :disabled="offset === 0"
              @click="prevPage"
            >
              <ChevronLeft :size="14" />
            </Button>
            <span class="px-2 text-xs text-muted-foreground">Page {{ currentPage }}</span>
            <Button
              variant="outline"
              size="icon"
              class="h-7 w-7"
              :disabled="!hasMore"
              @click="nextPage"
            >
              <ChevronRight :size="14" />
            </Button>
          </div>
        </div>
      </template>
    </template>

    <!-- Table list -->
    <template v-else>
      <!-- Loading -->
      <div v-if="tablesLoading" class="space-y-2">
        <Skeleton v-for="i in 4" :key="i" class="h-12 w-full" />
      </div>

      <!-- Error -->
      <div
        v-else-if="tablesError"
        class="rounded-lg border border-destructive/30 bg-destructive/5 p-3"
      >
        <p class="font-mono text-xs text-destructive">{{ tablesError.message }}</p>
      </div>

      <!-- Empty state -->
      <div
        v-else-if="tables.length === 0"
        class="flex flex-col items-center justify-center gap-2 py-12 text-center"
      >
        <Table2 :size="24" class="text-muted-foreground" />
        <p class="text-sm text-muted-foreground">No tables found in this database.</p>
      </div>

      <!-- Table list -->
      <div v-else class="space-y-1">
        <button
          v-for="table in tables"
          :key="`${table.schema}.${table.name}`"
          class="flex w-full items-center gap-3 rounded-lg border px-3 py-2.5 text-left transition-colors hover:bg-accent/50"
          @click="openTable(table.name, table.schema)"
        >
          <Table2 :size="16" class="shrink-0 text-muted-foreground" />
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-medium text-foreground">{{ table.name }}</span>
              <Badge v-if="table.schema !== 'public'" variant="outline" class="text-xs">
                {{ table.schema }}
              </Badge>
            </div>
            <div class="flex items-center gap-3 text-xs text-muted-foreground">
              <span>~{{ table.estimatedRows }} rows</span>
              <span>{{ table.columns.length }} columns</span>
              <span
                v-if="table.columns.some((c: { primaryKey: boolean }) => c.primaryKey)"
                class="flex items-center gap-0.5"
              >
                <Key :size="10" />
                {{ table.columns.filter((c: { primaryKey: boolean }) => c.primaryKey).map((c: { name: string }) => c.name).join(', ') }}
              </span>
            </div>
          </div>
        </button>
      </div>
    </template>

    <!-- Loading overlay for pagination -->
    <div
      v-if="selectedTable && dataLoading && dataColumns.length > 0"
      class="flex items-center justify-center gap-2 py-2"
    >
      <Loader2 :size="14" class="animate-spin text-muted-foreground" />
      <span class="text-xs text-muted-foreground">Loading...</span>
    </div>
  </div>
</template>
