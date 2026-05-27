<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import {
  Trash2, Copy, X, Globe, Plus, Minus,
  ChevronDown, Network, ExternalLink, Scaling, GitBranch, Github, Play, Container, ArrowRight,
  Cpu, MemoryStick, Leaf, ShieldCheck,
} from 'lucide-vue-next';
import { graphql } from '@/gql';
import {
  type SetServiceScalingInput,
  ResourceTier,
} from '@/gql/graphql';

const RemoveServiceDocument = graphql(`
  mutation RemoveService($service: ServiceID!) {
    removeService(service: $service)
  }
`);

const SetCustomStartCommandDocument = graphql(`
  mutation SetCustomStartCommand($service: ServiceID!, $command: String!) {
    setCustomStartCommand(service: $service, command: $command) {
      id
    }
  }
`);

const GenerateDomainDocument = graphql(`
  mutation GenerateDomain($service: ServiceID!) {
    generateDomain(service: $service) {
      id
      endpoints {
        host
      }
    }
  }
`);

const AddCustomDomainDocument = graphql(`
  mutation AddCustomDomain($service: ServiceID!, $hostname: String!) {
    addCustomDomain(service: $service, hostname: $hostname) {
      id
    }
  }
`);

const RemoveDomainDocument = graphql(`
  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {
    removeDomain(service: $service, hostname: $hostname) {
      id
    }
  }
`);

const PlatformConfigDocument = graphql(`
  query PlatformConfig {
    platformConfig {
      workloadDomain
      domainTarget
      ipAddress
    }
  }
`);

const SetServiceScalingDocument = graphql(`
  mutation SetServiceScaling($input: SetServiceScalingInput!) {
    setServiceScaling(input: $input) {
      id
      replicas {
        desired
        ready
      }
      autoscaling {
        minReplicas
        maxReplicas
        targetCpu
      }
    }
  }
`);
import { useEnvironment } from '@/composables/useEnvironment';
import type { Service } from '@/composables/useEnvironment';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast, errorToast } from '@/components/ui/sonner';
import { Switch } from '@/components/ui/switch';
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
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  service: Service;
}>();

const emit = defineEmits<{
  (e: 'removed'): void;
  (e: 'refetch'): void;
}>();

const { activeEnvironment } = useEnvironment();

const activeDeployment = computed(() => props.service.activeDeployment ?? null);

const endpoints = computed(() => props.service.endpoints ?? []);
const platformEndpoint = computed(() => endpoints.value[0] ?? null);
const customEndpoints = computed(() => endpoints.value.slice(1));

const resources = computed(() => props.service.resources ?? null);
const resourceTier = computed(() => activeEnvironment.value?.resourceTier ?? null);

// Platform config
const { result: platformConfigResult } = useQuery(PlatformConfigDocument);
const domainTarget = computed(() => platformConfigResult.value?.platformConfig?.domainTarget ?? '');
const ipAddress = computed(() => platformConfigResult.value?.platformConfig?.ipAddress ?? '');

function domainUrl(hostname: string) {
  if (hostname.endsWith('.local')) return `http://${hostname}:8880`;
  return `https://${hostname}`;
}

// Custom domain input
const customDomainInput = ref('');

const hostnamePattern = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

function normalizeHostname(input: string): string {
  let h = input.trim().toLowerCase();
  h = h.replace(/^https?:\/\//, '');
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
  if (endpoints.value.some(d => d.host === hostname)) {
    return 'This domain is already added';
  }
  return '';
});

const canAddDomain = computed(() => {
  const raw = customDomainInput.value.trim();
  return raw.length > 0 && !hostnameError.value && !addingCustomDomain.value;
});

// Internal DNS — derived from service ID (workspace/project/env/name)
const internalDns = computed(() => {
  const parts = props.service.id.split('/');
  if (parts.length < 4) return '';
  const projectSlug = parts[1];
  const envName = parts[2];
  const svcName = parts[3];
  const ns = `${projectSlug}-${envName}`;
  return `${projectSlug}-lucity-app-${svcName}.${ns}.svc.cluster.local`;
});

// Command override
const customStartCommand = ref(
  props.service.command !== props.service.defaultCommand ? props.service.command : '',
);
const commandSaving = ref(false);
const { mutate: setCustomStartCommandMutate } = useMutation(SetCustomStartCommandDocument);

watch(
  () => [props.service.command, props.service.defaultCommand] as const,
  ([cmd, def]) => {
    customStartCommand.value = cmd !== def ? cmd : '';
  },
);

