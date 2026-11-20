<script setup lang="ts">
import { projectRoute } from '@/lib/routes';
import FormFooter from '@/components/FormFooter.vue';
import { useUnsavedChanges } from '@/composables/useUnsavedChanges';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import BlockingResources from '@/components/BlockingResources.vue';
import { ref, watch, computed } from 'vue';
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
import { failureToast, successToast } from '@/components/ui/sonner';
import { useAuth } from '@/composables/useAuth';

const props = defineProps<{
  workspace: {
    id: string;
    name: string;
    personal: boolean;
  };
  projects: { id: string; name: string }[] | null;
  isAdmin: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'changed'): void;
}>();

const router = useRouter();
const { user, refreshToken, setActiveWorkspace } = useAuth();

const editName = ref(props.workspace.name);
const nameChanged = computed(() => editName.value.trim() !== props.workspace.name);

useUnsavedChanges(nameChanged);

watch(() => props.workspace.name, (name) => {
  editName.value = name;
});

const { mutate: updateMutate, loading: updating } = useMutation(UpdateWorkspaceDocument);

async function handleUpdateName() {
  const name = editName.value.trim();
  if (!name || name === props.workspace.name) return;
  try {
    await updateMutate({ input: { name } });
    emit('changed');
  } catch (e: unknown) {
    failureToast('Failed to update workspace', e);
  }
}

const { mutate: deleteMutate, loading: deleting } = useMutation(DeleteWorkspaceDocument, {
  refetchQueries: () => [{ query: WorkspacesDocument }],
});

const deleteOpen = ref(false);

const projectRows = computed(() =>
  (props.projects ?? []).map(project => ({
    id: project.id,
    name: project.name,
    to: projectRoute(project.id),
  })),
);

async function handleDelete() {
  const { name } = props.workspace;

  try {
    await deleteMutate();
    await refreshToken();
    const firstWorkspace = user.value?.workspaces[0]?.workspace;
    if (firstWorkspace) {
      setActiveWorkspace(firstWorkspace);
    }
    apolloClient.resetStore();
    router.push({ name: 'home' });
    successToast(`Workspace "${name}" deleted`);
  } catch (e: unknown) {
    failureToast('Failed to delete workspace', e);
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
        <h3 class="text-sm font-medium text-muted-foreground">General</h3>
        <div class="overflow-hidden rounded-lg border">
          <fieldset :disabled="updating" class="min-w-0">
            <div class="space-y-3.5 p-3">
              <form class="space-y-1.5" @submit.prevent="nameChanged && handleUpdateName()">
                <Label for="workspace-name" class="text-xs text-muted-foreground">Name</Label>
                <Input
                  id="workspace-name"
                  v-model="editName"
                  class="h-9 min-w-0"
                  :disabled="!isAdmin || updating"
                />
              </form>
              <div class="flex items-center justify-between gap-4">
                <span class="text-xs text-muted-foreground">ID</span>
                <span class="flex items-center gap-2">
                  <span v-if="workspace.personal" class="rounded-full bg-muted px-2 py-px text-xs text-muted-foreground">Personal</span>
                  <span class="truncate font-mono text-xs text-foreground">{{ workspace.id }}</span>
                </span>
              </div>
            </div>
          </fieldset>
          <FormFooter
            v-if="isAdmin"
            :dirty="nameChanged"
            :saving="updating"
            :disabled="!editName.trim()"
            @save="handleUpdateName"
            @discard="editName = workspace.name"
          />
        </div>
      </section>

      <div
        v-if="isAdmin && !workspace.personal"
        class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3"
      >
        <div>
          <p class="text-sm font-medium text-foreground">Delete workspace</p>
          <p class="text-xs text-muted-foreground">
            {{ projectRows.length ? 'Delete its projects first.' : 'Removes the workspace and its members.' }}
          </p>
        </div>
        <ConfirmDialog
          v-model:open="deleteOpen"
          :title="projectRows.length ? `Workspace &quot;${workspace.name}&quot; still has projects` : `Delete workspace &quot;${workspace.name}&quot;?`"
          :description="projectRows.length ? 'Delete these first, then the workspace.' : 'Every member loses access. This can\'t be undone.'"
          confirm-label="Delete workspace"
          pending-label="Deleting…"
          :pending="deleting"
          :blocked="projectRows.length > 0"
          destructive
          @confirm="handleDelete"
        >
          <BlockingResources v-if="projectRows.length" :items="projectRows" @select="deleteOpen = false" />
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
