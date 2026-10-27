<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';

const WorkspacesDocument = graphql(`
  query Workspaces {
    workspaces {
      id
      name
      personal
    }
  }
`);

const UpdateWorkspaceDocument = graphql(`
  mutation UpdateWorkspace($input: UpdateWorkspaceInput!) {
    updateWorkspace(input: $input) {
      id
      name
    }
  }
`);

const DeleteWorkspaceDocument = graphql(`
  mutation DeleteWorkspace {
    deleteWorkspace
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';
import { useAuth } from '@/composables/useAuth';

const props = defineProps<{
  workspace: {
    id: string;
    name: string;
    personal: boolean;
  };
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'changed'): void;
}>();

const router = useRouter();
const { user, refreshToken, setActiveWorkspace } = useAuth();

const editName = ref(props.workspace.name);

watch(() => props.workspace.name, (name) => {
  editName.value = name;
});

const { mutate: updateMutate, loading: updating } = useMutation(UpdateWorkspaceDocument);

async function handleUpdateName() {
  const name = editName.value.trim();
  if (!name || name === props.workspace.name) return;
  try {
    await updateMutate({ input: { name } });
    toast.success('Workspace name updated');
    emit('changed');
  } catch (e: unknown) {
    errorToast('Failed to update workspace', { description: errorMessage(e) });
  }
}

const { mutate: deleteMutate, loading: deleting } = useMutation(DeleteWorkspaceDocument, {
  refetchQueries: () => [{ query: WorkspacesDocument }],
});

async function handleDelete() {
  try {
    const res = await deleteMutate();
    if (res?.errors?.length) {
      errorToast('Failed to delete workspace', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }
    await refreshToken();
    const firstWorkspace = user.value?.workspaces[0]?.workspace;
    if (firstWorkspace) {
      setActiveWorkspace(firstWorkspace);
    }
    apolloClient.resetStore();
    toast.success('Workspace deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    errorToast('Failed to delete workspace', { description: errorMessage(e) });
  }
}
</script>

<template>
  <ContextDrawer
    title="Workspace settings"
    @close="emit('close')"
  >
    <div class="flex min-h-full flex-col gap-6">
      <section class="space-y-2.5">
        <h3 class="text-[13px] font-medium text-muted-foreground">General</h3>
        <div class="space-y-3.5 rounded-lg border p-3">
          <form class="space-y-1.5" @submit.prevent="handleUpdateName">
            <Label for="workspace-name" class="text-xs text-muted-foreground">Name</Label>
            <div class="flex gap-2">
              <Input
                id="workspace-name"
                v-model="editName"
                class="h-9 min-w-0 flex-1 text-[13px]"
                :disabled="!isAdmin || updating"
              />
              <Button
                v-if="isAdmin"
                type="submit"
                class="shrink-0"
                :disabled="!editName.trim() || editName.trim() === workspace.name || updating"
              >
                {{ updating ? 'Saving...' : 'Save' }}
              </Button>
            </div>
          </form>
          <div class="flex items-center justify-between gap-4">
            <span class="text-xs text-muted-foreground">ID</span>
            <span class="flex items-center gap-2">
              <span v-if="workspace.personal" class="rounded-full bg-muted px-2 py-px text-[11px] text-muted-foreground">Personal</span>
              <span class="truncate font-mono text-xs text-foreground">{{ workspace.id }}</span>
            </span>
          </div>
        </div>
      </section>

      <div
        v-if="isAdmin && !workspace.personal"
        class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3"
      >
        <div>
          <p class="text-[13px] font-medium text-foreground">Delete workspace</p>
          <p class="text-xs text-muted-foreground">
            Removes the workspace and its members. All projects have to be deleted first.
          </p>
        </div>
        <ConfirmDialog
          :title="`Delete workspace &quot;${workspace.name}&quot;?`"
          description="Every member loses access. Its projects have to be deleted first. This can't be undone."
          confirm-label="Delete workspace"
          pending-label="Deleting…"
          :pending="deleting"
          destructive
          @confirm="handleDelete"
        >
          <template #trigger>
            <Button variant="destructive" class="shrink-0" :disabled="deleting">
              <Trash2 />
              Delete
            </Button>
          </template>
        </ConfirmDialog>
      </div>
    </div>
  </ContextDrawer>
</template>
