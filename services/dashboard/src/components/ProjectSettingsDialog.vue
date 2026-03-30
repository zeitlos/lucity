<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from 'lucide-vue-next';
import { DeleteProjectDocument } from '@/gql/graphql';
import { apolloClient } from '@/lib/apollo';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  projectId: string;
  projectName: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const router = useRouter();
const { mutate: deleteProjectMutate, loading: deleting } = useMutation(DeleteProjectDocument);

async function handleDeleteProject() {
  try {
    const res = await deleteProjectMutate({ id: props.projectId });

    if (res?.errors?.length) {
      errorToast('Failed to delete project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    apolloClient.cache.evict({ id: `Project:${props.projectId}` });
    apolloClient.cache.gc();

    emit('update:open', false);
    toast.success('Project deleted');
    router.push({ name: 'projects' });
  } catch (e: unknown) {
    errorToast('Failed to delete project', { description: errorMessage(e) });
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>Project Settings</DialogTitle>
      </DialogHeader>

      <!-- Danger Zone -->
      <section class="space-y-4">
        <h3 class="text-sm font-medium text-destructive">Danger Zone</h3>

        <div class="rounded-lg border border-destructive/30 p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-foreground">Delete Project</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this project and all its data.
              </p>
            </div>
            <AlertDialog>
              <AlertDialogTrigger as-child>
                <Button variant="destructive" size="sm">
                  <Trash2 :size="14" class="mr-1" />
                  Delete
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete project</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete <strong>{{ projectName }}</strong>.
                    All environments, services, and deployments will be permanently deleted.
                    This action cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction
                    :disabled="deleting"
                    @click="handleDeleteProject"
                  >
                    {{ deleting ? 'Deleting...' : 'Delete' }}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </section>
    </DialogContent>
  </Dialog>
</template>
