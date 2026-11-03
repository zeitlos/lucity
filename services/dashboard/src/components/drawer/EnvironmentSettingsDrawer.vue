<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref } from 'vue';
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

const props = defineProps<{
  projectId: string;
  environment: {
    id: string;
    name: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();

const { mutate: deleteEnvironmentMutate, loading: deleting } = useMutation(DeleteEnvironmentDocument);
const confirmOpen = ref(false);

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
          <p class="text-xs text-muted-foreground">
            It has to be empty, and the last environment of a project can't be deleted.
          </p>
        </div>
        <ConfirmDialog
          v-model:open="confirmOpen"
          :title="`Delete environment &quot;${environment.name}&quot;?`"
          description="Its deployment history goes with it. This can't be undone."
          confirm-label="Delete environment"
          pending-label="Deleting…"
          :pending="deleting"
          destructive
          @confirm="handleDelete"
        >
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
