<script setup lang="ts">
import { ref } from 'vue';
import { Download } from 'lucide-vue-next';
import { toast } from '@/components/ui/sonner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  projectId: string;
  projectName: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const ejecting = ref(false);

async function handleEject() {
  ejecting.value = true;
  try {
    const url = `/api/eject/${encodeURIComponent(props.projectId)}`;
    const res = await fetch(url, { credentials: 'include' });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || 'Eject failed');
    }

    const blob = await res.blob();
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    const shortName = props.projectName.split('/').pop() || props.projectName;
    a.download = `${shortName}-ejected.zip`;
    a.click();
    URL.revokeObjectURL(a.href);

    emit('update:open', false);
    toast.success('Project ejected', {
      description: 'Your project has been exported as a self-contained zip file.',
    });
  } catch (e: unknown) {
    toast.error('Failed to eject project', { description: errorMessage(e) });
  } finally {
    ejecting.value = false;
  }
}
</script>

<template>
  <AlertDialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>Eject project</AlertDialogTitle>
        <AlertDialogDescription class="space-y-2">
          <span>
            This will export <strong>{{ projectName }}</strong> as a self-contained
            zip file with Helm charts, ArgoCD manifests, and setup instructions.
            You can run it independently without Lucity.
          </span>
          <span class="block text-xs text-muted-foreground">
            Your project will continue running on Lucity. Ejection is non-destructive.
          </span>
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>Cancel</AlertDialogCancel>
        <AlertDialogAction
          :disabled="ejecting"
          @click="handleEject"
        >
          <Download
            :size="14"
            class="mr-1"
          />
          {{ ejecting ? 'Ejecting...' : 'Download zip' }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