async function handleSaveCommand() {
  commandSaving.value = true;
  try {
    await setCustomStartCommandMutate({
      service: props.service.id,
      command: customStartCommand.value,
    });
    toast.success(customStartCommand.value ? 'Start command updated' : 'Start command cleared');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update start command', { description: errorMessage(e) });
  } finally {
    commandSaving.value = false;
  }
}

const commandIsCustom = computed(() => props.service.command !== props.service.defaultCommand);
const commandChanged = computed(() => {
  const current = commandIsCustom.value ? props.service.command : '';
  return customStartCommand.value !== current;
});

// Mutations
const { mutate: removeServiceMutate, loading: removing } = useMutation(RemoveServiceDocument);
const { mutate: generateDomainMutate, loading: generatingDomain } = useMutation(GenerateDomainDocument);
const { mutate: addCustomDomainMutate, loading: addingCustomDomain } = useMutation(AddCustomDomainDocument);
const { mutate: removeDomainMutate } = useMutation(RemoveDomainDocument);

async function handleGenerateDomain() {
  try {
    const res = await generateDomainMutate({ service: props.service.id });

    if (res?.errors?.length) {
      errorToast('Failed to generate domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const host = res?.data?.generateDomain?.endpoints?.[0]?.host;
    toast.success(host ? `Domain generated: ${host}` : 'Domain generated');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to generate domain', { description: errorMessage(e) });
  }
}

async function handleAddCustomDomain() {
  const hostname = normalizeHostname(customDomainInput.value);
  if (!hostname || hostnameError.value) return;

  try {
    const res = await addCustomDomainMutate({
      service: props.service.id,
      hostname,
    });

    if (res?.errors?.length) {
      errorToast('Failed to add custom domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    customDomainInput.value = '';
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to add custom domain', { description: errorMessage(e) });
  }
}

async function handleRemoveDomain(hostname: string) {
  try {
    const res = await removeDomainMutate({
      service: props.service.id,
      hostname,
    });

    if (res?.errors?.length) {
      errorToast('Failed to remove domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Domain removed');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to remove domain', { description: errorMessage(e) });
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
  toast.success('Copied to clipboard');
}

// Source
const sourceRepo = computed(() => {
  const url = props.service.sourceUrl;
  if (!url) return null;
  const match = url.match(/github\.com\/([^/]+\/[^/]+)/);
  return match ? match[1] : url.replace(/^https?:\/\//, '');
});

const sourceRepoUrl = computed(() => {
  const url = props.service.sourceUrl;
  if (!url) return null;
  return url.startsWith('http') ? url : `https://${url}`;
});

const isFromRepo = computed(() => !!props.service.sourceUrl);

// Scaling
const autoscalingEnabled = ref(false);
const scalingReplicas = ref(1);
const scalingMinReplicas = ref(1);
const scalingMaxReplicas = ref(10);
const scalingTargetCPU = ref(70);
const scalingSaving = ref(false);

const { mutate: setScalingMutate } = useMutation(SetServiceScalingDocument);

function syncScalingFromService() {
  const svc = props.service;
  if (svc.autoscaling) {
    autoscalingEnabled.value = true;
    scalingReplicas.value = svc.replicas?.desired ?? 1;
    scalingMinReplicas.value = svc.autoscaling.minReplicas;
    scalingMaxReplicas.value = svc.autoscaling.maxReplicas;
    scalingTargetCPU.value = svc.autoscaling.targetCpu;
  } else {
    autoscalingEnabled.value = false;
    scalingReplicas.value = svc.replicas?.desired ?? 1;
    scalingMinReplicas.value = 1;
    scalingMaxReplicas.value = 10;
    scalingTargetCPU.value = 70;
  }
}

watch(() => props.service, syncScalingFromService, { immediate: true });

const scalingSummary = computed(() => {
  const svc = props.service;
  if (!svc.replicas) return 'Not deployed';
  if (svc.autoscaling) {
    return `${svc.autoscaling.minReplicas}–${svc.autoscaling.maxReplicas} replicas · autoscaling`;
  }
  const r = svc.replicas.desired ?? 1;
  return `${r} replica${r !== 1 ? 's' : ''} · manual`;
});

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

async function handleSaveScaling() {
  scalingSaving.value = true;
  try {
    const input: SetServiceScalingInput = {
      service: props.service.id,
      replicas: scalingReplicas.value,
    };

    if (autoscalingEnabled.value) {
      input.autoscaling = {
        enabled: true,
        minReplicas: scalingMinReplicas.value,
        maxReplicas: scalingMaxReplicas.value,
        targetCPU: scalingTargetCPU.value,
      };
    } else {
      input.autoscaling = {
        enabled: false,
        minReplicas: 1,
        maxReplicas: 1,
        targetCPU: 70,
      };
    }

    await setScalingMutate({ input });
    toast.success('Scaling updated');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update scaling', { description: errorMessage(e) });
  } finally {
    scalingSaving.value = false;
  }
}

async function handleRemoveService() {
  try {
    const res = await removeServiceMutate({ service: props.service.id });

    if (res?.errors?.length) {
      errorToast('Failed to remove service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service removed');
    emit('removed');
  } catch (e: unknown) {
    errorToast('Failed to remove service', { description: errorMessage(e) });
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- General -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        General
      </h3>

      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <div class="rounded-lg bg-muted/60 p-1.5">
              <Github v-if="isFromRepo" :size="20" />
              <Container v-else :size="20" />
            </div>
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">{{ service.name }}</p>
              <p class="text-xs text-muted-foreground">
                {{ isFromRepo ? 'GitHub repository' : 'Container image' }}
              </p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <div v-if="activeDeployment?.image" class="space-y-1.5">
                <Label class="text-xs font-medium">Image</Label>
                <div class="group flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
                  <Container :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ activeDeployment.image }}</span>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-5 w-5 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
                    @click="copyToClipboard(activeDeployment!.image)"
                  >
                    <Copy :size="10" />
                  </Button>
                </div>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </section>

    <!-- Source -->
    <section v-if="service.sourceUrl" class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Source
      </h3>

      <Collapsible>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Github :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Repository</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ sourceRepo ?? 'Not connected' }}
              </p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <div class="space-y-1.5">
                <Label class="text-xs font-medium">Source Repo</Label>
                <a
                  v-if="sourceRepoUrl"
                  :href="sourceRepoUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2 transition-colors hover:bg-muted/80"
                >
                  <Github :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ service.sourceUrl }}</span>
                  <ExternalLink :size="12" class="shrink-0 text-muted-foreground" />
                </a>
              </div>

              <div v-if="service.contextPath && service.contextPath !== '.'" class="space-y-1.5">
                <Label class="text-xs font-medium">Root Directory</Label>
                <div class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
                  <span class="truncate font-mono text-sm">{{ service.contextPath }}</span>
                </div>
              </div>

              <div class="space-y-1.5">
                <Label class="text-xs font-medium">Branch</Label>
                <div class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
                  <GitBranch :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="font-mono text-sm">Default branch</span>
                  <span class="text-xs text-muted-foreground">(auto-deploy)</span>
                </div>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </section>

    <!-- Deploy -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Deploy
      </h3>

      <Collapsible :default-open="commandIsCustom">
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Play :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Custom Start Command</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ service.command || service.defaultCommand || 'Not configured' }}
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
                Command that will be run to start new deployments. Overrides the default command.
              </p>
              <Input
                v-model="customStartCommand"
                :placeholder="service.defaultCommand || 'npm run start'"
                class="font-mono text-sm"
                @keyup.enter="commandChanged && handleSaveCommand()"
              />
              <div class="flex justify-end">
                <Button
                  size="sm"
                  :disabled="!commandChanged || commandSaving"
                  @click="handleSaveCommand"
                >
                  {{ commandSaving ? 'Saving...' : 'Save' }}
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </section>

    <!-- Networking -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Networking
      </h3>

      <!-- Platform Domain (first endpoint) -->
      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Globe :size="16" class="shrink-0 text-primary" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Platform Domain</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ platformEndpoint ? platformEndpoint.host : 'Not configured' }}
              </p>
            </div>
            <Badge
              v-if="platformEndpoint"
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
              <div v-if="!platformEndpoint">
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

              <div v-else class="space-y-2">
                <div class="flex items-center gap-2">
                  <a
                    :href="domainUrl(platformEndpoint.host)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex flex-1 items-center gap-2 rounded-md border bg-muted/50 px-3 py-2 transition-colors hover:bg-muted/80"
                  >
                    <Globe :size="14" class="shrink-0 text-muted-foreground" />
                    <span class="truncate font-mono text-sm hover:underline">{{ platformEndpoint.host }}</span>
                    <ExternalLink :size="12" class="ml-auto shrink-0 text-muted-foreground" />
                  </a>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    @click="copyToClipboard(platformEndpoint!.host)"
                  >
                    <Copy :size="14" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0 text-destructive"
                    @click="handleRemoveDomain(platformEndpoint!.host)"
                  >
                    <X :size="14" />
                  </Button>
                </div>
                <div class="flex items-center gap-1.5 pl-1 text-xs text-muted-foreground">
                  <ArrowRight :size="10" class="shrink-0" />
                  <span>
                    Routes to port
                    <span class="font-mono font-medium text-foreground">{{ platformEndpoint.port }}</span>
                  </span>
                </div>
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
                {{ customEndpoints.length }} domain{{ customEndpoints.length !== 1 ? 's' : '' }} configured
              </p>
            </div>
            <span
              v-if="customEndpoints.length"
              class="flex h-5 w-5 items-center justify-center rounded-full bg-muted text-[0.6rem] font-medium text-muted-foreground"
            >
              {{ customEndpoints.length }}
            </span>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <div v-if="customEndpoints.length" class="space-y-2">
                <div
                  v-for="endpoint in customEndpoints"
                  :key="endpoint.host"
                  class="flex items-center gap-2"
                >
                  <a
                    :href="domainUrl(endpoint.host)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex flex-1 items-center gap-2 rounded-md border bg-muted/50 px-3 py-2 transition-colors hover:bg-muted/80"
                  >
                    <Globe :size="14" class="shrink-0 text-muted-foreground" />
                    <span class="truncate font-mono text-sm hover:underline">{{ endpoint.host }}</span>
                    <span class="ml-auto shrink-0 text-xs text-muted-foreground">:{{ endpoint.port }}</span>
                    <ExternalLink :size="12" class="shrink-0 text-muted-foreground" />
                  </a>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    @click="copyToClipboard(endpoint.host)"
                  >
                    <Copy :size="14" />
                  </Button>
                  <AlertDialog>
                    <AlertDialogTrigger as-child>
                      <Button
                        variant="ghost"
                        size="icon"
                        class="h-8 w-8 shrink-0 text-destructive"
                      >
                        <X :size="14" />
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Remove domain</AlertDialogTitle>
                        <AlertDialogDescription>
                          Remove <strong class="font-mono">{{ endpoint.host }}</strong> from this service? This will also delete the TLS certificate.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                          @click="handleRemoveDomain(endpoint.host)"
                        >
                          Remove
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </div>
              </div>

              <!-- DNS/TLS status will be added back when endpoint fields land server-side. -->
              <p v-if="customEndpoints.length" class="text-[11px] text-muted-foreground">
                Point a CNAME record at <span class="font-mono">{{ domainTarget || '...' }}</span>
                <template v-if="ipAddress"> (apex domains: A record to <span class="font-mono">{{ ipAddress }}</span>)</template>.
              </p>

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
              <div v-if="internalDns" class="space-y-2">
                <div class="group flex items-center gap-2">
                  <div class="flex-1 overflow-x-auto rounded-md border bg-muted/50 px-3 py-2">
                    <span class="whitespace-nowrap font-mono text-xs">{{ internalDns }}<template v-if="platformEndpoint">:{{ platformEndpoint.port }}</template></span>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    @click="copyToClipboard(platformEndpoint ? `${internalDns}:${platformEndpoint.port}` : internalDns)"
                  >
                    <Copy :size="14" />
                  </Button>
                </div>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </section>

    <!-- Scaling -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Scaling
      </h3>

      <Collapsible>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Scaling :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Replicas</p>
              <p class="text-xs text-muted-foreground">{{ scalingSummary }}</p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-4 border-t px-4 py-3">
              <!-- Replicas -->
              <div class="space-y-1.5">
                <Label class="text-xs font-medium">Replicas</Label>
                <div class="flex items-center gap-1">
                  <Button
                    variant="outline"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    :disabled="autoscalingEnabled || scalingReplicas <= 1"
                    @click="scalingReplicas = clamp(scalingReplicas - 1, 1, 20)"
                  >
                    <Minus :size="14" />
                  </Button>
                  <Input
                    v-model.number="scalingReplicas"
                    type="number"
                    class="h-8 w-16 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    :min="1"
                    :max="20"
                    :disabled="autoscalingEnabled"
                    @blur="scalingReplicas = clamp(scalingReplicas, 1, 20)"
                  />
                  <Button
                    variant="outline"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    :disabled="autoscalingEnabled || scalingReplicas >= 20"
                    @click="scalingReplicas = clamp(scalingReplicas + 1, 1, 20)"
                  >
                    <Plus :size="14" />
                  </Button>
                </div>
                <p v-if="autoscalingEnabled" class="text-[11px] text-muted-foreground">
                  Managed by autoscaler.
                </p>
              </div>

              <!-- Autoscaling toggle -->
              <div class="flex items-center justify-between">
                <div>
                  <Label class="text-sm font-medium">Autoscaling</Label>
                  <p class="text-xs text-muted-foreground">Scale replicas based on CPU usage.</p>
                </div>
                <Switch v-model="autoscalingEnabled" class="data-[state=unchecked]:bg-border" />
              </div>

              <!-- Autoscaling settings -->
              <div v-if="autoscalingEnabled" class="grid grid-cols-3 gap-3">
                <div class="space-y-1.5">
                  <Label class="text-xs font-medium">Min</Label>
                  <div class="flex items-center gap-0.5">
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingMinReplicas <= 1"
                      @click="scalingMinReplicas = clamp(scalingMinReplicas - 1, 1, 20)"
                    >
                      <Minus :size="12" />
                    </Button>
                    <Input
                      v-model.number="scalingMinReplicas"
                      type="number"
                      class="h-8 w-full min-w-0 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                      :min="1"
                      :max="20"
                      @blur="scalingMinReplicas = clamp(scalingMinReplicas, 1, 20)"
                    />
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingMinReplicas >= 20"
                      @click="scalingMinReplicas = clamp(scalingMinReplicas + 1, 1, 20)"
                    >
                      <Plus :size="12" />
                    </Button>
                  </div>
                </div>

                <div class="space-y-1.5">
                  <Label class="text-xs font-medium">Max</Label>
                  <div class="flex items-center gap-0.5">
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingMaxReplicas <= 1"
                      @click="scalingMaxReplicas = clamp(scalingMaxReplicas - 1, 1, 20)"
                    >
                      <Minus :size="12" />
                    </Button>
                    <Input
                      v-model.number="scalingMaxReplicas"
                      type="number"
                      class="h-8 w-full min-w-0 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                      :min="1"
                      :max="20"
                      @blur="scalingMaxReplicas = clamp(scalingMaxReplicas, 1, 20)"
                    />
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingMaxReplicas >= 20"
                      @click="scalingMaxReplicas = clamp(scalingMaxReplicas + 1, 1, 20)"
                    >
                      <Plus :size="12" />
                    </Button>
                  </div>
                </div>

                <div class="space-y-1.5">
                  <Label class="text-xs font-medium">CPU target</Label>
                  <div class="flex items-center gap-0.5">
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingTargetCPU <= 10"
                      @click="scalingTargetCPU = clamp(scalingTargetCPU - 5, 10, 95)"
                    >
                      <Minus :size="12" />
                    </Button>
                    <div class="relative flex-1">
                      <Input
                        v-model.number="scalingTargetCPU"
                        type="number"
                        class="h-8 w-full min-w-0 pr-6 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                        :min="10"
                        :max="95"
                        @blur="scalingTargetCPU = clamp(scalingTargetCPU, 10, 95)"
                      />
                      <span class="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">%</span>
                    </div>
                    <Button
                      variant="outline"
                      size="icon"
                      class="h-8 w-8 shrink-0"
                      :disabled="scalingTargetCPU >= 95"
                      @click="scalingTargetCPU = clamp(scalingTargetCPU + 5, 10, 95)"
                    >
                      <Plus :size="12" />
                    </Button>
                  </div>
                </div>
              </div>

              <div class="flex justify-end">
                <Button
                  size="sm"
                  :disabled="scalingSaving"
                  @click="handleSaveScaling"
                >
                  {{ scalingSaving ? 'Saving...' : 'Save' }}
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>
    </section>

    <!-- Compute -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Compute
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div v-if="resources" class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <Cpu :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">CPU</p>
              <p class="text-xs text-muted-foreground">Per instance</p>
            </div>
            <span class="font-mono text-sm font-medium">{{ resources.cpu }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <MemoryStick :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Memory</p>
              <p class="text-xs text-muted-foreground">Per instance</p>
            </div>
            <span class="font-mono text-sm font-medium">{{ resources.memory }}</span>
          </div>
          <div v-if="resourceTier === ResourceTier.Eco" class="flex items-center gap-3 px-4 py-3">
            <Leaf :size="16" class="shrink-0 text-emerald-500" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Eco</p>
              <p class="text-xs text-muted-foreground">Burstable compute, shared resources</p>
            </div>
          </div>
          <div v-else-if="resourceTier === ResourceTier.Production" class="flex items-center gap-3 px-4 py-3">
            <ShieldCheck :size="16" class="shrink-0 text-violet-500" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Production</p>
              <p class="text-xs text-muted-foreground">Guaranteed performance, dedicated resources</p>
            </div>
          </div>
        </div>
        <div v-else class="px-4 py-3">
          <p class="text-sm text-muted-foreground">
            No deployment yet.
          </p>
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
  </div>
</template>
