<script setup lang="ts">
import { ref, computed } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import {
  Trash2, Copy, X, Globe, Plus, CircleCheck, CircleAlert,
  HelpCircle, ChevronDown, Network, ExternalLink,
} from 'lucide-vue-next';
import {
  RemoveServiceMutation,
  GenerateDomainMutation,
  AddCustomDomainMutation,
  RemoveDomainMutation,
  PlatformConfigQuery,
} from '@/graphql/services';
import { useEnvironment } from '@/composables/useEnvironment';
import type { DomainInfo } from '@/composables/useEnvironment';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from '@/components/ui/sonner';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
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

function domainUrl(hostname: string) {
  if (hostname.endsWith('.local')) return `http://${hostname}:8880`;
  return `https://${hostname}`;
}

// Custom domain input
const customDomainInput = ref('');

// Hostname validation
const hostnamePattern = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

function normalizeHostname(input: string): string {
  let h = input.trim();
  h = h.replace(/^https?:\/\//, '');
  h = h.replace(/^www\./, '');
  h = h.replace(/\/+$/, '');
  return h;
}

const hostnameError = computed(() => {
  const raw = customDomainInput.value.trim();
  if (!raw) return '';
  const hostname = normalizeHostname(raw);
  if (!hostnamePattern.test(hostname)) {
    return 'Enter a valid domain (e.g. api.example.com)';
  }
  if (domains.value.some(d => d.hostname === hostname)) {
    return 'This domain is already added';
  }
  return '';
});

const canAddDomain = computed(() => {
  const raw = customDomainInput.value.trim();
  return raw.length > 0 && !hostnameError.value && !addingCustomDomain.value;
});

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
  const hostname = normalizeHostname(customDomainInput.value);
  if (!hostname || hostnameError.value) return;

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
    <!-- Identity Header -->
    <div class="relative overflow-hidden rounded-lg border p-5">
      <div class="pattern-dots pointer-events-none absolute inset-0 opacity-[0.12]" />
      <div class="relative flex items-start gap-4">
        <div class="rounded-xl bg-muted/60 p-2.5">
          <FrameworkIcon :framework="service.framework" :size="36" />
        </div>
        <div class="min-w-0 flex-1 space-y-2">
          <h3 class="truncate text-base font-semibold text-foreground">{{ service.name }}</h3>
          <div class="flex flex-wrap items-center gap-2">
            <Badge variant="secondary" class="font-mono text-xs">
              :{{ service.port || '---' }}
            </Badge>
            <Badge
              v-if="service.framework"
              variant="outline"
              class="text-xs"
            >
              {{ service.framework }}
            </Badge>
          </div>
          <div v-if="service.image" class="group flex items-center gap-1.5">
            <span class="max-w-[220px] truncate font-mono text-[11px] text-muted-foreground">
              {{ service.image }}
            </span>
            <Button
              variant="ghost"
              size="icon"
              class="h-5 w-5 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
              @click="copyToClipboard(service.image)"
            >
              <Copy :size="10" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <!-- Networking -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Networking
      </h3>

      <!-- Platform Domain -->
      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Globe :size="16" class="shrink-0 text-primary" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Platform Domain</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ platformDomain ? platformDomain.hostname : 'Not configured' }}
              </p>
            </div>
            <Badge
              v-if="platformDomain"
              variant="default"
              class="text-[0.6rem]"
            >
              Active
            </Badge>
            <Badge v-else variant="secondary" class="text-[0.6rem]">Off</Badge>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <!-- No platform domain yet -->
              <div v-if="!platformDomain">
                <p class="mb-2 text-xs text-muted-foreground">
                  Auto-generated domain for {{ activeEnvironment?.name ?? 'this environment' }}.
                </p>
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
              <div v-else class="flex items-center gap-2">
                <a
                  :href="domainUrl(platformDomain.hostname)"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex flex-1 items-center gap-2 rounded-md border bg-muted/50 px-3 py-2 transition-colors hover:bg-muted/80"
                >
                  <CircleCheck :size="14" class="shrink-0 text-green-500" />
                  <span class="truncate font-mono text-sm hover:underline">{{ platformDomain.hostname }}</span>
                  <ExternalLink :size="12" class="ml-auto shrink-0 text-muted-foreground" />
                </a>
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
          </CollapsibleContent>
        </div>
      </Collapsible>

      <!-- Custom Domains -->
      <Collapsible>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Globe :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Custom Domains</p>
              <p class="text-xs text-muted-foreground">
                {{ customDomains.length }} domain{{ customDomains.length !== 1 ? 's' : '' }} configured
              </p>
            </div>
            <span
              v-if="customDomains.length"
              class="flex h-5 w-5 items-center justify-center rounded-full bg-muted text-[0.6rem] font-medium text-muted-foreground"
            >
              {{ customDomains.length }}
            </span>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <!-- List of custom domains -->
              <div v-if="customDomains.length" class="space-y-2">
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
                  <a
                    :href="domainUrl(domain.hostname)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex-1 truncate font-mono text-sm hover:underline"
                  >
                    {{ domain.hostname }}
                  </a>
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
              <div class="space-y-1.5">
                <div class="flex items-center gap-2">
                  <Input
                    v-model="customDomainInput"
                    placeholder="api.example.com"
                    class="flex-1 font-mono text-sm"
                    :class="{ 'border-destructive': hostnameError }"
                    @keyup.enter="canAddDomain && handleAddCustomDomain()"
                  />
                  <Button
                    size="sm"
                    variant="outline"
                    :disabled="!canAddDomain"
                    @click="handleAddCustomDomain"
                  >
                    <Plus :size="14" class="mr-1" />
                    {{ addingCustomDomain ? 'Adding...' : 'Add' }}
                  </Button>
                </div>
                <p v-if="hostnameError" class="px-1 text-xs text-destructive">
                  {{ hostnameError }}
                </p>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>

      <!-- Private Networking -->
      <Collapsible>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Network :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Private Networking</p>
              <p class="max-w-[220px] truncate text-xs text-muted-foreground">
                {{ internalDns || 'Internal DNS' }}
              </p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <p class="text-xs text-muted-foreground">
                Internal DNS name for service-to-service communication.
              </p>
              <div v-if="internalDns" class="group flex items-center gap-2">
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
          </CollapsibleContent>
        </div>
      </Collapsible>
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
