<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Eye, EyeOff } from '@lucide/vue';
import { graphql } from '@/gql';

const SharedVariablesDocument = graphql(`
  query SharedVariables($environment: EnvironmentID!) {
    sharedVariables(environment: $environment) {
      key
      value
    }
  }
`);

const SetSharedVariablesDocument = graphql(`
  mutation SetSharedVariables($environment: EnvironmentID!, $variables: [VariableInput!]!) {
    setSharedVariables(environment: $environment, variables: $variables)
  }
`);
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import IconButton from '@/components/IconButton.vue';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import { parseEnvAssignments } from '@/lib/env';

const props = defineProps<{
  environmentId: string;
  environmentName: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { result, loading, refetch } = useQuery(SharedVariablesDocument, () => ({
  environment: props.environmentId,
}));

interface VarRow {
  key: string;
  value: string;
  revealed?: boolean;
}

const rows = ref<VarRow[]>([]);
const hasChanges = ref(false);

watch(
  () => result.value?.sharedVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map(v => ({ key: v.key, value: v.value }));
      hasChanges.value = false;
    }
  },
  { immediate: true },
);

function addRow() {
  rows.value.push({ key: '', value: '', revealed: true });
  hasChanges.value = true;
}

const allRevealed = computed(
  () => rows.value.length > 0 && rows.value.every(r => r.revealed),
);

function toggleAll() {
  const next = !allRevealed.value;
  for (const row of rows.value) row.revealed = next;
}

function onPaste(event: ClipboardEvent, index: number) {
  const parsed = parseEnvAssignments(event.clipboardData?.getData('text') ?? '');

  if (parsed.length === 0) return;

  event.preventDefault();
  rows.value.splice(index, 1, ...parsed.map(p => ({ key: p.key, value: p.value })));
  hasChanges.value = true;
}

function removeRow(index: number) {
  rows.value.splice(index, 1);
  hasChanges.value = true;
}

function markChanged() {
  hasChanges.value = true;
}

const { mutate: setVarsMutate, loading: saving } = useMutation(SetSharedVariablesDocument);

async function handleSave() {
  const validRows = rows.value.filter(r => r.key.trim());
  try {
    const res = await setVarsMutate({
      environment: props.environmentId,
      variables: validRows.map(r => ({ key: r.key.trim(), value: r.value })),
    });

    if (res?.errors?.length) {
      errorToast('Failed to save variables', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Shared variables saved');
    hasChanges.value = false;
    refetch();
  } catch (e: unknown) {
    errorToast('Failed to save variables', { description: errorMessage(e) });
  }
}
</script>

<template>
  <ContextDrawer
    title="Shared variables"
    @close="emit('close')"
  >
    <div v-if="loading && rows.length === 0" class="space-y-2">
      <Skeleton class="h-9 w-full" />
      <Skeleton class="h-9 w-full" />
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="(row, index) in rows"
        :key="index"
        class="flex items-center gap-2"
      >
        <Input
          v-model="row.key"
          placeholder="KEY"
          aria-label="Key"
          class="h-9 w-[150px] shrink-0 font-mono text-xs uppercase"
          @input="markChanged"
          @paste="onPaste($event, index)"
        />
        <div class="relative min-w-0 flex-1">
          <Input
            v-model="row.value"
            :type="row.revealed ? 'text' : 'password'"
            placeholder="value"
            :aria-label="`Value of ${row.key || 'new variable'}`"
            autocomplete="off"
            data-1p-ignore="true"
            class="h-9 w-full pr-9 font-mono text-xs"
            @input="markChanged"
          />
          <IconButton
            :label="row.revealed ? 'Hide value' : 'Show value'"
            size="icon"
            tabindex="-1"
            class="absolute right-0 top-0 text-muted-foreground hover:bg-transparent"
            @click="row.revealed = !row.revealed"
          >
            <EyeOff v-if="row.revealed" />
            <Eye v-else />
          </IconButton>
        </div>
        <IconButton
          :label="`Remove ${row.key || 'variable'}`"
          class="shrink-0 text-muted-foreground hover:text-destructive"
          @click="removeRow(index)"
        >
          <Trash2 />
        </IconButton>
      </div>

      <div v-if="rows.length === 0" class="rounded-lg border border-dashed p-6 text-center">
        <p class="text-sm text-muted-foreground">No shared variables yet.</p>
        <p class="mt-1 text-xs text-muted-foreground">Services in {{ environmentName }} can reference them.</p>
      </div>

      <div class="flex items-center gap-2 pt-1.5">
        <Button variant="outline" @click="addRow">
          <Plus />
          Add variable
        </Button>
        <Button
          v-if="rows.length > 0"
          variant="ghost"
          class="text-muted-foreground"
          @click="toggleAll"
        >
          <EyeOff v-if="allRevealed" />
          <Eye v-else />
          {{ allRevealed ? 'Hide all' : 'Show all' }}
        </Button>
      </div>
    </div>

    <template #footer>
      <div class="flex items-center justify-between gap-4">
        <p class="text-xs leading-[18px] text-muted-foreground">
          Paste <code class="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">KEY=value</code> lines into a key field to add several.
        </p>
        <Button
          class="shrink-0"
          :disabled="!hasChanges || saving"
          @click="handleSave"
        >
          {{ saving ? 'Saving...' : 'Save' }}
        </Button>
      </div>
    </template>
  </ContextDrawer>
</template>
