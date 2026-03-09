<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, Database, Server, HardDrive } from 'lucide-vue-next';
import { DeleteDatabaseMutation } from '@/graphql/databases';
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
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Configuration
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <Database :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Version</span>
            <span class="font-mono text-sm font-medium text-foreground">PostgreSQL {{ database.version }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <Server :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Instances</span>
            <span class="text-sm font-medium text-foreground">{{ database.instances }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Storage</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ database.size }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Danger Zone -->
    <section class="mt-8">
      <div class="relative overflow-hidden rounded-lg border border-destructive/20">
        <div class="pattern-crosshatch pointer-events-none absolute inset-0 opacity-[0.04]" />
        <div class="relative border-b border-destructive/15 bg-destructive/[0.03] px-4 py-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-destructive/70">
            Danger Zone
          </h3>
        </div>
        <div class="relative px-4 py-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-foreground">Delete Database</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this database and all its data.
              </p>
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
                    This will permanently delete the database and all its data.
                    This action cannot be undone.
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
      </div>
    </section>
  </div>
</template>
