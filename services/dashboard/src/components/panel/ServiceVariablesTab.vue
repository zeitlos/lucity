<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Link, Database } from 'lucide-vue-next';
import { graphql } from '@/gql';

const ServiceVariablesDocument = graphql(`
  query ServiceVariables($service: ServiceID!) {
    serviceVariables(service: $service) {
      key
      value
      fromShared
      databaseRef {
        database
        key
      }
    }
  }
`);

const SetServiceVariablesDocument = graphql(`
  mutation SetServiceVariables($service: ServiceID!, $variables: [ServiceVariableInput!]!) {
    setServiceVariables(service: $service, variables: $variables)
  }
`);

const SharedVariablesDocument = graphql(`
  query SharedVariables($environment: EnvironmentID!) {
    sharedVariables(environment: $environment) {
      key
      value
    }
  }
`);
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  serviceId: string;
  serviceName: string;
}>();

const { activeEnvironment } = useEnvironment();
const environmentId = computed(() => activeEnvironment.value?.id ?? '');

// ── Data types ────────────────────────────────────────────────────────

interface DatabaseRefData {
  database: string;
  key: string;
}

interface VarRow {
  key: string;
  value: string;
  fromShared: boolean;
  databaseRef?: DatabaseRefData | null;
  isNew?: boolean;
}

// ── CNPG exports ──────────────────────────────────────────────────────

const CNPG_EXPORTS = [
  { key: 'uri', displayName: 'DATABASE_URL' },
  { key: 'host', displayName: 'PGHOST' },
  { key: 'port', displayName: 'PGPORT' },
  { key: 'dbname', displayName: 'PGDATABASE' },
  { key: 'user', displayName: 'PGUSER' },
  { key: 'password', displayName: 'PGPASSWORD' },
] as const;

// ── Queries ───────────────────────────────────────────────────────────

const { result, loading, refetch } = useQuery(ServiceVariablesDocument, () => ({
  service: props.serviceId,
}), () => ({
  enabled: !!props.serviceId,
}));

const { result: sharedResult } = useQuery(SharedVariablesDocument, () => ({
  environment: environmentId.value,
}), () => ({
  enabled: !!environmentId.value,
}));

// ── Reference option model ────────────────────────────────────────────

interface RefOption {
  type: 'database' | 'shared';
  key: string;
  displayName: string;
  displayValue: string;
  group: string;
  groupIcon: 'database' | 'link';
  databaseRef?: DatabaseRefData;
}

function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

const availableRefs = computed<RefOption[]>(() => {
  const options: RefOption[] = [];

  // Database references
  const databases = activeEnvironment.value?.databases ?? [];
  for (const db of databases) {
    for (const exp of CNPG_EXPORTS) {
      options.push({
        type: 'database',
        key: `${db.id}-${exp.key}`,
        displayName: exp.displayName,
        displayValue: `\${{${capitalize(db.name)}.${exp.displayName}}}`,
        group: `${capitalize(db.name)} (Postgres)`,
        groupIcon: 'database',
        databaseRef: { database: db.id, key: exp.key },
      });
    }
  }

  // Shared variable references
  const sharedVars = sharedResult.value?.sharedVariables ?? [];
  for (const v of sharedVars) {
    options.push({
      type: 'shared',
      key: `shared-${v.key}`,
      displayName: v.key,
      displayValue: v.value,
      group: 'Shared Variables',
      groupIcon: 'link',
    });
  }

  return options;
});

const refGroups = computed(() => {
  const groups: Record<string, { icon: 'database' | 'link'; items: RefOption[] }> = {};
  for (const opt of availableRefs.value) {
    if (!groups[opt.group]) {
      groups[opt.group] = { icon: opt.groupIcon, items: [] };
    }
    groups[opt.group]!.items.push(opt);
  }
  return groups;
});

// ── Row state ─────────────────────────────────────────────────────────

const rows = ref<VarRow[]>([]);
const hasChanges = ref(false);
const openPopoverIndex = ref<number | null>(null);

watch(
  () => result.value?.serviceVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map((v) => ({
        key: v.key,
        value: v.value,
        fromShared: v.fromShared,
        databaseRef: v.databaseRef ? { database: v.databaseRef.database, key: v.databaseRef.key } : undefined,
      }));
      hasChanges.value = false;
    }
  },
  { immediate: true },
);

// ── Row actions ───────────────────────────────────────────────────────

function addRow() {
  rows.value.push({ key: '', value: '', fromShared: false, isNew: true });
  hasChanges.value = true;
}

function selectRef(index: number, opt: RefOption) {
  rows.value[index] = {
    key: opt.displayName,
    value: opt.displayValue,
    fromShared: opt.type === 'shared',
    databaseRef: opt.databaseRef,
  };
  hasChanges.value = true;
  openPopoverIndex.value = null;
}

