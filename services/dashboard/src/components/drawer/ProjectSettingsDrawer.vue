<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from '@lucide/vue';
import { graphql } from '@/gql';

const DeleteProjectDocument = graphql(`
  mutation DeleteProject($id: ProjectID!) {
    deleteProject(id: $id)
  }
`);
import { apolloClient } from '@/lib/apollo';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import { Button } from '@/components/ui/button';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  project: {
    id: string;
    name: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const router = useRouter();

const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectDocument);

async function handleDelete() {
  try {
    const res = await deleteProjectMutate({ id: props.project.id });

    if (res?.errors?.length) {
      errorToast('Failed to delete project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    apolloClient.cache.evict({ id: `Project:${props.project.id}` });
    apolloClient.cache.gc();

    toast.success('Project deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    errorToast('Failed to delete project', { description: errorMessage(e) });
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
        <h3 class="text-[13px] font-medium text-muted-foreground">General</h3>
        <div class="space-y-2.5 rounded-lg border p-3">
          <div class="flex items-center justify-between gap-4">
            <span class="text-[13px] text-muted-foreground">Name</span>
            <span class="text-[13px] font-medium text-foreground">{{ project.name }}</span>
          </div>
          <div class="flex items-center justify-between gap-4">
            <span class="text-[13px] text-muted-foreground">ID</span>
            <span class="truncate font-mono text-xs text-foreground">{{ project.id }}</span>
          </div>
        </div>
      </section>

      <div class="mt-auto flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
        <div>
          <p class="text-[13px] font-medium text-foreground">Delete project</p>
          <p class="text-xs text-muted-foreground">Permanently delete this project and all its data.</p>
        </div>
        <ConfirmDialog
          :title="`Delete project &quot;${project.name}&quot;?`"
          description="Every environment, service and resource in it is deleted too. This can't be undone."
          confirm-label="Delete project"
          pending-label="Deleting…"
          :pending="deleting"
          :confirm-text="project.name"
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
