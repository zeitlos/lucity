<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Plus, Trash2 } from 'lucide-vue-next';
import { SharedVariablesDocument, SetSharedVariablesDocument } from '@/gql/graphql';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
}>();

const { activeEnvironment } = useEnvironment();
const envName = computed(() => activeEnvironment.value?.name ?? '');

const { result, loading, refetch } = useQuery(SharedVariablesDocument, () => ({
  projectId: props.projectId,
  environment: envName.value,
}), () => ({
  enabled: !!envName.value,
}));

interface VarRow {
  key: string;
  value: string;
  isNew?: boolean;
}

const rows = ref<VarRow[]>([]);
const hasChanges = ref(false);

watch(
  () => result.value?.sharedVariables,
  (vars) => {
    if (vars) {
      rows.value = vars.map((v: { key: string; value: string }) => ({
        key: v.key,
        value: v.value,
      }));
      hasChanges.value = false;
    }
  },
  { immediate: true },
);

function addRow() {
  rows.value.push({ key: '', value: '', isNew: true });
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
      projectId: props.projectId,
      environment: envName.value,
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
  <div class="space-y-4">
    <div>
      <h3 class="text-sm font-medium text-foreground">Shared Variables</h3>
      <p class="text-xs text-muted-foreground">
        Variables available for services to reference in {{ envName || 'this environment' }}.
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
        <p class="text-sm text-muted-foreground">No shared variables configured for this environment.</p>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-between pt-2">
        <Button variant="outline" size="sm" @click="addRow">
          <Plus :size="14" class="mr-1" />
          Add Variable
        </Button>

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
