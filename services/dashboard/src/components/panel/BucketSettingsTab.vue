<script setup lang="ts">
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, MapPin, HardDrive, Files } from '@lucide/vue';
import { graphql } from '@/gql';
import { Button } from '@/components/ui/button';
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
    toast.success('Bucket deleted');
    deleteDialogOpen.value = false;
    emit('bucket-removed');
  } catch (e: unknown) {
    errorToast('Failed to delete bucket', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Configuration -->
    <section class="space-y-2">
      <h3 class="text-[13px] font-medium text-muted-foreground">
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

    <div class="flex items-center justify-between gap-4 rounded-lg border border-destructive/30 p-3">
      <div>
        <p class="text-[13px] font-medium text-foreground">Delete bucket</p>
        <p class="text-xs text-muted-foreground">Permanently delete this bucket and all its objects.</p>
      </div>
      <ConfirmDialog
        v-model:open="deleteDialogOpen"
        :title="`Delete bucket &quot;${bucket.name}&quot;?`"
        description="Every object in it is deleted with it. This can't be undone."
        confirm-label="Delete bucket"
        pending-label="Deleting…"
        :pending="deleting"
        :confirm-text="bucket.name"
        destructive
        @confirm="handleDelete"
      >
        <template #trigger>
          <Button variant="destructive">
            <Trash2 />
            Delete
          </Button>
        </template>
      </ConfirmDialog>
    </div>
  </div>
</template>
