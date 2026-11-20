<script setup lang="ts">
import { projectRoute } from '@/lib/routes';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import BlockingResources from '@/components/BlockingResources.vue';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';

const RenameEnvironmentDocument = graphql(`
  mutation RenameEnvironment($environment: EnvironmentID!, $name: String!) {
    renameEnvironment(environment: $environment, name: $name) {
      id
      name
    }
  }
`);

const DeleteEnvironmentDocument = graphql(`
  mutation DeleteEnvironment($environment: EnvironmentID!) {
    deleteEnvironment(environment: $environment)
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import NameSettings from '@/components/NameSettings.vue';
import { Button } from '@/components/ui/button';
import { failureToast, successToast } from '@/components/ui/sonner';
import { usePanel } from '@/composables/usePanel';
import { describeResources, environmentResources, type EnvironmentContents, type ResourceRow } from '@/lib/resources';

const props = defineProps<{
  projectId: string;
  environment: EnvironmentContents & {
    id: string;
    name: string;
  };
  only: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();

const { mutate: renameMutate, loading: renaming } = useMutation(RenameEnvironmentDocument);

async function handleRename(name: string) {
  try {
    await renameMutate({ environment: props.environment.id, name });
  } catch (e: unknown) {
    failureToast('Failed to rename environment', e);
  }
}

const { mutate: deleteEnvironmentMutate, loading: deleting } = useMutation(DeleteEnvironmentDocument);
const confirmOpen = ref(false);

const resources = computed(() => environmentResources(props.environment));

const hint = computed(() => {
  if (props.only) return "It's the only environment in this project, which needs at least one.";
  if (resources.value.length) return `It still holds ${describeResources(props.environment)}.`;
  return 'Permanently delete this environment and its deployment history.';
});

const dialog = computed(() => {
  const name = props.environment.name;

  if (resources.value.length) {
    return {
      title: `Environment "${name}" isn't empty`,
      description: 'Delete these first, then the environment.',
    };
  }

  return {
    title: `Delete environment "${name}"?`,
    description: "Its deployment history goes with it. This can't be undone.",
  };
});

const { openPanel } = usePanel();

function openResource(resource: ResourceRow) {
  const type = resource.kind === 'repositoryService' || resource.kind === 'imageService' ? 'service' : resource.kind;
  confirmOpen.value = false;
  emit('close');
  openPanel({ type, id: resource.id, label: resource.name, tab: 'settings' });
}

async function handleDelete() {
  const { id, name } = props.environment;

  try {
    await deleteEnvironmentMutate({ environment: id });

    apolloClient.cache.evict({ id: apolloClient.cache.identify({ __typename: 'Environment', id }) });
    apolloClient.cache.gc();

    confirmOpen.value = false;
    router.push(projectRoute(props.projectId));
    successToast(`Environment "${name}" deleted`);
  } catch (e: unknown) {
    failureToast('Failed to delete environment', e);
  }
}
</script>

<template>
  <ContextDrawer
    title="Environment settings"
    @close="emit('close')"
  >
    <div class="flex min-h-full flex-col gap-6">
      <section class="space-y-2.5">
        <h3 class="text-sm font-medium text-muted-foreground">General</h3>
        <NameSettings
          :name="environment.name"
          :id="environment.id"
          :saving="renaming"
          @save="handleRename"
        />
      </section>

      <div class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-sm font-medium text-foreground">Delete environment</p>
          <p class="text-xs text-muted-foreground">{{ hint }}</p>
        </div>
        <ConfirmDialog
          v-model:open="confirmOpen"
          :title="dialog.title"
          :description="dialog.description"
          confirm-label="Delete environment"
          pending-label="Deleting…"
          :pending="deleting"
          :blocked="resources.length > 0"
          destructive
          @confirm="handleDelete"
        >
          <BlockingResources v-if="resources.length" :items="resources" @select="openResource" />
          <template #trigger>
            <Button variant="destructive" class="shrink-0" :disabled="only">
              <Trash2 />
              Delete
            </Button>
          </template>
        </ConfirmDialog>
      </div>
    </div>
  </ContextDrawer>
</template>
