<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2, Link } from 'lucide-vue-next';
import { ServiceVariablesQuery, SetServiceVariablesMutation, SharedVariablesQuery } from '@/graphql/variables';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
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

// Fetch service variables
const { result, loading, refetch } = useQuery(ServiceVariablesQuery, () => ({
  projectId: props.projectId,
  environment: envName.value,
  service: props.service.name,
}), () => ({
  enabled: !!envName.value,
}));

// Fetch shared variables for reference dropdown
const { result: sharedResult } = useQuery(SharedVariablesQuery, () => ({
  projectId: props.projectId,
  environment: envName.value,
}), () => ({
  enabled: !!envName.value,
}));

const availableSharedKeys = computed(() => {
  const vars = sharedResult.value?.sharedVariables ?? [];
  return vars.map((v: { key: string }) => v.key);
});

interface VarRow {
  key: string;
  value: string;
  fromShared: boolean;
  isNew?: boolean;
}

const rows = ref<VarRow[]>([]);
const hasChanges = ref(false);

watch(
  () => result.value?.serviceVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map((v: { key: string; value: string; fromShared: boolean }) => ({
        key: v.key,
        value: v.value,
        fromShared: v.fromShared,
      }));
      hasChanges.value = false;
    }
  },
  { immediate: true },
);

function addRow() {
  rows.value.push({ key: '', value: '', fromShared: false, isNew: true });
  hasChanges.value = true;
}

function addSharedRef() {
  rows.value.push({ key: '', value: '', fromShared: true, isNew: true });
  hasChanges.value = true;
}

function removeRow(index: number) {
  rows.value.splice(index, 1);
  hasChanges.value = true;
}

function markChanged() {
  hasChanges.value = true;
}

function handleSharedKeySelect(index: number, key: string) {
  rows.value[index].key = key;
  // Resolve value from shared vars for display
  const shared = sharedResult.value?.sharedVariables?.find(
    (v: { key: string }) => v.key === key
  );
  if (shared) {
    rows.value[index].value = shared.value;
  }
  hasChanges.value = true;
}

const { mutate: setVarsMutate, loading: saving } = useMutation(SetServiceVariablesMutation);

async function handleSave() {
  const validRows = rows.value.filter(r => r.key.trim());
  try {
    const variables = validRows.map(r => ({
      key: r.key.trim(),
      value: r.fromShared ? undefined : r.value,
      fromShared: r.fromShared || undefined,
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
        <!-- Shared reference row -->
        <template v-if="row.fromShared">
          <div class="flex flex-1 items-center gap-2">
            <Select
              :model-value="row.key"
              @update:model-value="(val: string) => handleSharedKeySelect(index, val)"
            >
              <SelectTrigger class="flex-1 font-mono text-sm">
                <SelectValue placeholder="Select shared variable" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="key in availableSharedKeys"
                  :key="key"
                  :value="key"
                >
                  {{ key }}
                </SelectItem>
              </SelectContent>
            </Select>
            <Badge variant="secondary" class="shrink-0 gap-1 text-xs">
              <Link :size="10" />
              shared
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
          <Button
            v-if="availableSharedKeys.length > 0"
            variant="outline"
            size="sm"
            @click="addSharedRef"
          >
            <Link :size="14" class="mr-1" />
            Reference Shared
          </Button>
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
