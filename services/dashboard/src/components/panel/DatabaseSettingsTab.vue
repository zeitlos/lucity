<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Trash2, Database, Server, HardDrive, Cpu, MemoryStick } from '@lucide/vue';
import { graphql } from '@/gql';
import type { ResourcesInput } from '@/gql/graphql';

const DatabaseResourcesDocument = graphql(`
  query DatabaseResources($database: DatabaseID!) {
    database(id: $database) {
      id
      resources {
        cpu
        memory
      }
    }
  }
`);

const SetDatabaseResourcesDocument = graphql(`
  mutation SetDatabaseResources($database: DatabaseID!, $resources: ResourcesInput!) {
    setDatabaseResources(database: $database, resources: $resources) {
      id
      resources {
        cpu
        memory
      }
    }
  }
`);

const DeleteDatabaseDocument = graphql(`
  mutation DeleteDatabase($database: DatabaseID!) {
    deleteDatabase(database: $database)
  }
`);
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  databaseId: string;
  database: {
    name: string;
    version: string;
    instances: number;
    size: string;
  };
}>();

const emit = defineEmits<{
  (e: 'database-removed'): void;
}>();

// Compute (vertical scaling). Minimums sized for PostgreSQL.
const cpuOptions = [
  { value: '500m', label: '0.5 vCPU' },
  { value: '1', label: '1 vCPU' },
  { value: '2', label: '2 vCPU' },
  { value: '4', label: '4 vCPU' },
];

const memoryOptions = [
  { value: '512Mi', label: '512 MB' },
  { value: '1Gi', label: '1 GB' },
  { value: '2Gi', label: '2 GB' },
  { value: '4Gi', label: '4 GB' },
  { value: '8Gi', label: '8 GB' },
];

const { result: resourcesResult, refetch: refetchResources } = useQuery(DatabaseResourcesDocument, {
  database: props.databaseId,
});

const resources = computed(() => resourcesResult.value?.database.resources ?? null);

const selectedCpu = ref('');
const selectedMemory = ref('');
const resourcesSaving = ref(false);

const { mutate: setResourcesMutate } = useMutation(SetDatabaseResourcesDocument);

watch(
  resources,
  value => {
    selectedCpu.value = value && value.cpu !== '0' ? value.cpu : '';
    selectedMemory.value = value && value.memory !== '0' ? value.memory : '';
  },
  { immediate: true },
);

const resourcesChanged = computed(() =>
  !!selectedCpu.value
  && !!selectedMemory.value
  && (selectedCpu.value !== resources.value?.cpu || selectedMemory.value !== resources.value?.memory),
);

async function handleSaveResources() {
  resourcesSaving.value = true;
  try {
    const input: ResourcesInput = {
      cpu: selectedCpu.value,
      memory: selectedMemory.value,
    };

    const res = await setResourcesMutate({ database: props.databaseId, resources: input });

    if (res?.errors?.length) {
      errorToast('Failed to update compute', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success('Compute updated');
    await refetchResources();
  } catch (e: unknown) {
    errorToast('Failed to update compute', { description: errorMessage(e) });
  } finally {
    resourcesSaving.value = false;
  }
}

const { mutate: deleteDatabase, loading: deleting } = useMutation(DeleteDatabaseDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteDatabase({ database: props.databaseId });
    toast.success('Database removed');
    deleteDialogOpen.value = false;
    emit('database-removed');
  } catch (e: unknown) {
    errorToast('Failed to remove database', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Configuration -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Configuration
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <Database :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Version</span>
            <span class="font-mono text-sm font-medium text-foreground">PostgreSQL {{ database.version }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <Server :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Instances</span>
            <span class="text-sm font-medium text-foreground">{{ database.instances }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Storage</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ database.size }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Compute -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Compute
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <Cpu :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">CPU</p>
              <p class="text-xs text-muted-foreground">Limit per instance</p>
            </div>
            <Select v-model="selectedCpu">
              <SelectTrigger class="h-8 w-36">
                <SelectValue placeholder="Select" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in cpuOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <MemoryStick :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Memory</p>
              <p class="text-xs text-muted-foreground">Limit per instance</p>
            </div>
            <Select v-model="selectedMemory">
              <SelectTrigger class="h-8 w-36">
                <SelectValue placeholder="Select" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in memoryOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center justify-between gap-3 px-4 py-3">
            <p class="text-xs text-muted-foreground">
              Applied across all {{ database.instances }} instances. Saving triggers a rolling restart.
            </p>
            <Button
              size="sm"
              :disabled="!resourcesChanged || resourcesSaving"
              @click="handleSaveResources"
            >
              {{ resourcesSaving ? 'Saving...' : 'Save' }}
            </Button>
          </div>
        </div>
      </div>
    </section>

    <!-- Danger Zone -->
    <section class="mt-8">
      <div class="relative overflow-hidden rounded-lg border border-destructive/20">
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.04]" />
        <div class="relative border-b border-destructive/15 bg-destructive/[0.03] px-4 py-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-destructive/70">
            Danger Zone
          </h3>
        </div>
        <div class="relative px-4 py-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-foreground">Delete Database</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this database and all its data.
              </p>
            </div>
            <AlertDialog v-model:open="deleteDialogOpen">
              <AlertDialogTrigger as-child>
                <Button variant="destructive" size="sm">
                  <Trash2 :size="14" class="mr-1" />
                  Delete
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete database "{{ database.name }}"?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete the database and all its data.
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    :disabled="deleting"
                    class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                    @click="handleDelete"
                  >
                    {{ deleting ? 'Deleting...' : 'Delete Database' }}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
