<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { CreateWorkspaceMutation, WorkspacesQuery } from '@/graphql/workspaces';
import { useAuth } from '@/composables/useAuth';
import { apolloClient } from '@/lib/apollo';
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
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const router = useRouter();
const { setActiveWorkspace, refreshToken } = useAuth();

const name = ref('');

const id = computed(() =>
  name.value
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 63),
);

const isValid = computed(() => id.value.length >= 3 && name.value.trim().length > 0);

const { mutate, loading } = useMutation(CreateWorkspaceMutation, {
  refetchQueries: () => [{ query: WorkspacesQuery }],
});

async function handleCreate() {
  if (!isValid.value) return;

  try {
    const res = await mutate({
      input: { id: id.value, name: name.value.trim() },
    });

    if (res?.errors?.length) {
      toast.error('Failed to create workspace', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    // Refresh JWT to include new workspace membership
    await refreshToken();

    // Switch to the new workspace
    setActiveWorkspace(id.value);
    apolloClient.resetStore();

    toast.success(`Workspace "${name.value.trim()}" created`);
    name.value = '';
    emit('update:open', false);
    router.push('/');
  } catch (e: unknown) {
    toast.error('Failed to create workspace', { description: errorMessage(e) });
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
        <DialogTitle>New Workspace</DialogTitle>
        <DialogDescription>
          Create a workspace to collaborate with your team.
        </DialogDescription>
      </DialogHeader>

      <form class="space-y-4" @submit.prevent="handleCreate">
        <div class="space-y-2">
          <Label for="ws-name">Name</Label>
          <Input
            id="ws-name"
            v-model="name"
            placeholder="e.g. My Team"
            :disabled="loading"
          />
        </div>

        <div class="space-y-2">
          <Label for="ws-id">ID</Label>
          <Input
            id="ws-id"
            :model-value="id"
            disabled
            class="font-mono text-sm"
          />
          <p class="text-xs text-muted-foreground">
            Auto-generated from the name. Used in URLs and API calls.
          </p>
        </div>

        <DialogFooter>
          <Button
            type="submit"
            :disabled="!isValid || loading"
          >
            {{ loading ? 'Creating...' : 'Create Workspace' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
