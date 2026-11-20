<script setup lang="ts">
import { environmentRoute } from '@/lib/routes';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import BlockingResources from '@/components/BlockingResources.vue';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';

const RenameProjectDocument = graphql(`
  mutation RenameProject($project: ProjectID!, $name: String!) {
    renameProject(project: $project, name: $name) {
      id
      name
    }
  }
`);

const DeleteProjectDocument = graphql(`
  mutation DeleteProject($id: ProjectID!) {
    deleteProject(id: $id)
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import NameSettings from '@/components/NameSettings.vue';
import { Button } from '@/components/ui/button';
import { failureToast, successToast } from '@/components/ui/sonner';
import { describeResources, environmentResources, type EnvironmentContents } from '@/lib/resources';

const props = defineProps<{
  project: {
    id: string;
    name: string;
    environments: (EnvironmentContents & { id: string; name: string })[];
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();

const { mutate: renameMutate, loading: renaming } = useMutation(RenameProjectDocument);

async function handleRename(name: string) {
  try {
    await renameMutate({ project: props.project.id, name });
  } catch (e: unknown) {
    failureToast('Failed to rename project', e);
  }
}

const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectDocument);
const confirmOpen = ref(false);

const occupied = computed(() =>
  props.project.environments
    .filter(environment => environmentResources(environment).length > 0)
    .map(environment => ({
      id: environment.id,
      name: environment.name,
      detail: describeResources(environment),
      to: environmentRoute(environment.id),
    })),
);

const dialog = computed(() => {
  const name = props.project.name;

  if (occupied.value.length) {
    return {
      title: `Project "${name}" isn't empty`,
      description: 'Delete the resources in these environments first, then the project.',
    };
  }

  return {
    title: `Delete project "${name}"?`,
    description: "Its environments and their deployment history go with it. This can't be undone.",
  };
});

async function handleDelete() {
  try {
    await deleteProjectMutate({ id: props.project.id });

    apolloClient.cache.evict({ id: `Project:${props.project.id}` });
    apolloClient.cache.gc();

    router.push({ name: 'projects' });
    successToast(`Project "${props.project.name}" deleted`);
  } catch (e: unknown) {
    failureToast('Failed to delete project', e);
  }
}
</script>

<template>
  <ContextDrawer
    title="Project settings"
    @close="emit('close')"
  >
    <div class="flex min-h-full flex-col gap-6">
      <section class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">General</h3>
        <NameSettings
          :name="project.name"
          :id="project.id"
          :saving="renaming"
          @save="handleRename"
        />
      </section>

      <div class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete project</p>
          <p class="text-xs text-muted-foreground">
            {{ occupied.length ? 'Delete the resources in its environments first.' : 'Permanently delete this project and its environments.' }}
          </p>
        </div>
        <ConfirmDialog
          v-model:open="confirmOpen"
          :title="dialog.title"
          :description="dialog.description"
          confirm-label="Delete project"
          pending-label="Deleting…"
          :pending="deleting"
          :confirm-text="project.name"
          :blocked="occupied.length > 0"
          destructive
          @confirm="handleDelete"
        >
          <BlockingResources v-if="occupied.length" :items="occupied" @select="confirmOpen = false" />
          <template #trigger>
            <Button variant="destructive" class="shrink-0">
              <Trash2 />
              Delete
            </Button>
          </template>
        </ConfirmDialog>
      </div>
    </div>
  </ContextDrawer>
</template>
