<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { Component } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Link, Database, Layers, HardDrive } from '@lucide/vue';
import { graphql } from '@/gql';
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

// ── Row state ─────────────────────────────────────────────────────────

interface VarRow {
  key: string;
  value: string;
  ref: string | null;
}

const rows = ref<VarRow[]>([]);
const hasChanges = ref(false);
const openPopoverIndex = ref<number | null>(null);

// ── Queries ───────────────────────────────────────────────────────────

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

watch(
  () => result.value?.serviceVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map((v) => ({
        key: v.key,
        value: v.value ?? '',
        ref: v.ref ?? null,
      }));
      hasChanges.value = false;
    }
  },
  { immediate: true },
);

// ── Reference catalog ─────────────────────────────────────────────────

type SourceTypename = 'DatabaseSource' | 'KeyValueStoreSource' | 'BucketSource' | 'SharedSource';

const SOURCE_META: Record<SourceTypename, { label: string; icon: Component }> = {
  DatabaseSource: { label: 'Postgres', icon: Database },
  KeyValueStoreSource: { label: 'Redis', icon: Layers },
  BucketSource: { label: 'Object Storage', icon: HardDrive },
  SharedSource: { label: 'Shared', icon: Link },
};

function sourceIcon(typename: SourceTypename): Component {
  return SOURCE_META[typename].icon;
}

function sourceLabel(typename: SourceTypename, name: string): string {
  const label = SOURCE_META[typename].label;
  return name ? `${label}: ${name}` : label;
}

interface RefOption {
  id: string;
  key: string;
  typename: SourceTypename;
  sourceName: string;
}

const availableRefs = computed<RefOption[]>(() =>
  (availableResult.value?.availableVariables ?? []).map((v) => ({
    id: v.id,
    key: v.key,
    typename: v.source.__typename as SourceTypename,
    sourceName: v.source.name,
  })),
);

interface RefGroup {
  key: string;
  typename: SourceTypename;
  label: string;
  items: RefOption[];
}

const refGroups = computed<RefGroup[]>(() => {
  const map = new Map<string, RefGroup>();
  for (const opt of availableRefs.value) {
    const key = `${opt.typename}:${opt.sourceName}`;
    let group = map.get(key);
    if (!group) {
      group = { key, typename: opt.typename, label: sourceLabel(opt.typename, opt.sourceName), items: [] };
      map.set(key, group);
    }
    group.items.push(opt);
  }
  return [...map.values()];
});

const refById = computed(() => {
  const map = new Map<string, RefOption>();
  for (const opt of availableRefs.value) {
    map.set(opt.id, opt);
  }
  return map;
});

// ── Row actions ───────────────────────────────────────────────────────

function addRow() {
  rows.value.push({ key: '', value: '', ref: null });
  hasChanges.value = true;
}

function selectRef(index: number, opt: RefOption) {
  const row = rows.value[index]!;
  row.ref = opt.id;
  row.value = '';
  if (!row.key.trim()) {
    row.key = opt.key.toUpperCase();
  }
  hasChanges.value = true;
  openPopoverIndex.value = null;
}

function clearRef(index: number) {
  rows.value[index]!.ref = null;
  hasChanges.value = true;
}

function removeRow(index: number) {
  rows.value.splice(index, 1);
  hasChanges.value = true;
}

function markChanged() {
  hasChanges.value = true;
}

function isRefRow(row: VarRow): boolean {
  return !!row.ref;
}

function refDisplay(row: VarRow): string {
  if (!row.ref) return '';
  const opt = refById.value.get(row.ref);
  if (opt) return `${sourceLabel(opt.typename, opt.sourceName)} · ${opt.key}`;
  return row.ref.split('/').pop() ?? row.ref;
}

// ── Save ──────────────────────────────────────────────────────────────

const { mutate: setVarsMutate, loading: saving } = useMutation(SetServiceVariablesDocument);

async function handleSave() {
  const validRows = rows.value.filter((r) => r.key.trim());
  try {
    const variables = validRows.map((r) => ({
      key: r.key.trim(),
      value: r.ref ? undefined : r.value,
      ref: r.ref ?? undefined,
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
                <Link :size="14" :class="row.ref ? '' : 'opacity-50'" />
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
                  <CommandGroup
                    v-for="group in refGroups"
                    :key="group.key"
                  >
                    <template #heading>
                      <div class="flex items-center gap-1.5">
                        <component :is="sourceIcon(group.typename)" :size="12" />
                        {{ group.label }}
                      </div>
                    </template>
                    <CommandItem
                      v-for="opt in group.items"
                      :key="opt.id"
                      :value="opt.id"
                      class="font-mono text-xs"
                      @select="selectRef(index, opt)"
                    >
                      {{ opt.key }}
                    </CommandItem>
                  </CommandGroup>
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
          {{ refDisplay(row) }}
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

    <Separator />

    <p class="text-sm text-muted-foreground">
      These variables are available at both build time and runtime.
    </p>
  </div>
</template>
