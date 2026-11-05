<script setup lang="ts">
import { computed, nextTick, ref } from 'vue';
import { Eye, EyeOff, Link, Plus, Trash2, Undo2, Unlink, Braces } from '@lucide/vue';
import EmptyState from '@/components/EmptyState.vue';
import IconButton from '@/components/IconButton.vue';
import SourceIcon from '@/components/SourceIcon.vue';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Input } from '@/components/ui/input';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { parseEnvAssignments } from '@/lib/env';
import { darkTooltipClass } from '@/lib/tooltip';
import { duplicateKeys, type VariableReference, type VariableRow, type VariableSourceTypename } from '@/lib/variables';

const rows = defineModel<VariableRow[]>('rows', { required: true });

const props = withDefaults(defineProps<{
  references?: VariableReference[];
  loading?: boolean;
  disabled?: boolean;
  emptyTitle: string;
  emptyDescription?: string;
}>(), {
  emptyDescription: undefined,
  references: undefined,
  loading: false,
  disabled: false,
});

const SOURCE_LABELS: Record<VariableSourceTypename, string> = {
  DatabaseSource: 'Postgres',
  KeyValueStoreSource: 'Redis',
  BucketSource: 'Object Storage',
  SharedSource: 'Shared',
};

interface ReferenceGroup {
  key: string;
  typename: VariableSourceTypename;
  label: string;
  items: VariableReference[];
}

const openPopoverIndex = ref<number | null>(null);
const root = ref<HTMLFieldSetElement | null>(null);

const linkable = computed(() => props.references !== undefined);

const skeletonRows = [
  { key: 'w-28', value: 'w-40' },
  { key: 'w-36', value: 'w-24' },
  { key: 'w-20', value: 'w-32' },
];

const referenceGroups = computed<ReferenceGroup[]>(() => {
  const groups = new Map<string, ReferenceGroup>();
  for (const reference of props.references ?? []) {
    const key = `${reference.typename}:${reference.sourceName}`;
    let group = groups.get(key);
    if (!group) {
      const label = SOURCE_LABELS[reference.typename];
      group = { key, typename: reference.typename, label: reference.sourceName ? `${label}: ${reference.sourceName}` : label, items: [] };
      groups.set(key, group);
    }
    group.items.push(reference);
  }
  return [...groups.values()];
});

const referenceById = computed(() => new Map((props.references ?? []).map(reference => [reference.id, reference])));

const valueRows = computed(() => rows.value.filter(row => !row.ref && !row.removed));
const allRevealed = computed(() => valueRows.value.length > 0 && valueRows.value.every(row => row.revealed));

const duplicates = computed(() => new Set(duplicateKeys(rows.value)));
const repeated = computed(() => {
  const seen = new Set<string>();
  return rows.value.map(row => {
    const key = row.key.trim();
    if (row.removed || !key) return false;
    const repeat = seen.has(key);
    seen.add(key);
    return repeat;
  });
});

function isDuplicate(row: VariableRow) {
  return !row.removed && duplicates.value.has(row.key.trim());
}

async function addRow() {
  rows.value.push({ key: '', value: '', ref: null, revealed: true });
  await nextTick();
  const keys = root.value?.querySelectorAll<HTMLInputElement>('input[aria-label="Key"]');
  keys?.[keys.length - 1]?.focus();
}

function removeRow(index: number) {
  const row = rows.value[index]!;
  if (row.persisted) {
    row.removed = true;
  } else {
    rows.value.splice(index, 1);
  }
}

function restoreRow(index: number) {
  rows.value[index]!.removed = false;
}

function toggleAll() {
  const next = !allRevealed.value;
  for (const row of rows.value) {
    if (!row.ref) row.revealed = next;
  }
}

function onPasteKey(event: ClipboardEvent, index: number) {
  if (rows.value[index]?.removed) return;
  const parsed = parseEnvAssignments(event.clipboardData?.getData('text') ?? '');
  if (parsed.length === 0) return;
  event.preventDefault();
  rows.value.splice(index, 1, ...parsed.map(p => ({ key: p.key, value: p.value, ref: null })));
}

function selectReference(index: number, reference: VariableReference) {
  const row = rows.value[index]!;
  row.ref = reference.id;
  row.value = '';
  if (!row.key.trim()) row.key = reference.key.toUpperCase();
  openPopoverIndex.value = null;
}

function clearReference(index: number) {
  rows.value[index]!.ref = null;
}

function referenceInfo(row: VariableRow) {
  const reference = row.ref ? referenceById.value.get(row.ref) : undefined;
  if (reference) {
    return { typename: reference.typename, label: reference.sourceName || SOURCE_LABELS[reference.typename], key: reference.key };
  }
  return { typename: null, label: 'Reference', key: row.ref?.split('/').pop() ?? '' };
}
</script>

