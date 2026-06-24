<script setup lang="ts">
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, MapPin, HardDrive, Files } from '@lucide/vue';
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
import { errorMessage, formatBytes } from '@/lib/utils';

const DeleteBucketDocument = graphql(`
  mutation DeleteBucket($bucket: BucketID!) {
    deleteBucket(bucket: $bucket)
  }
`);

const props = defineProps<{
  bucketId: string;
  bucket: {
    name: string;
    region: string;
    sizeBytes: number;
    objectCount: number;
  };
}>();

const emit = defineEmits<{
  (e: 'bucket-removed'): void;
}>();

const { mutate: deleteBucket, loading: deleting } = useMutation(DeleteBucketDocument);
const deleteDialogOpen = ref(false);

async function handleDelete() {
  try {
    await deleteBucket({ bucket: props.bucketId });
    toast.success('Bucket removed');
    deleteDialogOpen.value = false;
    emit('bucket-removed');
  } catch (e: unknown) {
    errorToast('Failed to remove bucket', { description: errorMessage(e) });
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
            <MapPin :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Region</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ bucket.region }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <HardDrive :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Stored</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ formatBytes(bucket.sizeBytes) }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <Files :size="16" class="shrink-0 text-muted-foreground" />
            <span class="flex-1 text-sm text-muted-foreground">Objects</span>
            <span class="font-mono text-sm font-medium text-foreground">{{ bucket.objectCount.toLocaleString() }}</span>
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
              <p class="text-sm font-medium text-foreground">Delete Bucket</p>
              <p class="text-xs text-muted-foreground">
                Permanently delete this bucket and all its objects.
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
                  <AlertDialogTitle>Delete bucket "{{ bucket.name }}"?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This will permanently delete the bucket and every object in it.
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
                    {{ deleting ? 'Deleting...' : 'Delete Bucket' }}
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
