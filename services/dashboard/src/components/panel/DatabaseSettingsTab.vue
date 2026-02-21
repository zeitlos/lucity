<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2 } from 'lucide-vue-next';
import { DeleteDatabaseMutation } from '@/graphql/databases';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
import { toast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  database: {
    name: string;
    version: string;
    instances: number;
    size: string;
  };
}>();

const emit = defineEmits<{
  (e: 'database-removed'): void;
}>();

const { mutate: deleteDatabase, loading: deleting } = useMutation(DeleteDatabaseMutation);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteDatabase({
      projectId: props.projectId,
      name: props.database.name,
    });
    toast.success('Database removed');
    deleteDialogOpen.value = false;
    emit('database-removed');
  } catch (e: unknown) {
    toast.error('Failed to remove database', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Configuration -->
    <section class="space-y-3">
      <h3 class="text-sm font-medium text-foreground">Configuration</h3>
      <div class="rounded-lg border p-4">
        <dl class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt class="text-muted-foreground">Version</dt>
            <dd class="mt-1 font-medium">
              <Badge variant="secondary">PostgreSQL {{ database.version }}</Badge>
            </dd>
          </div>
          <div>
            <dt class="text-muted-foreground">Instances</dt>
            <dd class="mt-1 font-medium">{{ database.instances }}</dd>
          </div>
          <div>
            <dt class="text-muted-foreground">Storage</dt>
            <dd class="mt-1 font-mono font-medium">{{ database.size }}</dd>
          </div>
          <div>
            <dt class="text-muted-foreground">Operator</dt>
            <dd class="mt-1 font-medium text-muted-foreground">CloudNativePG</dd>
          </div>
        </dl>
      </div>
    </section>

    <!-- Danger Zone -->
    <section class="space-y-3">
      <h3 class="text-sm font-medium text-destructive">Danger Zone</h3>
      <div class="rounded-lg border border-destructive/30 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-foreground">Delete Database</p>
            <p class="text-xs text-muted-foreground">This will remove the database from the GitOps configuration.</p>
          </div>
          <AlertDialog v-model:open="deleteDialogOpen">
            <AlertDialogTrigger as-child>
              <Button variant="destructive" size="sm">
                <Trash2 :size="14" class="mr-1" />
                Delete
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete database "{{ database.name }}"?</AlertDialogTitle>
                <AlertDialogDescription>
                  This will remove the PostgreSQL cluster definition from the project configuration.
                  The CNPG operator will delete the cluster and its data.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  :disabled="deleting"
                  class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  @click="handleDelete"
                >
                  {{ deleting ? 'Deleting...' : 'Delete Database' }}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </section>
  </div>
</template>
