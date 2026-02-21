<script setup lang="ts">
import { ref, computed } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { Trash2, Copy, X } from 'lucide-vue-next';
import { RemoveServiceMutation, UpdateServiceConfigMutation, SetServiceDomainMutation } from '@/graphql/services';
import { useEnvironment } from '@/composables/useEnvironment';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
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
  (e: 'updated'): void;
}>();

const { activeEnvironment } = useEnvironment();

// Derive the current service instance for the active environment
const activeInstance = computed(() => {
  return activeEnvironment.value?.services.find(s => s.name === props.service.name);
});

const currentHost = computed(() => activeInstance.value?.host ?? '');

// Domain input state
const domainInput = ref('');
const editingDomain = ref(false);

// Internal DNS name
const internalDns = computed(() => {
  const envName = activeEnvironment.value?.name;
  if (!envName) return '';
  const parts = props.projectId.split('/');
  const shortProject = parts.length > 1 ? parts[1] : parts[0];
  const ns = `${shortProject}-${envName}`;
  return `${shortProject}-lucity-app-${props.service.name}.${ns}.svc.cluster.local`;
});

// Mutations
const { mutate: removeServiceMutate, loading: removing } = useMutation(RemoveServiceMutation);
const { mutate: updateConfigMutate, loading: updatingConfig } = useMutation(UpdateServiceConfigMutation);
const { mutate: setDomainMutate, loading: settingDomain } = useMutation(SetServiceDomainMutation);

async function handleTogglePublic() {
  try {
    const res = await updateConfigMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        public: !props.service.public,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to update visibility', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success(`Service is now ${props.service.public ? 'private' : 'public'}`);
    emit('updated');
  } catch (e: unknown) {
    toast.error('Failed to update visibility', { description: errorMessage(e) });
  }
}

async function handleSetDomain() {
  const host = domainInput.value.trim();
  if (!host) return;

  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const res = await setDomainMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        environment: envName,
        host,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to set domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success(`Domain set to ${host}`);
    domainInput.value = '';
    editingDomain.value = false;
    emit('updated');
  } catch (e: unknown) {
    toast.error('Failed to set domain', { description: errorMessage(e) });
  }
}

async function handleRemoveDomain() {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const res = await setDomainMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        environment: envName,
        host: '',
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to remove domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Domain removed');
    emit('updated');
  } catch (e: unknown) {
    toast.error('Failed to remove domain', { description: errorMessage(e) });
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
  toast.success('Copied to clipboard');
}

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
          <span class="text-sm font-medium text-foreground">{{ service.port || '---' }}</span>
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

    <!-- Networking -->
    <section class="space-y-4">
      <h3 class="text-sm font-medium text-muted-foreground">Networking</h3>

      <div class="space-y-4 rounded-lg border p-4">
        <!-- Visibility toggle -->
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-foreground">Public</p>
            <p class="text-xs text-muted-foreground">
              Expose this service to the internet via Gateway API.
            </p>
          </div>
          <Switch
            :checked="service.public"
            :disabled="updatingConfig"
            @update:checked="handleTogglePublic"
          />
        </div>

        <!-- Domain (only when public) -->
        <template v-if="service.public">
          <Separator />

          <div>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-foreground">Domain</p>
                <p class="text-xs text-muted-foreground">
                  Custom hostname for {{ activeEnvironment?.name ?? 'this environment' }}.
                </p>
              </div>
              <Badge v-if="currentHost" variant="outline" class="font-mono text-xs">
                {{ currentHost }}
              </Badge>
            </div>

            <!-- Current domain with remove option -->
            <div v-if="currentHost && !editingDomain" class="mt-3 flex items-center gap-2">
              <div class="flex-1 rounded-md border bg-muted/50 px-3 py-2">
                <span class="font-mono text-sm">{{ currentHost }}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                class="h-8 w-8 shrink-0"
                @click="copyToClipboard(currentHost)"
              >
                <Copy :size="14" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="h-8 w-8 shrink-0 text-destructive"
                :disabled="settingDomain"
                @click="handleRemoveDomain"
              >
                <X :size="14" />
              </Button>
            </div>

            <!-- Edit/add domain input -->
            <div v-if="!currentHost || editingDomain" class="mt-3 flex items-center gap-2">
              <Input
                v-model="domainInput"
                placeholder="api.example.com"
                class="flex-1 font-mono text-sm"
                @keyup.enter="handleSetDomain"
              />
              <Button
                size="sm"
                :disabled="!domainInput.trim() || settingDomain"
                @click="handleSetDomain"
              >
                {{ settingDomain ? 'Saving...' : 'Save' }}
              </Button>
            </div>

            <!-- Change button when domain is set -->
            <div v-if="currentHost && !editingDomain" class="mt-2">
              <Button
                variant="link"
                size="sm"
                class="h-auto p-0 text-xs"
                @click="editingDomain = true; domainInput = currentHost"
              >
                Change domain
              </Button>
            </div>
          </div>
        </template>

        <!-- Internal DNS -->
        <Separator />
        <div>
          <p class="text-sm font-medium text-foreground">Private Networking</p>
          <p class="text-xs text-muted-foreground">
            Internal DNS name for service-to-service communication.
          </p>
          <div v-if="internalDns" class="mt-2 flex items-center gap-2">
            <div class="flex-1 overflow-x-auto rounded-md border bg-muted/50 px-3 py-2">
              <span class="whitespace-nowrap font-mono text-xs">{{ internalDns }}</span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              class="h-8 w-8 shrink-0"
              @click="copyToClipboard(internalDns)"
            >
              <Copy :size="14" />
            </Button>
          </div>
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
