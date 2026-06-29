<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, HardDrive } from '@lucide/vue';
import { graphql } from '@/gql';
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
import { toast, errorToast } from '@/components/ui/sonner';
import { errorMessage } from '@/lib/utils';

const DeleteVolumeDocument = graphql(`
  mutation DeleteVolume($volume: VolumeID!) {
    deleteVolume(volume: $volume)
  }
`);

const props = defineProps<{
  volumeId: string;
  volume: {
    name: string;
    size: string;
    mount?: { service: string; path: string } | null;
  };
}>();

const emit = defineEmits<{
  (e: 'volume-removed'): void;
}>();

const { mutate: deleteVolume, loading: deleting } = useMutation(DeleteVolumeDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteVolume({ volume: props.volumeId });
    toast.success('Volume deleted');
    deleteDialogOpen.value = false;
    emit('volume-removed');
  } catch (e: unknown) {
    errorToast('Failed to delete volume', { description: errorMessage(e) });
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
            <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Size</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ volume.size }}</span>
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
              <p class="text-sm font-medium text-foreground">Delete Volume</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this volume and all its data.
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
                  <AlertDialogTitle>Delete volume "{{ volume.name }}"?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete the volume and all the data stored on it.
                    <template v-if="volume.mount">
                      The volume is still mounted; unmount it first if you only want to detach it.
                    </template>
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
                    {{ deleting ? 'Deleting...' : 'Delete Volume' }}
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
