<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import BlockingResources from '@/components/BlockingResources.vue';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';

const DeleteEnvironmentDocument = graphql(`
  mutation DeleteEnvironment($environment: EnvironmentID!) {
    deleteEnvironment(environment: $environment)
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
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

const { mutate: deleteEnvironmentMutate, loading: deleting } = useMutation(DeleteEnvironmentDocument);
const confirmOpen = ref(false);

const resources = computed(() => environmentResources(props.environment));
const blocked = computed(() => props.only || resources.value.length > 0);

const hint = computed(() => {
  if (props.only) return "It's the only environment in this project.";
  if (resources.value.length) return `It still holds ${describeResources(props.environment)}.`;
  return 'Permanently delete this environment and its deployment history.';
});

const dialog = computed(() => {
  const name = props.environment.name;

  if (props.only) {
    return {
      title: `Can't delete environment "${name}"`,
      description: 'A project needs at least one environment. Create another one first, or delete the project.',
    };
  }

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
    router.push({ name: 'project', params: { projectId: props.projectId } });
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
        <div class="space-y-2.5 rounded-lg border p-3">
          <div class="flex items-center justify-between gap-4">
            <span class="text-sm text-muted-foreground">Name</span>
            <span class="text-sm font-medium text-foreground">{{ environment.name }}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="text-sm text-muted-foreground">ID</span>
            <span class="truncate font-mono text-xs text-foreground">{{ environment.id }}</span>
          </div>
        </div>
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
          :blocked="blocked"
          destructive
          @confirm="handleDelete"
        >
          <BlockingResources v-if="!only && resources.length" :items="resources" @select="openResource" />
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
