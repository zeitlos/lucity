<script setup lang="ts">
import { useMutation } from '@vue/apollo-composable';
import { Globe, Lock, Trash2 } from 'lucide-vue-next';
import { RemoveServiceMutation } from '@/graphql/services';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
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
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    public: boolean;
    framework?: string;
  };
}>();

const emit = defineEmits<{
  (e: 'removed'): void;
}>();

const { mutate: removeServiceMutate, loading: removing } = useMutation(RemoveServiceMutation);

async function handleRemoveService() {
  try {
    const res = await removeServiceMutate({
      projectId: props.projectId,
      service: props.service.name,
    });

    if (res?.errors?.length) {
      toast.error('Failed to remove service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service removed');
    emit('removed');
  } catch (e: unknown) {
    toast.error('Failed to remove service', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Service Info -->
    <section class="space-y-4">
      <h3 class="text-sm font-medium text-muted-foreground">Service Info</h3>

      <div class="space-y-3 rounded-lg border p-4">
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Name</span>
          <span class="text-sm font-medium text-foreground">{{ service.name }}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Port</span>
          <span class="text-sm font-medium text-foreground">{{ service.port || '—' }}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Visibility</span>
          <Badge :variant="service.public ? 'default' : 'secondary'">
            <component
              :is="service.public ? Globe : Lock"
              :size="12"
              class="mr-1"
            />
            {{ service.public ? 'Public' : 'Private' }}
          </Badge>
        </div>
        <div v-if="service.image" class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">Image</span>
          <span class="max-w-[200px] truncate font-mono text-xs text-muted-foreground">
            {{ service.image }}
          </span>
        </div>
      </div>
    </section>

    <Separator />

    <!-- Danger Zone -->
    <section class="space-y-4">
      <h3 class="text-sm font-medium text-destructive">Danger Zone</h3>

      <div class="rounded-lg border border-destructive/30 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-foreground">Delete Service</p>
            <p class="text-xs text-muted-foreground">
              Permanently remove this service from the project.
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
                <AlertDialogTitle>Remove service</AlertDialogTitle>
                <AlertDialogDescription>
                  This will remove <strong>{{ service.name }}</strong> from the project
                  configuration. This action cannot be undone.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  :disabled="removing"
                  @click="handleRemoveService"
                >
                  {{ removing ? 'Removing...' : 'Remove' }}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </section>
  </div>
</template>
