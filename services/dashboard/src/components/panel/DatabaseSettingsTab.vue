<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import { Skeleton } from '@/components/ui/skeleton';
import FormFooter from '@/components/FormFooter.vue';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Trash2, Database, Server, HardDrive, Cpu, MemoryStick } from '@lucide/vue';
import { graphql } from '@/gql';
import type { ResourcesInput } from '@/gql/graphql';
import { useCapabilities } from '@/composables/useCapabilities';
import { useGate } from '@/composables/useGate';
import GateTag from '@/components/GateTag.vue';
import { aboveLimit, gatedChoices } from '@/lib/capabilities';
import { formatCpuQuantity, formatStorage } from '@/lib/format';

const DatabaseResourcesDocument = graphql(`
  query DatabaseResources($database: DatabaseID!) {
    database(id: $database) {
      id
      size
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
      status
      resources {
        cpu
        memory
      }
    }
  }
`);

const SetDatabaseStorageDocument = graphql(`
  mutation SetDatabaseStorage($database: DatabaseID!, $size: String!) {
    setDatabaseStorage(database: $database, size: $size) {
      id
      status
      size
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
import { failureToast } from '@/components/ui/sonner';

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

const storageOptions = [
  { value: '10Gi', gib: 10, label: '10 GB' },
  { value: '16Gi', gib: 16, label: '16 GB' },
  { value: '32Gi', gib: 32, label: '32 GB' },
  { value: '64Gi', gib: 64, label: '64 GB' },
  { value: '128Gi', gib: 128, label: '128 GB' },
];

const { result: resourcesResult, error: resourcesError, refetch: refetchResources } = useQuery(DatabaseResourcesDocument, {
  database: props.databaseId,
});

const resources = computed(() => resourcesResult.value?.database.resources ?? null);
const currentSize = computed(() => resourcesResult.value?.database.size ?? props.database.size);

const selectedCpu = ref('');
const selectedMemory = ref('');
const resourcesSaving = ref(false);

const { mutate: setResourcesMutate } = useMutation(SetDatabaseResourcesDocument);

function discardResources() {
  const value = resources.value;
  selectedCpu.value = value && value.cpu !== '0' ? value.cpu : '';
  selectedMemory.value = value && value.memory !== '0' ? value.memory : '';
}

watch(resources, discardResources, { immediate: true });

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

    await setResourcesMutate({ database: props.databaseId, resources: input });

    await refetchResources();
  } catch (e: unknown) {
    failureToast('Failed to update compute', e);
  } finally {
    resourcesSaving.value = false;
  }
}

// Storage can only grow: a PVC cannot be shrunk.
const parseGib = (size: string) => {
  const match = /^(\d+)(Gi|Ti)$/.exec(size);
  return match ? Number(match[1]) * (match[2] === 'Ti' ? 1024 : 1) : 0;
};
const { capabilities } = useCapabilities();
const { pick } = useGate();

const cpuChoices = computed(() => gatedChoices(cpuOptions, capabilities.value?.databases.cpu, resources.value?.cpu));
const memoryChoices = computed(() => gatedChoices(memoryOptions, capabilities.value?.databases.memory, resources.value?.memory));
const cpuAboveLimit = computed(() => aboveLimit(capabilities.value?.databases.cpu, resources.value?.cpu));
const memoryAboveLimit = computed(() => aboveLimit(capabilities.value?.databases.memory, resources.value?.memory));

const availableStorageOptions = computed(() => gatedChoices(
  storageOptions.filter(option => option.gib >= parseGib(currentSize.value)),
  capabilities.value?.databases.storage,
  currentSize.value,
));

function pickCpu(value: unknown) {
  pick(cpuChoices.value, value, v => { selectedCpu.value = v; });
}

function pickMemory(value: unknown) {
  pick(memoryChoices.value, value, v => { selectedMemory.value = v; });
}

function pickStorage(value: unknown) {
  pick(availableStorageOptions.value, value, v => { selectedStorage.value = v; });
}

const selectedStorage = ref('');
const storageSaving = ref(false);

const { mutate: setStorageMutate } = useMutation(SetDatabaseStorageDocument);

function discardStorage() {
  selectedStorage.value = currentSize.value;
}

watch(currentSize, discardStorage, { immediate: true });

const storageChanged = computed(() =>
  !!selectedStorage.value && selectedStorage.value !== currentSize.value,
);

useUnsavedChanges(() => resourcesChanged.value || storageChanged.value);

async function handleSaveStorage() {
  storageSaving.value = true;
  try {
    await setStorageMutate({ database: props.databaseId, size: selectedStorage.value });

    await refetchResources();
  } catch (e: unknown) {
    failureToast('Failed to update storage', e);
  } finally {
    storageSaving.value = false;
  }
}

const { mutate: deleteDatabase, loading: deleting } = useMutation(DeleteDatabaseDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteDatabase({ database: props.databaseId });
    deleteDialogOpen.value = false;
    emit('database-removed');
  } catch (e: unknown) {
    failureToast('Failed to delete database', e);
  }
}
</script>

<template>
  <QueryError
    v-if="resourcesError && !resourcesResult"
    title="Couldn't load database settings"
    :error="resourcesError"
    @retry="refetchResources()"
  />

  <div v-else-if="!resourcesResult" class="space-y-6">
    <Skeleton class="h-24 w-full" />
    <Skeleton class="h-40 w-full" />
    <Skeleton class="h-28 w-full" />
  </div>

  <div v-else class="space-y-6">
    <!-- Configuration -->
    <section class="space-y-2">
      <h3 class="text-sm font-medium text-muted-foreground">
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
        </div>
      </div>
    </section>

    <!-- Compute -->
    <section class="space-y-2">
      <h3 class="text-sm font-medium text-muted-foreground">
        Compute
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div>
          <fieldset :disabled="resourcesSaving" class="min-w-0 divide-y">
            <div class="flex items-center gap-3 px-4 py-3">
              <Cpu :size="16" class="shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-foreground">CPU</p>
                <p v-if="cpuAboveLimit" class="text-xs text-status-warn">Above the plan's limit of {{ formatCpuQuantity(capabilities!.databases.cpu.max) }}, so it can only shrink</p>
                <p v-else class="text-xs text-muted-foreground">Limit per instance</p>
              </div>
              <Select :model-value="selectedCpu" @update:model-value="pickCpu">
                <SelectTrigger class="h-8 w-36">
                  <SelectValue placeholder="Select" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="option in cpuChoices"
                    :key="option.value"
                    :value="option.value"
                  >
                    <span class="flex w-full items-center justify-between gap-3">
                      {{ option.label }}
                      <GateTag v-if="option.state.kind === 'plan'" :plan="option.state.plan" />
                    </span>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="flex items-center gap-3 px-4 py-3">
              <MemoryStick :size="16" class="shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-foreground">Memory</p>
                <p v-if="memoryAboveLimit" class="text-xs text-status-warn">Above the plan's limit of {{ formatStorage(capabilities!.databases.memory.max) }}, so it can only shrink</p>
                <p v-else class="text-xs text-muted-foreground">Limit per instance</p>
              </div>
              <Select :model-value="selectedMemory" @update:model-value="pickMemory">
                <SelectTrigger class="h-8 w-36">
                  <SelectValue placeholder="Select" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="option in memoryChoices"
                    :key="option.value"
                    :value="option.value"
                  >
                    <span class="flex w-full items-center justify-between gap-3">
                      {{ option.label }}
                      <GateTag v-if="option.state.kind === 'plan'" :plan="option.state.plan" />
                    </span>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </fieldset>
          <FormFooter
            :dirty="resourcesChanged"
            :saving="resourcesSaving"
            :hint="`Applied across all ${database.instances} instances. Saving triggers a rolling restart.`"
            @save="handleSaveResources"
            @discard="discardResources"
          />
        </div>
      </div>
    </section>

    <!-- Storage -->
    <section class="space-y-2">
      <h3 class="text-sm font-medium text-muted-foreground">
        Storage
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div>
          <fieldset :disabled="storageSaving" class="min-w-0 divide-y">
            <div class="flex items-center gap-3 px-4 py-3">
              <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-medium text-foreground">Disk</p>
                <p class="text-xs text-muted-foreground">Volume per instance</p>
              </div>
              <Select :model-value="selectedStorage" @update:model-value="pickStorage">
                <SelectTrigger class="h-8 w-36">
                  <SelectValue placeholder="Select" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="option in availableStorageOptions"
                    :key="option.value"
                    :value="option.value"
                  >
                    <span class="flex w-full items-center justify-between gap-3">
                      {{ option.label }}
                      <GateTag v-if="option.state.kind === 'plan'" :plan="option.state.plan" />
                    </span>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </fieldset>
          <FormFooter
            :dirty="storageChanged"
            :saving="storageSaving"
            hint="You can add more space anytime, but storage can't be made smaller later."
            @save="handleSaveStorage"
            @discard="discardStorage"
          />
        </div>
      </div>
    </section>

    <div class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
      <div>
        <p class="text-sm font-medium text-foreground">Delete database</p>
        <p class="text-xs text-muted-foreground">Permanently delete this database and all its data.</p>
      </div>
      <ConfirmDialog
        v-model:open="deleteDialogOpen"
        :title="`Delete database &quot;${database.name}&quot;?`"
        description="All its data is deleted with it. This can't be undone."
        confirm-label="Delete database"
        pending-label="Deleting…"
        :pending="deleting"
        :confirm-text="database.name"
        destructive
        @confirm="handleDelete"
      >
        <template #trigger>
          <Button variant="destructive">
            <Trash2 />
            Delete
          </Button>
        </template>
      </ConfirmDialog>
    </div>
  </div>
</template>
