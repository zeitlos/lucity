<script setup lang="ts">
import { ref, computed } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import { Trash2, Copy, X, Globe, Plus, CircleCheck, CircleAlert, HelpCircle } from 'lucide-vue-next';
import {
  RemoveServiceMutation,
  GenerateDomainMutation,
  AddCustomDomainMutation,
  RemoveDomainMutation,
  PlatformConfigQuery,
} from '@/graphql/services';
import { useEnvironment } from '@/composables/useEnvironment';
import type { DomainInfo } from '@/composables/useEnvironment';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  projectId: string;
  service: {
    name: string;
    image: string;
    port: number;
    framework?: string;
  };
}>();

const emit = defineEmits<{
  (e: 'removed'): void;
}>();

const { activeEnvironment, updateServiceDomains } = useEnvironment();

// Derive the current service instance for the active environment
const activeInstance = computed(() => {
  return activeEnvironment.value?.services.find(s => s.name === props.service.name);
});

const domains = computed<DomainInfo[]>(() => activeInstance.value?.domains ?? []);
const platformDomain = computed(() => domains.value.find(d => d.type === 'PLATFORM'));
const customDomains = computed(() => domains.value.filter(d => d.type === 'CUSTOM'));

// Platform config
const { result: platformConfigResult } = useQuery(PlatformConfigQuery);
const domainTarget = computed(() => platformConfigResult.value?.platformConfig?.domainTarget ?? '');

// Custom domain input
const customDomainInput = ref('');

// DNS help modal
const dnsHelpOpen = ref(false);
const dnsHelpDomain = ref('');

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
const { mutate: generateDomainMutate, loading: generatingDomain } = useMutation(GenerateDomainMutation);
const { mutate: addCustomDomainMutate, loading: addingCustomDomain } = useMutation(AddCustomDomainMutation);
const { mutate: removeDomainMutate } = useMutation(RemoveDomainMutation);

async function handleGenerateDomain() {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const res = await generateDomainMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        environment: envName,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to generate domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const domain = res?.data?.generateDomain;
    if (domain) {
      updateServiceDomains(props.service.name, [...domains.value, domain]);
    }
    toast.success(`Domain generated: ${domain?.hostname}`);
  } catch (e: unknown) {
    toast.error('Failed to generate domain', { description: errorMessage(e) });
  }
}

async function handleAddCustomDomain() {
  const hostname = customDomainInput.value.trim();
  if (!hostname) return;

  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const res = await addCustomDomainMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        environment: envName,
        hostname,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to add custom domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const domain = res?.data?.addCustomDomain;
    if (domain) {
      updateServiceDomains(props.service.name, [...domains.value, domain]);
    }
    toast.success(`Custom domain added: ${hostname}`);
    customDomainInput.value = '';
  } catch (e: unknown) {
    toast.error('Failed to add custom domain', { description: errorMessage(e) });
  }
}

