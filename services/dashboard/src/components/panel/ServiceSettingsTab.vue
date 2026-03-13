<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useMutation, useQuery } from '@vue/apollo-composable';
import {
  Trash2, Copy, X, Globe, Plus, Minus, CircleCheck, CircleAlert,
  ChevronDown, Network, ExternalLink, Loader2, Scaling, GitBranch, Github, Code, Play, Container, ArrowRight,
  Cpu, MemoryStick,
} from 'lucide-vue-next';
import {
  RemoveServiceMutation,
  SetCustomStartCommandMutation,
  GenerateDomainMutation,
  AddCustomDomainMutation,
  RemoveDomainMutation,
  PlatformConfigQuery,
} from '@/graphql/services';
import { SetServiceScalingMutation } from '@/graphql/projects';
import { useEnvironment } from '@/composables/useEnvironment';
import type { DomainInfo } from '@/composables/useEnvironment';
import { useDnsPolling } from '@/composables/useDnsPolling';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { toast } from '@/components/ui/sonner';
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
    port?: number;
    framework?: string;
    sourceUrl?: string;
    contextPath?: string;
    startCommand?: string;
    customStartCommand?: string;
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

// Compute resources
const resources = computed(() => activeInstance.value?.resources ?? null);
const resourceTier = computed(() => activeEnvironment.value?.resourceTier ?? null);
const resourceTierLabel = computed(() => {
  if (!resourceTier.value) return null;
  return resourceTier.value === 'PRODUCTION' ? 'Production' : 'Eco';
});
function formatCpu(millicores: number): string {
  const vcpu = millicores / 1000;
  return vcpu % 1 === 0 ? `${vcpu} vCPU` : `${vcpu} vCPU`;
}
function formatMemory(mb: number): string {
  if (mb >= 1024 && mb % 1024 === 0) return `${mb / 1024} GB`;
  return `${mb} MB`;
}

// Platform config
const { result: platformConfigResult } = useQuery(PlatformConfigQuery);
const domainTarget = computed(() => platformConfigResult.value?.platformConfig?.domainTarget ?? '');

// DNS polling for custom domains
const dnsPolling = useDnsPolling();

// Start polling for unverified custom domains when they change
watch(customDomains, (domains) => {
  const unverified = domains
    .filter(d => d.dnsStatus !== 'VALID')
    .map(d => d.hostname);
  dnsPolling.trackHostnames(unverified);
}, { immediate: true });

// Get live DNS status for a custom domain (from polling, or fallback to static)
function dnsStatus(hostname: string): 'VALID' | 'PENDING' | 'MISCONFIGURED' | 'ERROR' {
  return dnsPolling.checks[hostname]?.status
    ?? customDomains.value.find(d => d.hostname === hostname)?.dnsStatus
    ?? 'PENDING';
}

// DNS records modal
const dnsModalOpen = ref(false);
const dnsModalHostname = ref('');

function showDnsRecords(hostname: string) {
  dnsModalHostname.value = hostname;
  dnsModalOpen.value = true;
}

const dnsModalStatus = computed(() => dnsStatus(dnsModalHostname.value));
const dnsModalMessage = computed(() => dnsPolling.checks[dnsModalHostname.value]?.message ?? null);
const dnsModalTarget = computed(() =>
  dnsPolling.checks[dnsModalHostname.value]?.expectedTarget ?? domainTarget.value,
);

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

// Internal DNS name
const internalDns = computed(() => {
  const envName = activeEnvironment.value?.name;
  if (!envName) return '';
  const parts = props.projectId.split('/');
  const shortProject = parts.length > 1 ? parts[1] : parts[0];
  const ns = `${shortProject}-${envName}`;
  return `${shortProject}-lucity-app-${props.service.name}.${ns}.svc.cluster.local`;
});

// Custom Start Command
const customStartCommand = ref(props.service.customStartCommand ?? '');
const commandSaving = ref(false);
const { mutate: setCustomStartCommandMutate } = useMutation(SetCustomStartCommandMutation);

watch(() => props.service.customStartCommand, (val) => {
  customStartCommand.value = val ?? '';
});

async function handleSaveCommand() {
  commandSaving.value = true;
  try {
    await setCustomStartCommandMutate({
      projectId: props.projectId,
      environment: activeEnvironment.value?.name,
      service: props.service.name,
      command: customStartCommand.value,
    });
    toast.success(customStartCommand.value ? 'Start command updated' : 'Start command cleared');
  } catch (e: unknown) {
    toast.error('Failed to update start command', { description: errorMessage(e) });
  } finally {
    commandSaving.value = false;
  }
}

