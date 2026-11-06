<script setup lang="ts">
import { mobileSheetClass } from '@/lib/dialog';
import DialogIcon from '@/components/DialogIcon.vue';
import { Layers } from '@lucide/vue';
import { ref, computed, watch } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import type { CreateEnvironmentInput } from '@/gql/graphql';

const CreateEnvironmentDocument = graphql(`
  mutation CreateEnvironment($input: CreateEnvironmentInput!) {
    createEnvironment(input: $input) {
      id
      name
    }
  }
`);
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import NameIdField from '@/components/NameIdField.vue';
import { isValidId, lastSegment } from '@/lib/id';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { failureToast } from '@/components/ui/sonner';

const props = defineProps<{
  open: boolean;
  projectId: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'created', environmentId: string): void;
}>();

const { environments } = useEnvironment();

const name = ref('');
const id = ref('');
const mode = ref<'duplicate' | 'empty'>('duplicate');
const fromEnvironmentId = ref<string>('');

const { mutate, loading } = useMutation(CreateEnvironmentDocument);

const availableEnvs = computed(() => environments.value);
const takenIds = computed(() => new Set(availableEnvs.value.map(environment => lastSegment(environment.id))));
const isValid = computed(() => name.value.trim().length > 0 && isValidId(id.value) && !takenIds.value.has(id.value));

// Default to 'duplicate' when environments exist, 'empty' when none
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    if (availableEnvs.value.length > 0) {
      mode.value = 'duplicate';
      fromEnvironmentId.value = availableEnvs.value[0]!.id;
    } else {
      mode.value = 'empty';
    }
  }
});

async function handleCreate() {
  if (!isValid.value) return;

  try {
    const input: CreateEnvironmentInput = {
      project: props.projectId,
      name: name.value.trim(),
      id: id.value,
    };
    if (mode.value === 'duplicate' && fromEnvironmentId.value) {
      input.fromEnvironment = fromEnvironmentId.value;
    }

    const res = await mutate({ input });

    const newId = res?.data?.createEnvironment?.id;
    name.value = '';
    fromEnvironmentId.value = '';
    emit('update:open', false);
    if (newId) emit('created', newId);
  } catch (e: unknown) {
    failureToast('Failed to create environment', e);
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent :class="[mobileSheetClass, 'sm:max-w-md']">
      <DialogHeader>
        <DialogIcon :icon="Layers" />
        <DialogTitle>New environment</DialogTitle>
        <DialogDescription>
          All changes will be isolated from other environments.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="handleCreate">
        <NameIdField
          v-model:name="name"
          v-model:id="id"
          fallback="environment"
          :taken="takenIds"
          scope="project"
          :disabled="loading"
          name-placeholder="e.g. Staging"
        />

        <div
          v-if="availableEnvs.length > 0"
          class="space-y-2"
        >
          <RadioGroup v-model="mode" class="space-y-3">
            <label
              class="flex cursor-pointer flex-col gap-2 rounded-lg border p-3 transition-colors"
              :class="mode === 'duplicate' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="duplicate" />
                <span class="text-sm font-medium">Duplicate environment</span>
              </div>
              <p class="text-xs text-muted-foreground pl-6">
                Copy all the services, variables, and configuration from an existing environment.
              </p>
              <div v-if="mode === 'duplicate'" class="pl-6 pt-1">
                <Select v-model="fromEnvironmentId">
                  <SelectTrigger>
                    <SelectValue placeholder="Select environment" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem
                      v-for="env in availableEnvs"
                      :key="env.id"
                      :value="env.id"
                    >
                      {{ env.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </label>
            <label
              class="flex cursor-pointer flex-col gap-1 rounded-lg border p-3 transition-colors"
              :class="mode === 'empty' ? 'border-primary bg-primary/5' : 'border-border'"
            >
              <div class="flex items-center gap-2">
                <RadioGroupItem value="empty" />
                <span class="text-sm font-medium">Empty environment</span>
              </div>
              <p class="text-xs text-muted-foreground pl-6">
                An empty environment with no services or variables included.
              </p>
            </label>
          </RadioGroup>
        </div>

        <DialogFooter>
          <Button
            type="submit"
            :pending="loading"
            :disabled="!isValid"
          >
            {{ loading ? 'Creating…' : 'Create environment' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
