<script setup lang="ts">
import { ref, computed } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { CreateEnvironmentMutation, ProjectQuery } from '@/graphql/projects';
import { useEnvironment } from '@/composables/useEnvironment';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
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
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  projectId: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const { environments } = useEnvironment();

const name = ref('');
const fromEnvironment = ref<string>('');

const { mutate, loading } = useMutation(CreateEnvironmentMutation, {
  refetchQueries: () => [{ query: ProjectQuery, variables: { id: props.projectId } }],
});

const nonEphemeralEnvs = computed(() =>
  environments.value.filter(e => !e.ephemeral),
);

async function handleCreate() {
  if (!name.value.trim()) return;

  try {
    const input: Record<string, string> = {
      projectId: props.projectId,
      name: name.value.trim(),
    };
    if (fromEnvironment.value) {
      input.fromEnvironment = fromEnvironment.value;
    }

    const res = await mutate({ input });

    if (res?.errors?.length) {
      toast.error('Failed to create environment', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success(`Environment "${name.value.trim()}" created`);
    name.value = '';
    fromEnvironment.value = '';
    emit('update:open', false);
  } catch (e: unknown) {
    toast.error('Failed to create environment', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>New Environment</DialogTitle>
        <DialogDescription>
          Create a new environment for this project.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="handleCreate">
        <div class="space-y-2">
          <Label for="env-name">Name</Label>
          <Input
            id="env-name"
            v-model="name"
            placeholder="e.g. staging, preview"
            :disabled="loading"
          />
        </div>

        <div v-if="nonEphemeralEnvs.length > 0" class="space-y-2">
          <Label for="env-from">Clone from</Label>
          <Select v-model="fromEnvironment">
            <SelectTrigger id="env-from">
              <SelectValue placeholder="Start empty" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="env in nonEphemeralEnvs"
                :key="env.id"
                :value="env.name"
              >
                {{ env.name }}
              </SelectItem>
            </SelectContent>
          </Select>
          <p class="text-xs text-muted-foreground">
            Copy configuration and image tags from an existing environment.
          </p>
        </div>

        <DialogFooter>
          <Button
            type="submit"
            :disabled="!name.trim() || loading"
          >
            {{ loading ? 'Creating...' : 'Create' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