const commandChanged = computed(() => {
  return customStartCommand.value !== (props.service.customStartCommand ?? '');
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
      dnsPolling.addHostname(hostname);
    }
    customDomainInput.value = '';
    // Auto-open the DNS records modal so the user sees what to configure
    showDnsRecords(hostname);
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

    dnsPolling.removeHostname(hostname);
    updateServiceDomains(props.service.name, domains.value.filter(d => d.hostname !== hostname));
    toast.success('Domain removed');
  } catch (e: unknown) {
    toast.error('Failed to remove domain', { description: errorMessage(e) });
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

// Scaling
const autoscalingEnabled = ref(false);
const scalingReplicas = ref(1);
const scalingMinReplicas = ref(1);
const scalingMaxReplicas = ref(10);
const scalingTargetCPU = ref(70);
const scalingSaving = ref(false);

const { mutate: setScalingMutate } = useMutation(SetServiceScalingMutation);

function syncScalingFromService() {
  const svc = activeInstance.value;
  if (!svc) return;

  if (svc.scaling?.autoscaling?.enabled) {
    autoscalingEnabled.value = true;
    scalingReplicas.value = svc.scaling.replicas || svc.replicas || 1;
    scalingMinReplicas.value = svc.scaling.autoscaling.minReplicas;
    scalingMaxReplicas.value = svc.scaling.autoscaling.maxReplicas;
    scalingTargetCPU.value = svc.scaling.autoscaling.targetCPU;
  } else {
    autoscalingEnabled.value = false;
    scalingReplicas.value = svc.scaling?.replicas || svc.replicas || 1;
    scalingMinReplicas.value = 1;
    scalingMaxReplicas.value = 10;
    scalingTargetCPU.value = 70;
  }
}

watch(activeInstance, syncScalingFromService, { immediate: true });

const scalingSummary = computed(() => {
  const svc = activeInstance.value;
  if (!svc) return 'Not deployed';
  if (svc.scaling?.autoscaling?.enabled) {
    return `${svc.scaling.autoscaling.minReplicas}–${svc.scaling.autoscaling.maxReplicas} replicas · autoscaling`;
  }
  const r = svc.scaling?.replicas || svc.replicas || 1;
  return `${r} replica${r !== 1 ? 's' : ''} · manual`;
});

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

async function handleSaveScaling() {
  const envName = activeEnvironment.value?.name;
  if (!envName) return;

  scalingSaving.value = true;
  try {
    const input: Record<string, unknown> = {
      projectId: props.projectId,
      environment: envName,
      service: props.service.name,
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
  } catch (e: unknown) {
    toast.error('Failed to update scaling', { description: errorMessage(e) });
  } finally {
    scalingSaving.value = false;
  }
}

async function handleRemoveService() {
  try {
    const res = await removeServiceMutate({
      projectId: props.projectId,
      environment: activeEnvironment.value?.name,
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
    <!-- General -->
    <section class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        General
      </h3>

      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <div class="rounded-lg bg-muted/60 p-1.5">
              <FrameworkIcon :framework="service.framework" :size="20" />
            </div>
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">{{ service.name }}</p>
              <p class="text-xs text-muted-foreground">
                {{ service.framework || 'Container' }}
              </p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-3 border-t px-4 py-3">
              <!-- Image -->
              <div v-if="service.image" class="space-y-1.5">
                <Label class="text-xs font-medium">Image</Label>
                <div class="group flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
                  <Container :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="min-w-0 flex-1 truncate font-mono text-sm">{{ service.image }}</span>
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
              <!-- Source Repo -->
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

              <!-- Context Path -->
              <div v-if="service.contextPath && service.contextPath !== '.'" class="space-y-1.5">
                <Label class="text-xs font-medium">Root Directory</Label>
                <div class="flex items-center gap-2 rounded-md border bg-muted/50 px-3 py-2">
                  <Code :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="truncate font-mono text-sm">{{ service.contextPath }}</span>
                </div>
              </div>

              <!-- Branch -->
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

      <Collapsible :default-open="!!service.customStartCommand">
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Play :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Custom Start Command</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ service.customStartCommand || service.startCommand || 'Not configured' }}
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
                Command that will be run to start new deployments. Overrides the image's default entrypoint.
              </p>
              <Input
                v-model="customStartCommand"
                :placeholder="service.startCommand || 'npm run start'"
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
              <div v-else class="space-y-2">
                <div class="flex items-center gap-2">
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
                <div class="flex items-center gap-1.5 pl-1 text-xs text-muted-foreground">
                  <ArrowRight :size="10" class="shrink-0" />
                  <span>
                    Routes to port
                    <span class="font-mono font-medium text-foreground">{{ service.port }}</span>
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
                  class="rounded-md border bg-muted/50 px-3 py-2.5"
                >
                  <div class="flex items-center gap-2">
                    <!-- Status icon -->
                    <CircleCheck
                      v-if="dnsStatus(domain.hostname) === 'VALID'"
                      :size="14"
                      class="shrink-0 text-green-500"
                    />
                    <CircleAlert
                      v-else-if="dnsStatus(domain.hostname) === 'MISCONFIGURED'"
                      :size="14"
                      class="shrink-0 text-orange-500"
                    />
                    <Loader2
                      v-else
                      :size="14"
                      class="shrink-0 animate-spin text-yellow-500"
                    />
                    <a
                      :href="domainUrl(domain.hostname)"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="min-w-0 flex-1 truncate font-mono text-sm hover:underline"
                    >
                      {{ domain.hostname }}
                    </a>
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
                  <!-- Port routing + status line -->
                  <div class="mt-1 flex items-center gap-1.5 pl-[22px] text-xs text-muted-foreground">
                    <ArrowRight :size="10" class="shrink-0" />
                    <span>
                      Port
                      <span class="font-mono font-medium text-foreground">{{ service.port }}</span>
                    </span>
                    <template v-if="dnsStatus(domain.hostname) !== 'VALID'">
                      <span class="text-muted-foreground/50">&middot;</span>
                      <span>Waiting for DNS</span>
                      <span class="text-muted-foreground/50">&middot;</span>
                      <button
                        class="font-medium text-primary hover:underline"
                        @click="showDnsRecords(domain.hostname)"
                      >
                        Show records
                      </button>
                    </template>
                  </div>
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
              <div v-if="internalDns" class="space-y-2">
                <div class="group flex items-center gap-2">
                  <div class="flex-1 overflow-x-auto rounded-md border bg-muted/50 px-3 py-2">
                    <span class="whitespace-nowrap font-mono text-xs">{{ internalDns }}:{{ service.port }}</span>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    @click="copyToClipboard(`${internalDns}:${service.port}`)"
                  >
                    <Copy :size="14" />
                  </Button>
                </div>
                <div class="flex items-center gap-1.5 pl-1 text-xs text-muted-foreground">
                  <ArrowRight :size="10" class="shrink-0" />
                  <span>
                    Listening on port
                    <span class="font-mono font-medium text-foreground">{{ service.port }}</span>
                  </span>
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
                    type="number"
                    v-model.number="scalingReplicas"
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
                      type="number"
                      v-model.number="scalingMinReplicas"
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
                      type="number"
                      v-model.number="scalingMaxReplicas"
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
                        type="number"
                        v-model.number="scalingTargetCPU"
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

              <!-- Save -->
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
          <div v-if="resourceTierLabel" class="flex items-center gap-3 px-4 py-3">
            <Scaling :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Plan</p>
              <p class="text-xs text-muted-foreground">Resource tier</p>
            </div>
            <span class="text-sm font-medium">{{ resourceTierLabel }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <Cpu :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">CPU</p>
              <p class="text-xs text-muted-foreground">Per instance</p>
            </div>
            <span class="font-mono text-sm font-medium">{{ formatCpu(resources.cpuMillicores) }}</span>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <MemoryStick :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Memory</p>
              <p class="text-xs text-muted-foreground">Per instance</p>
            </div>
            <span class="font-mono text-sm font-medium">{{ formatMemory(resources.memoryMB) }}</span>
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

    <!-- DNS Records Modal -->
    <Dialog v-model:open="dnsModalOpen">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Configure DNS Records</DialogTitle>
          <DialogDescription>
            Add the following DNS record to
            <strong class="font-mono">{{ dnsModalHostname }}</strong>
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
                  <div class="flex items-center gap-1.5">
                    <CircleCheck
                      v-if="dnsModalStatus === 'VALID'"
                      :size="12"
                      class="shrink-0 text-green-500"
                    />
                    <CircleAlert
                      v-else-if="dnsModalStatus === 'MISCONFIGURED'"
                      :size="12"
                      class="shrink-0 text-orange-500"
                    />
                    <Loader2
                      v-else
                      :size="12"
                      class="shrink-0 animate-spin text-yellow-500"
                    />
                    <Badge variant="outline" class="font-mono text-xs">CNAME</Badge>
                  </div>
                </td>
                <td class="px-3 py-2 font-mono text-xs">{{ dnsModalHostname }}</td>
                <td class="px-3 py-2">
                  <div class="flex items-center gap-1">
                    <span class="font-mono text-xs">{{ dnsModalTarget }}</span>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-6 w-6 shrink-0"
                      @click="copyToClipboard(dnsModalTarget)"
                    >
                      <Copy :size="12" />
                    </Button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Live status message -->
        <p
          v-if="dnsModalMessage"
          class="text-xs"
          :class="{
            'text-green-600': dnsModalStatus === 'VALID',
            'text-orange-500': dnsModalStatus === 'MISCONFIGURED',
            'text-muted-foreground': dnsModalStatus === 'PENDING' || dnsModalStatus === 'ERROR',
          }"
        >
          {{ dnsModalMessage }}
        </p>
        <p v-else class="text-xs text-muted-foreground">
          DNS changes can take up to 48 hours to propagate. The status will update automatically once the record is detected.
        </p>

        <DialogFooter>
          <Button variant="outline" @click="dnsModalOpen = false">
            Dismiss
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
