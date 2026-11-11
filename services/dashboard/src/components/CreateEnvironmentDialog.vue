<script setup lang="ts">
import { mobileSheetClass } from '@/lib/dialog';
import DialogIcon from '@/components/DialogIcon.vue';
import { Layers } from '@lucide/vue';
import { ref, computed } from 'vue';
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

const { mutate, loading } = useMutation(CreateEnvironmentDocument);

const availableEnvs = computed(() => environments.value);
const takenIds = computed(() => new Set(availableEnvs.value.map(environment => lastSegment(environment.id))));
const isValid = computed(() => name.value.trim().length > 0 && isValidId(id.value) && !takenIds.value.has(id.value));

async function handleCreate() {
  if (!isValid.value) return;

  try {
    const input: CreateEnvironmentInput = {
      project: props.projectId,
      name: name.value.trim(),
      id: id.value,
    };

    const res = await mutate({ input });

    const newId = res?.data?.createEnvironment?.id;
    name.value = '';
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