async function handleRemoveDomain(hostname: string) {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  try {
    const res = await removeDomainMutate({
      input: {
        projectId: props.projectId,
        service: props.service.name,
        environment: envName,
        hostname,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to remove domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    updateServiceDomains(props.service.name, domains.value.filter(d => d.hostname !== hostname));
    toast.success('Domain removed');
  } catch (e: unknown) {
    toast.error('Failed to remove domain', { description: errorMessage(e) });
  }
}

function showDnsHelp(hostname: string) {
  dnsHelpDomain.value = hostname;
  dnsHelpOpen.value = true;
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
        <!-- Platform Domain -->
        <div>
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-foreground">Platform Domain</p>
              <p class="text-xs text-muted-foreground">
                Auto-generated domain for {{ activeEnvironment?.name ?? 'this environment' }}.
              </p>
            </div>
          </div>

          <!-- No platform domain yet: show generate button -->
          <div v-if="!platformDomain" class="mt-3">
            <Button
              size="sm"
              variant="outline"
              :disabled="generatingDomain"
              @click="handleGenerateDomain"
            >
              <Globe :size="14" class="mr-1.5" />
              {{ generatingDomain ? 'Generating...' : 'Generate Domain' }}
            </Button>
          </div>

          <!-- Platform domain exists -->
          <div v-else class="mt-3 flex items-center gap-2">
            <div class="flex flex-1 items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
              <CircleCheck :size="14" class="shrink-0 text-green-500" />
              <span class="truncate font-mono text-sm">{{ platformDomain.hostname }}</span>
            </div>
            <Button
              variant="ghost"
              size="icon"
              class="h-8 w-8 shrink-0"
              @click="copyToClipboard(platformDomain.hostname)"
            >
              <Copy :size="14" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              class="h-8 w-8 shrink-0 text-destructive"
              @click="handleRemoveDomain(platformDomain.hostname)"
            >
              <X :size="14" />
            </Button>
          </div>
        </div>

        <Separator />

        <!-- Custom Domains -->
        <div>
          <p class="text-sm font-medium text-foreground">Custom Domains</p>
          <p class="text-xs text-muted-foreground">
            Add your own domains. Requires DNS configuration.
          </p>

          <!-- List of custom domains -->
          <div v-if="customDomains.length" class="mt-3 space-y-2">
            <div
              v-for="domain in customDomains"
              :key="domain.hostname"
              class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2"
            >
              <CircleCheck
                v-if="domain.dnsStatus === 'VALID'"
                :size="14"
                class="shrink-0 text-green-500"
              />
              <CircleAlert
                v-else
                :size="14"
                class="shrink-0 text-yellow-500"
              />
              <span class="flex-1 truncate font-mono text-sm">{{ domain.hostname }}</span>
              <Button
                v-if="domain.dnsStatus !== 'VALID' && domainTarget"
                variant="ghost"
                size="icon"
                class="h-7 w-7 shrink-0"
                title="Show DNS configuration"
                @click="showDnsHelp(domain.hostname)"
              >
                <HelpCircle :size="14" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 shrink-0"
                @click="copyToClipboard(domain.hostname)"
              >
                <Copy :size="14" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 shrink-0 text-destructive"
                @click="handleRemoveDomain(domain.hostname)"
              >
                <X :size="14" />
              </Button>
            </div>
          </div>

          <!-- Add custom domain input -->
          <div class="mt-3 flex items-center gap-2">
            <Input
              v-model="customDomainInput"
              placeholder="api.example.com"
              class="flex-1 font-mono text-sm"
              @keyup.enter="handleAddCustomDomain"
            />
            <Button
              size="sm"
              variant="outline"
              :disabled="!customDomainInput.trim() || addingCustomDomain"
              @click="handleAddCustomDomain"
            >
              <Plus :size="14" class="mr-1" />
              {{ addingCustomDomain ? 'Adding...' : 'Add' }}
            </Button>
          </div>
        </div>

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

    <!-- DNS Help Dialog -->
    <Dialog v-model:open="dnsHelpOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Configure DNS Records</DialogTitle>
          <DialogDescription>
            Add the following DNS record to point <strong class="font-mono">{{ dnsHelpDomain }}</strong> to your application.
          </DialogDescription>
        </DialogHeader>

        <div class="rounded-md border">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b bg-muted/50">
                <th class="px-3 py-2 text-left font-medium text-muted-foreground">Type</th>
                <th class="px-3 py-2 text-left font-medium text-muted-foreground">Name</th>
                <th class="px-3 py-2 text-left font-medium text-muted-foreground">Value</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td class="px-3 py-2">
                  <Badge variant="outline" class="font-mono text-xs">CNAME</Badge>
                </td>
                <td class="px-3 py-2 font-mono text-xs">{{ dnsHelpDomain }}</td>
                <td class="px-3 py-2">
                  <div class="flex items-center gap-1">
                    <span class="font-mono text-xs">{{ domainTarget }}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-6 w-6 shrink-0"
                      @click="copyToClipboard(domainTarget)"
                    >
                      <Copy :size="12" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <p class="text-xs text-muted-foreground">
          DNS changes can take up to 48 hours to propagate. The status will update automatically once the record is detected.
        </p>

        <DialogFooter>
          <Button variant="outline" @click="dnsHelpOpen = false">
            Done
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