<template>
  <div v-if="loading && rows.length === 0" class="space-y-4 sm:space-y-2">
    <div
      v-for="row in skeletonRows"
      :key="row.key"
      class="flex flex-wrap items-center gap-x-2 gap-y-1.5"
    >
      <div class="flex h-9 w-full shrink-0 items-center rounded-md border border-input bg-background px-3 sm:w-2/5">
        <Skeleton :class="['h-3', row.key]" />
      </div>
      <div class="flex min-w-0 flex-1">
        <span v-if="linkable" class="size-9 shrink-0 border border-r-0 border-input bg-background" />
        <div
          class="flex h-9 min-w-0 flex-1 items-center rounded-md border border-input bg-background px-3"
          :class="{ 'rounded-l-none': linkable }"
        >
          <Skeleton :class="['h-3', row.value]" />
        </div>
      </div>
      <span class="flex size-9 shrink-0 items-center justify-center">
        <Skeleton class="size-4 rounded-sm" />
      </span>
    </div>
    <div class="flex justify-end pt-2">
      <Skeleton class="h-9 w-32 rounded-none" />
    </div>
  </div>

  <fieldset v-else ref="root" :disabled="disabled" class="min-w-0 space-y-4 sm:space-y-2">
    <div
      v-for="(row, index) in rows"
      :key="index"
      class="flex flex-wrap items-center gap-x-2 gap-y-1.5"
    >
      <Input
        v-model="row.key"
        placeholder="KEY"
        aria-label="Key"
        :readonly="row.removed"
        :aria-invalid="isDuplicate(row) || undefined"
        class="w-full shrink-0 font-mono text-sm sm:w-2/5"
        :class="{
          'text-muted-foreground line-through': row.removed,
          'border-destructive': isDuplicate(row),
        }"
        @paste="onPasteKey($event, index)"
      />

      <div
        class="flex min-w-0 flex-1"
        :class="{ 'opacity-50': row.removed }"
        :inert="row.removed || undefined"
      >
        <template v-if="linkable">
          <Tooltip v-if="!row.ref">
            <TooltipTrigger as-child>
              <span class="inline-flex shrink-0">
                <Popover
                  :open="openPopoverIndex === index"
                  @update:open="(open: boolean) => openPopoverIndex = open ? index : null"
                >
                  <PopoverTrigger as-child>
                    <Button
                      variant="outline"
                      size="icon"
                      class="rounded-r-none border-r-0"
                      :disabled="!references?.length"
                      aria-label="Link to a dynamic value"
                    >
                      <Link class="opacity-50" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent class="w-80 p-0" align="start">
                    <Command>
                      <CommandInput placeholder="Search references…" />
                      <CommandList>
                        <CommandEmpty>No references found.</CommandEmpty>
                        <CommandGroup v-for="group in referenceGroups" :key="group.key">
                          <div class="flex items-center gap-1.5 px-2 py-1.5 text-xs font-medium text-muted-foreground">
                            <SourceIcon :typename="group.typename" />
                            {{ group.label }}
                          </div>
                          <CommandItem
                            v-for="reference in group.items"
                            :key="reference.id"
                            :value="reference.id"
                            class="font-mono text-xs"
                            @select="selectReference(index, reference)"
                          >
                            {{ reference.key }}
                          </CommandItem>
                        </CommandGroup>
                      </CommandList>
                    </Command>
                  </PopoverContent>
                </Popover>
              </span>
            </TooltipTrigger>
            <TooltipContent :class="darkTooltipClass">Link to a dynamic value</TooltipContent>
          </Tooltip>

          <IconButton
            v-else
            label="Detach from dynamic value"
            variant="outline"
            size="icon"
            class="shrink-0 rounded-r-none border-r-0"
            @click="clearReference(index)"
          >
            <Unlink class="opacity-50" />
          </IconButton>
        </template>

        <div
          v-if="row.ref"
          class="magic-border flex h-9 min-w-0 flex-1 items-center gap-1.5 rounded-md rounded-l-none px-3"
        >
          <SourceIcon :typename="referenceInfo(row).typename" />
          <span class="truncate text-xs font-medium">{{ referenceInfo(row).label }}</span>
          <span class="text-muted-foreground/40">·</span>
          <span class="truncate font-mono text-xs text-muted-foreground">{{ referenceInfo(row).key }}</span>
        </div>

        <div v-else class="relative min-w-0 flex-1">
          <Input
            v-model="row.value"
            :type="row.revealed ? 'text' : 'password'"
            placeholder="value"
            :aria-label="`Value of ${row.key || 'new variable'}`"
            autocomplete="off"
            data-1p-ignore="true"
            class="w-full pr-9 font-mono text-sm"
            :class="{ 'rounded-l-none': linkable }"
          />
          <IconButton
            :label="row.revealed ? 'Hide value' : 'Show value'"
            :aria-pressed="row.revealed"
            size="icon"
            tabindex="-1"
            class="absolute right-0 top-0 text-muted-foreground hover:bg-transparent"
            @click="row.revealed = !row.revealed"
          >
            <EyeOff v-if="row.revealed" />
            <Eye v-else />
          </IconButton>
        </div>
      </div>

      <IconButton
        v-if="row.removed"
        :label="`Keep &quot;${row.key}&quot;`"
        size="icon"
        class="shrink-0 text-muted-foreground hover:text-foreground"
        @click="restoreRow(index)"
      >
        <Undo2 />
      </IconButton>
      <IconButton
        v-else
        :label="`Delete ${row.key || 'variable'}`"
        size="icon"
        class="shrink-0 text-muted-foreground hover:text-destructive"
        @click="removeRow(index)"
      >
        <Trash2 />
      </IconButton>

      <p v-if="repeated[index]" class="basis-full text-xs text-destructive">
        "{{ row.key.trim() }}" is already defined
      </p>
    </div>

    <EmptyState
      v-if="rows.length === 0"
      size="compact"
      :icon="Braces"
      :title="emptyTitle"
      :description="emptyDescription"
    />

    <div class="flex items-center justify-end gap-2 pt-2">
      <Button
        v-if="valueRows.length > 0"
        variant="ghost"
        class="text-muted-foreground"
        @click="toggleAll"
      >
        <EyeOff v-if="allRevealed" />
        <Eye v-else />
        {{ allRevealed ? 'Hide all' : 'Show all' }}
      </Button>
      <Button variant="outline" @click="addRow">
        <Plus />
        Add variable
      </Button>
    </div>
  </fieldset>
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
