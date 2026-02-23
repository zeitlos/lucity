<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Link, Database, Globe, ChevronDown } from 'lucide-vue-next';
import { ServiceVariablesQuery, SetServiceVariablesMutation, SharedVariablesQuery } from '@/graphql/variables';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
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
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
  };
}>();

const { activeEnvironment } = useEnvironment();
const envName = computed(() => activeEnvironment.value?.name ?? '');

// ── Data types ────────────────────────────────────────────────────────

interface DatabaseRefData {
  database: string;
  key: string;
}

interface ServiceRefData {
  service: string;
}

interface VarRow {
  key: string;
  value: string;
  fromShared: boolean;
  databaseRef?: DatabaseRefData;
  serviceRef?: ServiceRefData;
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

const { result, loading, refetch } = useQuery(ServiceVariablesQuery, () => ({
  projectId: props.projectId,
  environment: envName.value,
  service: props.service.name,
}), () => ({
  enabled: !!envName.value,
}));

const { result: sharedResult } = useQuery(SharedVariablesQuery, () => ({
  projectId: props.projectId,
  environment: envName.value,
}), () => ({
  enabled: !!envName.value,
}));

// ── Reference option model ────────────────────────────────────────────

interface RefOption {
  type: 'database' | 'service' | 'shared';
  key: string;
  displayName: string;
  displayValue: string;
  group: string;
  groupIcon: 'database' | 'globe' | 'link';
  databaseRef?: DatabaseRefData;
  serviceRef?: ServiceRefData;
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
        key: `${db.name}-${exp.key}`,
        displayName: exp.displayName,
        displayValue: `\${{${capitalize(db.name)}.${exp.displayName}}}`,
        group: `${capitalize(db.name)} (Postgres)`,
        groupIcon: 'database',
        databaseRef: { database: db.name, key: exp.key },
      });
    }
  }

  // Service references
  const services = activeEnvironment.value?.services ?? [];
  for (const svc of services) {
    if (svc.name === props.service.name) continue;
    const envKey = svc.name.toUpperCase().replace(/-/g, '_');
    options.push({
      type: 'service',
      key: `svc-${svc.name}`,
      displayName: `${envKey}_URL`,
      displayValue: `\${{${svc.name}.URL}}`,
      group: capitalize(svc.name),
      groupIcon: 'globe',
      serviceRef: { service: svc.name },
    });
  }

  // Shared variable references
  const sharedVars = sharedResult.value?.sharedVariables ?? [];
  for (const v of sharedVars as { key: string; value: string }[]) {
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
  const groups: Record<string, { icon: 'database' | 'globe' | 'link'; items: RefOption[] }> = {};
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
const refPopoverOpen = ref(false);

watch(
  () => result.value?.serviceVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map((v: {
        key: string;
        value: string;
        fromShared: boolean;
        databaseRef?: DatabaseRefData;
        serviceRef?: ServiceRefData;
      }) => ({
        key: v.key,
        value: v.value,
        fromShared: v.fromShared,
        databaseRef: v.databaseRef ? { database: v.databaseRef.database, key: v.databaseRef.key } : undefined,
        serviceRef: v.serviceRef ? { service: v.serviceRef.service } : undefined,
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

function addRef(opt: RefOption) {
  const row: VarRow = {
    key: opt.displayName,
    value: opt.displayValue,
    fromShared: opt.type === 'shared',
    databaseRef: opt.databaseRef,
    serviceRef: opt.serviceRef,
    isNew: true,
  };
  rows.value.push(row);
  hasChanges.value = true;
  refPopoverOpen.value = false;
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
  return !!row.databaseRef || !!row.serviceRef || row.fromShared;
}

function refBadgeLabel(row: VarRow): string {
  if (row.databaseRef) return row.databaseRef.database;
  if (row.serviceRef) return row.serviceRef.service;
  return 'shared';
}

// ── Save ──────────────────────────────────────────────────────────────

const { mutate: setVarsMutate, loading: saving } = useMutation(SetServiceVariablesMutation);

async function handleSave() {
  const validRows = rows.value.filter(r => r.key.trim());
  try {
    const variables = validRows.map(r => ({
      key: r.key.trim(),
      value: (!r.databaseRef && !r.serviceRef && !r.fromShared) ? r.value : undefined,
      fromShared: r.fromShared || undefined,
      databaseRef: r.databaseRef || undefined,
      serviceRef: r.serviceRef || undefined,
    }));

    const res = await setVarsMutate({
      projectId: props.projectId,
      environment: envName.value,
      service: props.service.name,
      variables,
    });

    if (res?.errors?.length) {
      toast.error('Failed to save variables', {
        description: res.errors.map((e: { message: string }) => e.message).join(', '),
      });
      return;
    }

    toast.success('Service variables saved');
    hasChanges.value = false;
    refetch();
  } catch (e: unknown) {
    toast.error('Failed to save variables', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-4">
    <div>
      <h3 class="text-sm font-medium text-foreground">Service Variables</h3>
      <p class="text-xs text-muted-foreground">
        Environment variables for <strong>{{ service.name }}</strong> in {{ envName || 'this environment' }}.
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
        <!-- Reference row (database, service, or shared) -->
        <template v-if="isRefRow(row)">
          <Input
            v-model="row.key"
            placeholder="KEY"
            class="flex-1 font-mono text-sm uppercase"
            @input="markChanged"
          />
          <div class="flex flex-1 items-center gap-2">
            <div
              class="flex h-9 flex-1 items-center rounded-md bg-muted px-3 font-mono text-sm text-muted-foreground"
            >
              {{ row.value }}
            </div>
            <Badge
              variant="secondary"
              class="shrink-0 gap-1 text-xs"
            >
              <Database v-if="row.databaseRef" :size="10" />
              <Globe v-else-if="row.serviceRef" :size="10" />
              <Link v-else :size="10" />
              {{ refBadgeLabel(row) }}
            </Badge>
          </div>
        </template>

        <!-- Direct variable row -->
        <template v-else>
          <Input
            v-model="row.key"
            placeholder="KEY"
            class="flex-1 font-mono text-sm uppercase"
            @input="markChanged"
          />
          <Input
            v-model="row.value"
            placeholder="value"
            class="flex-1 font-mono text-sm"
            @input="markChanged"
          />
        </template>

        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive"
          @click="removeRow(index)"
        >
          <Trash2 :size="14" />
        </Button>
      </div>

      <div v-if="rows.length === 0" class="rounded-lg border border-dashed p-6 text-center">
        <p class="text-sm text-muted-foreground">No variables configured for this service.</p>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-between pt-2">
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" @click="addRow">
            <Plus :size="14" class="mr-1" />
            Add Variable
          </Button>

          <Popover v-model:open="refPopoverOpen">
            <PopoverTrigger as-child>
              <Button
                v-if="availableRefs.length > 0"
                variant="outline"
                size="sm"
              >
                <Link :size="14" class="mr-1" />
                Add Reference
                <ChevronDown :size="12" class="ml-1 opacity-50" />
              </Button>
            </PopoverTrigger>
            <PopoverContent
              class="w-80 p-0"
              align="start"
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
                          <Globe v-else-if="group.icon === 'globe'" :size="12" />
                          <Link v-else :size="12" />
                          {{ groupName }}
                        </div>
                      </template>
                      <CommandItem
                        v-for="opt in group.items"
                        :key="opt.key"
                        :value="opt.key"
                        class="flex items-center justify-between"
                        @select="addRef(opt)"
                      >
                        <span class="font-mono text-xs">{{ opt.displayName }}</span>
                        <span class="text-xs text-muted-foreground">{{ opt.displayValue }}</span>
                      </CommandItem>
                    </CommandGroup>
                  </template>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
        </div>

        <div class="flex items-center gap-2">
          <Badge v-if="hasChanges" variant="secondary" class="text-xs">
            Unsaved changes
          </Badge>
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
  </div>
</template>