function clearRef(index: number) {
  const row = rows.value[index]!;
  row.databaseRef = undefined;
  row.fromShared = false;
  row.value = '';
  hasChanges.value = true;
}

function removeRow(index: number) {
  rows.value.splice(index, 1);
  hasChanges.value = true;
}

function markChanged() {
  hasChanges.value = true;
}

// ── Row display helpers ───────────────────────────────────────────────

function isRefRow(row: VarRow): boolean {
  return !!row.databaseRef || row.fromShared;
}

// ── Save ──────────────────────────────────────────────────────────────

const { mutate: setVarsMutate, loading: saving } = useMutation(SetServiceVariablesDocument);

async function handleSave() {
  const validRows = rows.value.filter(r => r.key.trim());
  try {
    const variables = validRows.map(r => ({
      key: r.key.trim(),
      value: (!r.databaseRef && !r.fromShared) ? r.value : undefined,
      fromShared: r.fromShared || undefined,
      databaseRef: r.databaseRef || undefined,
    }));

    const res = await setVarsMutate({
      service: props.serviceId,
      variables,
    });

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
    <div>
      <h3 class="text-sm font-medium text-foreground">Service Variables</h3>
      <p class="text-xs text-muted-foreground">
        Environment variables for <strong>{{ serviceName }}</strong> in {{ activeEnvironment?.name || 'this environment' }}.
        These are available at both build time and runtime.
      </p>
    </div>

    <Separator />

    <!-- Loading state -->
    <div v-if="loading" class="space-y-2">
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>

    <!-- Variable rows -->
    <div v-else class="space-y-2">
      <div
        v-for="(row, index) in rows"
        :key="index"
        class="flex items-center gap-2"
      >
        <!-- Key input with integrated reference picker -->
        <div class="flex flex-1">
          <Input
            v-model="row.key"
            placeholder="KEY"
            class="font-mono text-sm uppercase rounded-r-none border-r-0"
            :readonly="isRefRow(row)"
            @input="markChanged"
          />
          <Popover
            :open="openPopoverIndex === index"
            @update:open="(v: boolean) => openPopoverIndex = v ? index : null"
          >
            <PopoverTrigger as-child>
              <Button
                variant="outline"
                size="icon"
                class="shrink-0 rounded-l-none"
                :disabled="availableRefs.length === 0"
              >
                <Database v-if="row.databaseRef" :size="14" />
                <Link v-else :size="14" class="opacity-50" />
              </Button>
            </PopoverTrigger>
            <PopoverContent
              class="w-80 p-0"
              align="end"
            >
              <Command>
                <CommandInput placeholder="Search references..." />
                <CommandList>
                  <CommandEmpty>No references found.</CommandEmpty>
                  <template v-for="(group, groupName) in refGroups" :key="groupName">
                    <CommandGroup>
                      <template #heading>
                        <div class="flex items-center gap-1.5">
                          <Database v-if="group.icon === 'database'" :size="12" />
                          <Link v-else :size="12" />
                          {{ groupName }}
                        </div>
                      </template>
                      <CommandItem
                        v-for="opt in group.items"
                        :key="opt.key"
                        :value="opt.key"
                        class="flex items-center justify-between"
                        @select="selectRef(index, opt)"
                      >
                        <span class="font-mono text-xs">{{ opt.displayName }}</span>
                        <span class="text-xs text-muted-foreground">{{ opt.displayValue }}</span>
                      </CommandItem>
                    </CommandGroup>
                  </template>
                  <!-- Clear reference option -->
                  <CommandGroup v-if="isRefRow(row)">
                    <CommandItem
                      value="__clear__"
                      class="text-muted-foreground"
                      @select="clearRef(index)"
                    >
                      Clear reference
                    </CommandItem>
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
        </div>

        <!-- Value -->
        <div
          v-if="isRefRow(row)"
          class="flex h-9 flex-1 items-center rounded-md border border-input bg-muted px-3 font-mono text-xs text-muted-foreground"
        >
          {{ row.value }}
        </div>
        <Input
          v-else
          v-model="row.value"
          placeholder="value"
          class="flex-1 font-mono text-sm"
          @input="markChanged"
        />

        <!-- Delete -->
        <Button
          variant="ghost"
          size="icon"
          class="h-9 w-9 shrink-0 text-muted-foreground hover:text-destructive"
          @click="removeRow(index)"
        >
          <Trash2 :size="14" />
        </Button>
      </div>

      <!-- Empty state -->
      <div v-if="rows.length === 0" class="rounded-lg border border-dashed p-6 text-center">
        <p class="text-sm text-muted-foreground">No variables configured for this service.</p>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-between pt-2">
        <Button variant="outline" size="sm" @click="addRow">
          <Plus :size="14" class="mr-1" />
          Add Variable
        </Button>

        <Button
          size="sm"
          :disabled="!hasChanges || saving"
          @click="handleSave"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </Button>
      </div>
    </div>
  </div>
</template>
