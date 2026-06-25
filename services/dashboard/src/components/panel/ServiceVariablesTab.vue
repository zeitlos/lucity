<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Link } from '@lucide/vue';
import SourceIcon from '@/components/SourceIcon.vue';
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
import { parseEnvAssignments } from '@/lib/env';

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

const SOURCE_LABELS: Record<SourceTypename, string> = {
  DatabaseSource: 'Postgres',
  KeyValueStoreSource: 'Redis',
  BucketSource: 'Object Storage',
  SharedSource: 'Shared',
};

function sourceLabel(typename: SourceTypename, name: string): string {
  const label = SOURCE_LABELS[typename];
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

function onPasteKey(event: ClipboardEvent, index: number) {
  const parsed = parseEnvAssignments(event.clipboardData?.getData('text') ?? '');

  if (parsed.length === 0) return;

  event.preventDefault();
  rows.value.splice(index, 1, ...parsed.map((p) => ({ key: p.key, value: p.value, ref: null })));
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

function refInfo(row: VarRow): { typename: SourceTypename | null; label: string; key: string } {
  const opt = row.ref ? refById.value.get(row.ref) : undefined;

  if (opt) {
    return { typename: opt.typename, label: opt.sourceName || SOURCE_LABELS[opt.typename], key: opt.key };
  }

  return { typename: null, label: 'Reference', key: row.ref?.split('/').pop() ?? '' };
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
            @paste="onPasteKey($event, index)"
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
                    <div class="flex items-center gap-1.5 px-2 py-1.5 text-xs font-medium text-muted-foreground">
                      <SourceIcon :typename="group.typename" />
                      {{ group.label }}
                    </div>
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
          class="magic-border flex h-10 flex-1 items-center gap-1.5 rounded-md px-3"
        >
          <SourceIcon :typename="refInfo(row).typename" />
          <span class="text-xs font-medium">{{ refInfo(row).label }}</span>
          <span class="text-muted-foreground/40">·</span>
          <span class="font-mono text-xs text-muted-foreground">{{ refInfo(row).key }}</span>
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

<style scoped>
.magic-border {
  border: 1.5px solid transparent;
  background:
    linear-gradient(var(--background), var(--background)) padding-box,
    linear-gradient(90deg, #f43f5e, #f59e0b, #facc15, #22c55e, #3b82f6, #a855f7, #f43f5e) border-box;
  background-size:
    100% 100%,
    200% 100%;
  animation: magic-flow 5s linear infinite;
}

@keyframes magic-flow {
  to {
    background-position:
      0 0,
      -200% 0;
  }
}
</style>
