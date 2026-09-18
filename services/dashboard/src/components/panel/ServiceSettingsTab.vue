<script setup lang="ts">
import { ref, computed, watch, nextTick, type ComponentPublicInstance } from 'vue';
import { useMutation, useApolloClient } from '@vue/apollo-composable';
import {
  Trash2, Copy, X, Globe, Plus, Minus, RefreshCw,
  ChevronDown, Network, ExternalLink, Scaling, GitBranch, Play, Zap, ArrowRight,
  Cpu, MemoryStick, Leaf, ShieldCheck, Check, ChevronsUpDown, Activity, UserCog, ClipboardList,
} from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import { getDomain } from 'tldts';
import { graphql } from '@/gql';
import {
  type SetServiceScalingInput,
  type ResourcesInput,
  type HealthCheckInput,
  DnsRecordType,
  DnsStatus,
  EndpointType,
  Protocol,
  ResourceTier,
  TlsStatus,
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
        port
        type
        protocol
        redirectTo
        dns {
          status
          requiredRecords {
            type
            host
            value
          }
        }
      }
    }
  }
`);

const AddCustomDomainDocument = graphql(`
  mutation AddCustomDomain($service: ServiceID!, $hostname: String!, $redirectTo: String) {
    addCustomDomain(service: $service, hostname: $hostname, redirectTo: $redirectTo) {
      id
      endpoints {
        host
        port
        type
        protocol
        redirectTo
        dns {
          status
          requiredRecords {
            type
            host
            value
          }
        }
      }
    }
  }
`);

const RemoveDomainDocument = graphql(`
  mutation RemoveDomain($service: ServiceID!, $hostname: String!) {
    removeDomain(service: $service, hostname: $hostname) {
      id
      endpoints {
        host
        port
        type
        protocol
        redirectTo
        dns {
          status
          requiredRecords {
            type
            host
            value
          }
        }
      }
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

const SetServicePortDocument = graphql(`
  mutation SetServicePort($service: ServiceID!, $port: Int) {
    setServicePort(service: $service, port: $port) {
      id
      port
    }
  }
`);

const SetServiceHealthCheckDocument = graphql(`
  mutation SetServiceHealthCheck($service: ServiceID!, $healthCheck: HealthCheckInput) {
    setServiceHealthCheck(service: $service, healthCheck: $healthCheck) {
      id
      healthCheck {
        path
        port
        initialDelaySeconds
        periodSeconds
        timeoutSeconds
        failureThreshold
        startupFailureThreshold
      }
    }
  }
`);

const SetServiceBranchDocument = graphql(`
  mutation SetServiceBranch($service: ServiceID!, $branch: String) {
    setServiceBranch(service: $service, branch: $branch) {
      id
      branch
    }
  }
`);

const SetAutoDeployDocument = graphql(`
  mutation SetAutoDeploy($service: ServiceID!, $enabled: Boolean!) {
    setAutoDeploy(service: $service, enabled: $enabled) {
      id
      autoDeploy
    }
  }
`);

const SetCIDeployDocument = graphql(`
  mutation SetCIDeploy($service: ServiceID!, $enabled: Boolean!) {
    setCIDeploy(service: $service, enabled: $enabled) {
      id
      ciDeploy
    }
  }
`);

const RepositoryBranchesDocument = graphql(`
  query RepositoryBranches($repositoryUrl: String!) {
    repositoryBranches(repositoryUrl: $repositoryUrl)
  }
`);

const SetServiceResourcesDocument = graphql(`
  mutation SetServiceResources($service: ServiceID!, $resources: ResourcesInput!) {
    setServiceResources(service: $service, resources: $resources) {
      id
      resources {
        cpu
        memory
      }
    }
  }
`);

const SetServiceUserDocument = graphql(`
  mutation SetServiceUser($service: ServiceID!, $user: Int) {
    setServiceUser(service: $service, user: $user) {
      id
      user
    }
  }
`);
import { useEnvironment } from '@/composables/useEnvironment';
import type { Endpoint, Service } from '@/composables/useEnvironment';
import Spinner from '@/components/LoadingSpinner.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { toast, errorToast } from '@/components/ui/sonner';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Popover, PopoverTrigger, PopoverContent } from '@/components/ui/popover';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  AlertDialog,
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

const endpoints = computed(() => props.service.endpoints ?? []);
const platformEndpoint = computed(() => endpoints.value.find(e => e.type === EndpointType.Platform));
const customEndpoints = computed(() => endpoints.value.filter(e => e.type === EndpointType.Custom));
const internalEndpoint = computed(() => endpoints.value.find(e => e.type === EndpointType.Internal));

const resources = computed(() => props.service.resources ?? null);
const resourceTier = computed(() => activeEnvironment.value?.resourceTier ?? null);

function domainUrl(endpoint: Endpoint) {
  let url = `${endpoint.protocol}://${endpoint.host}`;

  if (endpoint.protocol === Protocol.Http && endpoint.port != 80) {
    url += `:${endpoint.port}`
  }

  if (endpoint.protocol === Protocol.Https && endpoint.port != 443) {
    url += `:${endpoint.port}`
  }

  return url
}

const { resolveClient } = useApolloClient();
const refreshingDomains = ref(false);

async function refreshDomains() {
  if (refreshingDomains.value) return;

  refreshingDomains.value = true;
  try {
    await resolveClient().refetchQueries({ include: ['Environment'] });
  } catch (e: unknown) {
    errorToast('Failed to refresh domain status', { description: errorMessage(e) });
  } finally {
    refreshingDomains.value = false;
  }
}

function isEndpointVerified(endpoint: Endpoint): boolean {
  return endpoint.dns.status === DnsStatus.Valid && endpoint.tls === TlsStatus.Active;
}

function dnsStatusColor(status: DnsStatus): string {
  switch (status) {
    case DnsStatus.Valid: return 'text-emerald-600 dark:text-emerald-400';
    case DnsStatus.Pending: return 'text-amber-600 dark:text-amber-400';
    case DnsStatus.Misconfigured: return 'text-orange-600 dark:text-orange-400';
    case DnsStatus.Error: return 'text-red-600 dark:text-red-400';
    default: return 'text-muted-foreground';
  }
}

function tlsStatusColor(status: TlsStatus): string {
  switch (status) {
    case TlsStatus.Active: return 'text-emerald-600 dark:text-emerald-400';
    case TlsStatus.Provisioning: return 'text-amber-600 dark:text-amber-400';
    case TlsStatus.Error: return 'text-red-600 dark:text-red-400';
    default: return 'text-muted-foreground';
  }
}

// Custom domain input
const addingDomain = ref(false);
const customDomainInput = ref('');
const redirectEnabled = ref(false);
const redirectTarget = ref('');
const recordsHost = ref<string | null>(null);

const hostnamePattern = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

function normalizeHostname(input: string): string {
  let h = input.trim().toLowerCase();
  h = h.replace(/^https?:\/\//, '');
  h = h.replace(/\/+$/, '');
  return h;
}

const normalizedHostname = computed(() => normalizeHostname(customDomainInput.value));

const isApexInput = computed(() => {
  const host = normalizedHostname.value;
  return hostnamePattern.test(host) && getDomain(host, { allowPrivateDomains: true }) === host;
});

const redirectTargets = computed(() =>
  endpoints.value.filter(e => e.type !== EndpointType.Internal && !e.redirectTo && e.host !== normalizedHostname.value),
);

const existingEndpoint = computed(() => endpoints.value.find(d => d.host === normalizedHostname.value) ?? null);

const requestedRedirect = computed(() => (redirectEnabled.value ? redirectTarget.value : ''));

const isDomainUpdate = computed(() => {
  const existing = existingEndpoint.value;
  return existing !== null && existing.type === EndpointType.Custom && (existing.redirectTo ?? '') !== requestedRedirect.value;
});

const hostnameError = computed(() => {
  const raw = customDomainInput.value.trim();
  if (!raw) return '';
  const hostname = normalizeHostname(raw);
  if (!hostnamePattern.test(hostname)) {
    return 'Enter a valid domain (e.g. www.example.com)';
  }
  if (existingEndpoint.value && !isDomainUpdate.value) {
    return 'This domain is already added';
  }
  return '';
});

const redirectSourceByTarget = computed(() => {
  const sources = new Map<string, string>();
  for (const endpoint of customEndpoints.value) {
    if (endpoint.redirectTo) sources.set(endpoint.redirectTo, endpoint.host);
  }
  return sources;
});

const canAddDomain = computed(() => {
  const raw = customDomainInput.value.trim();
  const redirectReady = !redirectEnabled.value || redirectTargets.value.some(e => e.host === redirectTarget.value);
  return raw.length > 0 && !hostnameError.value && redirectReady && !addingCustomDomain.value;
});

const recordsEndpoint = computed(() => customEndpoints.value.find(e => e.host === recordsHost.value) ?? null);

const domainInputRef = ref<ComponentPublicInstance | null>(null);

function openAddDomain() {
  customDomainInput.value = '';
  redirectEnabled.value = false;
  redirectTarget.value = '';
  addingDomain.value = true;
  nextTick(() => (domainInputRef.value?.$el as HTMLInputElement | undefined)?.focus());
}

function closeAddDomain() {
  addingDomain.value = false;
  customDomainInput.value = '';
  redirectEnabled.value = false;
  redirectTarget.value = '';
}

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
const removingHostname = ref<string | null>(null);
const domainToRemove = ref<string | null>(null);
const endpointToRemove = computed(() => customEndpoints.value.find(e => e.host === domainToRemove.value) ?? null);

const { mutate: setServicePortMutate, loading: portSaving } = useMutation(SetServicePortDocument);

const currentPort = computed(() => props.service.port);
const portInput = ref<number | undefined>(currentPort.value || undefined);

watch(currentPort, value => {
  portInput.value = value || undefined;
});

const portChanged = computed(() => (portInput.value || null) !== (currentPort.value || null));

async function handleSavePort() {
  const port = portInput.value && portInput.value > 0 ? portInput.value : null;

  try {
    const res = await setServicePortMutate({ service: props.service.id, port });

    if (res?.errors?.length) {
      errorToast('Failed to update port', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success(port ? `Port set to ${port}` : 'Port removed');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update port', { description: errorMessage(e) });
  }
}

// Health check
const { mutate: setHealthCheckMutate, loading: healthCheckSaving } = useMutation(SetServiceHealthCheckDocument);

const healthCheckEnabled = ref(false);
const hcPath = ref('');
const hcPort = ref<number | undefined>(undefined);
const hcInitialDelay = ref<number | undefined>(undefined);
const hcPeriod = ref<number | undefined>(undefined);
const hcTimeout = ref<number | undefined>(undefined);
const hcFailureThreshold = ref<number | undefined>(undefined);
const hcStartupFailureThreshold = ref<number | undefined>(undefined);

function syncHealthCheckFromService() {
  const hc = props.service.healthCheck;
  healthCheckEnabled.value = !!hc;
  hcPath.value = hc?.path ?? '';
  hcPort.value = hc && hc.port && hc.port !== props.service.port ? hc.port : undefined;
  hcInitialDelay.value = hc?.initialDelaySeconds || undefined;
  hcPeriod.value = hc?.periodSeconds || undefined;
  hcTimeout.value = hc?.timeoutSeconds || undefined;
  hcFailureThreshold.value = hc?.failureThreshold || undefined;
  hcStartupFailureThreshold.value = hc?.startupFailureThreshold || undefined;
}

watch(() => props.service.healthCheck, syncHealthCheckFromService, { immediate: true });

const healthCheckSummary = computed(() => {
  const hc = props.service.healthCheck;
  if (!hc) return 'TCP port check (default)';
  return `HTTP GET ${hc.path}`;
});

const startupBudgetSeconds = computed(() => {
  const threshold = hcStartupFailureThreshold.value || 0;
  const period = hcPeriod.value || 5;
  return threshold * period;
});

const startupBudgetLabel = computed(() => {
  const total = startupBudgetSeconds.value;
  if (total <= 0) return '';
  const minutes = Math.floor(total / 60);
  const seconds = total % 60;
  if (minutes && seconds) return `~${minutes}m ${seconds}s`;
  if (minutes) return `~${minutes}m`;
  return `~${seconds}s`;
});

const healthCheckPathError = computed(() => {
  if (!healthCheckEnabled.value) return '';
  const path = hcPath.value.trim();
  if (!path) return 'A path is required for an HTTP health check';
  if (!path.startsWith('/')) return 'Path must start with /';
  return '';
});

const canSaveHealthCheck = computed(
  () => !healthCheckSaving.value && (!healthCheckEnabled.value || !healthCheckPathError.value),
);

async function handleSaveHealthCheck() {
  let input: HealthCheckInput | null = null;

  if (healthCheckEnabled.value) {
    const path = hcPath.value.trim();
    if (!path || healthCheckPathError.value) return;

    input = {
      path,
      port: hcPort.value || null,
      initialDelaySeconds: hcInitialDelay.value ?? null,
      periodSeconds: hcPeriod.value || null,
      timeoutSeconds: hcTimeout.value || null,
      failureThreshold: hcFailureThreshold.value || null,
      startupFailureThreshold: hcStartupFailureThreshold.value ?? null,
    };
  }

  try {
    const res = await setHealthCheckMutate({ service: props.service.id, healthCheck: input });

    if (res?.errors?.length) {
      errorToast('Failed to update health check', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success(input ? 'Health check updated' : 'Health check reset to default');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update health check', { description: errorMessage(e) });
  }
}

// Source branch + auto-deploy
const { mutate: setServiceBranchMutate, loading: branchSaving } = useMutation(SetServiceBranchDocument);
const { mutate: setAutoDeployMutate } = useMutation(SetAutoDeployDocument);
const { mutate: setCIDeployMutate } = useMutation(SetCIDeployDocument);

const autoDeploySaving = ref(false);
const ciDeploySaving = ref(false);
const ciSnippetCopied = ref(false);

const ciWorkflowSnippet = computed(
  () => `name: Deploy
on:
  push:
    branches: [${props.service.branch || 'main'}]
permissions:
  id-token: write   # required
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v5
        with: { go-version: stable }
      - run: go install github.com/zeitlos/lucity/cli/cmd/lucity@latest
      - run: lucity deploy ${props.service.id.split('/').slice(1).join('/')} --ref "$GITHUB_SHA" --wait
        env:
          LUCITY_API_URL: ${window.location.origin}`,
);

async function copyCISnippet() {
  try {
    await navigator.clipboard.writeText(ciWorkflowSnippet.value);
    ciSnippetCopied.value = true;
    setTimeout(() => (ciSnippetCopied.value = false), 2000);
  } catch (e: unknown) {
    errorToast('Failed to copy', { description: errorMessage(e) });
  }
}

const branchPickerOpen = ref(false);
const branches = ref<string[]>([]);
const branchesLoading = ref(false);
const branchesLoaded = ref(false);

const currentBranch = computed(() => props.service.branch ?? '');
const currentBranchLabel = computed(() => props.service.branch || 'Default branch');

async function loadBranches() {
  if (branchesLoaded.value || branchesLoading.value || !props.service.sourceUrl) return;

  branchesLoading.value = true;

  try {
    const { data } = await resolveClient().query({
      query: RepositoryBranchesDocument,
      variables: { repositoryUrl: props.service.sourceUrl },
    });
    branches.value = data?.repositoryBranches ?? [];
    branchesLoaded.value = true;
  } catch (e: unknown) {
    errorToast('Failed to load branches', { description: errorMessage(e) });
  } finally {
    branchesLoading.value = false;
  }
}

function toggleBranchPicker(open: boolean) {
  branchPickerOpen.value = open;
  if (open) loadBranches();
}

async function chooseBranch(branch: string | null) {
  branchPickerOpen.value = false;

  const value = branch?.trim() || null;

  if (value === (props.service.branch ?? null)) return;

  try {
    const res = await setServiceBranchMutate({ service: props.service.id, branch: value });

    if (res?.errors?.length) {
      errorToast('Failed to update branch', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success(value ? `Tracking ${value}` : 'Tracking default branch');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update branch', { description: errorMessage(e) });
  }
}

async function handleToggleAutoDeploy(enabled: boolean) {
  autoDeploySaving.value = true;

  try {
    const res = await setAutoDeployMutate({ service: props.service.id, enabled });

    if (res?.errors?.length) {
      errorToast('Failed to update auto-deploy', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success(enabled ? 'Auto-deploy enabled' : 'Auto-deploy disabled');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update auto-deploy', { description: errorMessage(e) });
  } finally {
    autoDeploySaving.value = false;
  }
}

async function handleToggleCIDeploy(enabled: boolean) {
  ciDeploySaving.value = true;

  try {
    const res = await setCIDeployMutate({ service: props.service.id, enabled });

    if (res?.errors?.length) {
      errorToast('Failed to update CI deploys', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success(enabled ? 'CI deploys enabled' : 'CI deploys disabled');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update CI deploys', { description: errorMessage(e) });
  } finally {
    ciDeploySaving.value = false;
  }
}

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
  if (!hostname || !canAddDomain.value) return;

  try {
    const res = await addCustomDomainMutate({
      service: props.service.id,
      hostname,
      redirectTo: redirectEnabled.value ? redirectTarget.value : null,
    });

    if (res?.errors?.length) {
      errorToast('Failed to add custom domain', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    closeAddDomain();
    recordsHost.value = hostname;
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to add custom domain', { description: errorMessage(e) });
  }
}

async function handleRemoveDomain(hostname: string) {
  removingHostname.value = hostname;
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
    domainToRemove.value = null;
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to remove domain', { description: errorMessage(e) });
  } finally {
    removingHostname.value = null;
  }
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text);
  toast.success('Copied to clipboard');
}

// Source
const sourceRepoUrl = computed(() => {
  const url = props.service.sourceUrl;
  if (!url) return null;
  return url.startsWith('http') ? url : `https://${url}`;
});

const repoDisplay = computed(() =>
  props.service.sourceUrl.replace(/^https?:\/\//, '').replace(/\.git$/, ''),
);

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

// Compute (vertical scaling)
const cpuOptions = [
  { value: '250m', label: '0.25 vCPU' },
  { value: '500m', label: '0.5 vCPU' },
  { value: '1', label: '1 vCPU' },
  { value: '2', label: '2 vCPU' },
  { value: '4', label: '4 vCPU' },
];

const memoryOptions = [
  { value: '256Mi', label: '256 MB' },
  { value: '512Mi', label: '512 MB' },
  { value: '1Gi', label: '1 GB' },
  { value: '2Gi', label: '2 GB' },
  { value: '4Gi', label: '4 GB' },
  { value: '8Gi', label: '8 GB' },
  { value: '16Gi', label: '16 GB' },
];

const selectedCpu = ref('');
const selectedMemory = ref('');
const resourcesSaving = ref(false);

const { mutate: setResourcesMutate } = useMutation(SetServiceResourcesDocument);

watch(
  resources,
  value => {
    selectedCpu.value = value?.cpu ?? '';
    selectedMemory.value = value?.memory ?? '';
  },
  { immediate: true },
);

const resourcesChanged = computed(() =>
  !!selectedCpu.value
  && !!selectedMemory.value
  && (selectedCpu.value !== resources.value?.cpu || selectedMemory.value !== resources.value?.memory),
);

async function handleSaveResources() {
  resourcesSaving.value = true;
  try {
    const input: ResourcesInput = {
      cpu: selectedCpu.value,
      memory: selectedMemory.value,
    };

    const res = await setResourcesMutate({ service: props.service.id, resources: input });

    if (res?.errors?.length) {
      errorToast('Failed to update compute', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success('Compute updated');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update compute', { description: errorMessage(e) });
  } finally {
    resourcesSaving.value = false;
  }
}

// Run-as user (image-based services only)
const userInput = ref('');
const userSaving = ref(false);

const { mutate: setUserMutate } = useMutation(SetServiceUserDocument);

watch(
  () => props.service,
  s => {
    userInput.value = s.user != null ? String(s.user) : '';
  },
  { immediate: true },
);

const userChanged = computed(() => {
  const current = props.service.user != null ? String(props.service.user) : '';
  return userInput.value.trim() !== current;
});

async function handleSaveUser() {
  let user: number | null = null;

  if (userInput.value.trim() !== '') {
    user = Number(userInput.value);
    if (!Number.isInteger(user) || user < 0 || user > 65535) {
      errorToast('Invalid user id', { description: 'Must be a whole number between 0 and 65535' });
      return;
    }
  }

  userSaving.value = true;
  try {
    const res = await setUserMutate({ service: props.service.id, user });

    if (res?.errors?.length) {
      errorToast('Failed to update run-as user', { description: res.errors.map(e => e.message).join(', ') });
      return;
    }

    toast.success('Run-as user updated');
    emit('refetch');
  } catch (e: unknown) {
    errorToast('Failed to update run-as user', { description: errorMessage(e) });
  } finally {
    userSaving.value = false;
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
    <!-- Source -->
    <section v-if="service.sourceUrl" class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Source
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <!-- Repository (read-only) -->
          <div class="flex items-center gap-3 px-4 py-3">
            <GithubIcon :size="16" class="shrink-0" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-foreground">{{ repoDisplay }}</p>
              <p class="text-xs text-muted-foreground">
                Repository<span v-if="service.contextPath"> built from {{ service.contextPath }}</span>
              </p>
            </div>
            <a
              v-if="sourceRepoUrl"
              :href="sourceRepoUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground"
            >
              <ExternalLink :size="14" />
            </a>
          </div>

          <!-- Branch -->
          <div class="flex items-center gap-3 px-4 py-3">
            <GitBranch :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-foreground">{{ currentBranchLabel }}</p>
              <p class="text-xs text-muted-foreground">Tracked branch</p>
            </div>
            <Popover :open="branchPickerOpen" @update:open="toggleBranchPicker">
              <PopoverTrigger as-child>
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="branchSaving"
                  class="h-8"
                >
                  Change
                  <ChevronsUpDown :size="14" class="ml-1.5 opacity-50" />
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-64 p-0" align="end">
                <Command>
                  <CommandInput placeholder="Search branches..." />
                  <CommandList>
                    <CommandGroup>
                      <CommandItem value="Default branch" class="text-sm" @select="chooseBranch(null)">
                        <Check :size="14" :class="currentBranch === '' ? 'opacity-100' : 'opacity-0'" />
                        Default branch
                      </CommandItem>
                    </CommandGroup>
                    <div
                      v-if="branchesLoading"
                      class="flex items-center justify-center gap-2 py-4 text-xs text-muted-foreground"
                    >
                      <RefreshCw :size="14" class="animate-spin" />
                      Loading branches…
                    </div>
                    <template v-else>
                      <CommandEmpty>No branches found.</CommandEmpty>
                      <CommandGroup>
                        <CommandItem
                          v-for="branch in branches"
                          :key="branch"
                          :value="branch"
                          class="font-mono text-sm"
                          @select="chooseBranch(branch)"
                        >
                          <Check :size="14" :class="currentBranch === branch ? 'opacity-100' : 'opacity-0'" />
                          {{ branch }}
                        </CommandItem>
                      </CommandGroup>
                    </template>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>

          <!-- Auto-deploy -->
          <div class="flex items-center gap-3 px-4 py-3">
            <Zap
              :size="16"
              class="shrink-0"
              :class="service.autoDeploy ? 'text-violet-500' : 'text-muted-foreground'"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                Auto-Deploy {{ service.autoDeploy ? 'enabled' : 'disabled' }}
              </p>
              <p class="text-xs text-muted-foreground">Deploy on every push to the tracked branch</p>
            </div>
            <Switch
              :model-value="service.autoDeploy"
              :disabled="autoDeploySaving"
              class="data-[state=unchecked]:bg-border"
              @update:model-value="handleToggleAutoDeploy"
            />
          </div>

          <!-- CI deploys -->
          <div class="flex items-center gap-3 px-4 py-3">
            <ShieldCheck
              :size="16"
              class="shrink-0"
              :class="service.ciDeploy ? 'text-violet-500' : 'text-muted-foreground'"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">
                CI Deploys {{ service.ciDeploy ? 'enabled' : 'disabled' }}
              </p>
              <p class="text-xs text-muted-foreground">
                Let GitHub Actions from this repo deploy this service, keyless. No stored token.
              </p>
            </div>
            <Switch
              :model-value="service.ciDeploy"
              :disabled="ciDeploySaving"
              class="data-[state=unchecked]:bg-border"
              @update:model-value="handleToggleCIDeploy"
            />
          </div>

          <!-- CI workflow snippet -->
          <div v-if="service.ciDeploy" class="space-y-2 px-4 py-3">
            <div class="flex items-center justify-between">
              <p class="text-xs font-medium text-muted-foreground">
                Add to your workflow (needs <code class="font-mono">permissions: id-token: write</code>)
              </p>
              <Button variant="ghost" size="sm" class="h-7 gap-1.5" @click="copyCISnippet">
                <component :is="ciSnippetCopied ? Check : Copy" :size="13" />
                {{ ciSnippetCopied ? 'Copied' : 'Copy' }}
              </Button>
            </div>
            <pre class="overflow-x-auto rounded-md bg-muted/60 p-3 font-mono text-xs leading-relaxed text-foreground">{{ ciWorkflowSnippet }}</pre>
          </div>
        </div>
      </div>
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

      <!-- Listening port -->
      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Network :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Port</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ currentPort ? `Listening on ${currentPort}` : 'No port configured' }}
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
                The port your app listens on. Domains route to it. Leave empty if this service doesn't serve traffic.
              </p>
              <div class="flex items-center gap-2">
                <Input
                  v-model.number="portInput"
                  type="number"
                  placeholder="3000"
                  class="font-mono text-sm"
                  @keyup.enter="portChanged && handleSavePort()"
                />
                <Button
                  size="sm"
                  :disabled="!portChanged || portSaving"
                  @click="handleSavePort"
                >
                  {{ portSaving ? 'Saving...' : 'Save' }}
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>

      <!-- Health Check -->
      <Collapsible v-if="currentPort">
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Activity :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Health Check</p>
              <p class="truncate text-xs text-muted-foreground">{{ healthCheckSummary }}</p>
            </div>
            <ChevronDown
              :size="14"
              class="shrink-0 text-muted-foreground transition-transform duration-200 [[data-state=open]>&]:rotate-180"
            />
          </CollapsibleTrigger>
          <CollapsibleContent>
            <div class="space-y-4 border-t px-4 py-3">
              <p class="text-xs text-muted-foreground">
                How the platform decides an instance is ready to receive traffic. By default it
                checks that the port accepts connections. Switch to an HTTP check for apps that
                need to warm up before they can serve requests.
              </p>

              <!-- Enable HTTP check -->
              <div class="flex items-center justify-between">
                <div>
                  <Label class="text-sm font-medium">HTTP health check</Label>
                  <p class="text-xs text-muted-foreground">Probe an HTTP endpoint instead of the raw port.</p>
                </div>
                <Switch v-model="healthCheckEnabled" class="data-[state=unchecked]:bg-border" />
              </div>

              <template v-if="healthCheckEnabled">
                <!-- Path + port -->
                <div class="grid grid-cols-3 gap-3">
                  <div class="col-span-2 space-y-1.5">
                    <Label class="text-xs font-medium">Path</Label>
                    <Input
                      v-model="hcPath"
                      placeholder="/health/ready"
                      class="h-8 font-mono text-sm"
                      :class="{ 'border-destructive': healthCheckPathError }"
                    />
                  </div>
                  <div class="space-y-1.5">
                    <Label class="text-xs font-medium">Port</Label>
                    <Input
                      v-model.number="hcPort"
                      type="number"
                      :placeholder="String(currentPort)"
                      class="h-8 font-mono text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                  </div>
                </div>
                <div class="-mt-2 space-y-0.5 px-1">
                  <p v-if="healthCheckPathError" class="text-xs text-destructive">
                    {{ healthCheckPathError }}
                  </p>
                  <p class="text-[11px] text-muted-foreground">
                    Leave the port blank to use the service port ({{ currentPort }}).
                  </p>
                </div>

                <!-- Steady-state timing -->
                <div class="grid grid-cols-4 gap-2">
                  <div class="space-y-1.5">
                    <Label class="text-xs font-medium">Delay</Label>
                    <Input
                      v-model.number="hcInitialDelay"
                      type="number"
                      placeholder="0"
                      class="h-8 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                  </div>
                  <div class="space-y-1.5">
                    <Label class="text-xs font-medium">Period</Label>
                    <Input
                      v-model.number="hcPeriod"
                      type="number"
                      placeholder="5"
                      class="h-8 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                  </div>
                  <div class="space-y-1.5">
                    <Label class="text-xs font-medium">Timeout</Label>
                    <Input
                      v-model.number="hcTimeout"
                      type="number"
                      placeholder="3"
                      class="h-8 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                  </div>
                  <div class="space-y-1.5">
                    <Label class="text-xs font-medium">Failures</Label>
                    <Input
                      v-model.number="hcFailureThreshold"
                      type="number"
                      placeholder="3"
                      class="h-8 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                  </div>
                </div>
                <p class="-mt-2 px-1 text-[11px] text-muted-foreground">
                  All values in seconds, except failures. Left blank, sensible defaults apply.
                </p>

                <!-- Startup budget -->
                <div class="space-y-1.5">
                  <div class="flex items-center justify-between">
                    <Label class="text-xs font-medium">Startup budget</Label>
                    <span v-if="startupBudgetLabel" class="text-[11px] text-muted-foreground">
                      {{ startupBudgetLabel }} before traffic
                    </span>
                  </div>
                  <div class="flex items-center gap-2">
                    <Input
                      v-model.number="hcStartupFailureThreshold"
                      type="number"
                      placeholder="0"
                      class="h-8 w-24 text-center text-sm [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                    />
                    <p class="text-[11px] text-muted-foreground">
                      Extra failed checks tolerated only while the instance is starting. Lets a slow
                      warm-up finish without failing the steady-state check.
                    </p>
                  </div>
                </div>
              </template>

              <div class="flex justify-end">
                <Button
                  size="sm"
                  :disabled="!canSaveHealthCheck"
                  @click="handleSaveHealthCheck"
                >
                  {{ healthCheckSaving ? 'Saving...' : 'Save' }}
                </Button>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>

      <!-- Platform Domain -->
      <Collapsible default-open>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Globe :size="16" class="shrink-0" :class="platformEndpoint ? 'text-primary' : 'text-muted-foreground'" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Platform Domain</p>
              <p class="truncate text-xs text-muted-foreground">
                {{ platformEndpoint ? platformEndpoint.host : 'Not configured' }}
              </p>
            </div>
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
                  <Plus :size="14" class="mr-1" />
                  {{ generatingDomain ? 'Generating...' : 'Generate domain' }}
                </Button>
              </div>

              <div v-else class="flex items-center gap-2">
                <div class="flex min-w-0 flex-1 items-stretch overflow-hidden rounded-md border bg-muted/30">
                  <a
                    :href="domainUrl(platformEndpoint)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="flex min-w-0 flex-1 items-center gap-2 px-3 py-2 transition-colors hover:bg-muted/80"
                  >
                    <ExternalLink :size="14" class="shrink-0 text-muted-foreground" />
                    <span class="truncate font-mono text-sm hover:underline">{{ platformEndpoint.host }}</span>
                  </a>
                  <Button
                    variant="ghost"
                    class="h-auto shrink-0 rounded-none border-l px-2.5 text-muted-foreground hover:text-foreground"
                    title="Copy"
                    @click="copyToClipboard(platformEndpoint!.host)"
                  >
                    <Copy :size="14" />
                  </Button>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-8 w-8 shrink-0 text-destructive"
                  :disabled="removingHostname === platformEndpoint!.host"
                  @click="handleRemoveDomain(platformEndpoint!.host)"
                >
                  <Spinner v-if="removingHostname === platformEndpoint!.host" :size="14" />
                  <X v-else :size="14" />
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
                {{ customEndpoints.length }} domain{{ customEndpoints.length !== 1 ? 's' : '' }} configured
              </p>
            </div>
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
                  <div class="flex min-w-0 flex-1 items-stretch overflow-hidden rounded-md border bg-muted/30">
                    <a
                      :href="domainUrl(endpoint)"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="flex min-w-0 flex-1 items-center gap-2 px-3 py-2 transition-colors hover:bg-muted/80"
                    >
                      <ExternalLink :size="14" class="shrink-0 text-muted-foreground" />
                      <span class="truncate font-mono text-sm hover:underline">{{ endpoint.host }}</span>
                      <template v-if="endpoint.redirectTo">
                        <ArrowRight :size="12" class="shrink-0 text-muted-foreground" />
                        <span class="truncate font-mono text-sm text-muted-foreground">{{ endpoint.redirectTo }}</span>
                      </template>
                    </a>
                    <Button
                      variant="ghost"
                      class="h-auto shrink-0 rounded-none border-l px-2.5 text-muted-foreground hover:text-foreground"
                      title="Copy"
                      @click="copyToClipboard(endpoint.host)"
                    >
                      <Copy :size="14" />
                    </Button>
                  </div>
                  <div class="hidden shrink-0 items-center gap-3 text-[11px] text-muted-foreground sm:flex">
                    <span>DNS <span :class="['font-medium', dnsStatusColor(endpoint.dns.status)]">{{ endpoint.dns.status }}</span></span>
                    <span>TLS <span :class="['font-medium', tlsStatusColor(endpoint.tls)]">{{ endpoint.tls }}</span></span>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0"
                    title="DNS records"
                    @click="recordsHost = endpoint.host"
                  >
                    <ClipboardList :size="14" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 shrink-0 text-destructive"
                    :title="redirectSourceByTarget.has(endpoint.host) ? `Target of ${redirectSourceByTarget.get(endpoint.host)}` : 'Remove'"
                    :disabled="removingHostname === endpoint.host || redirectSourceByTarget.has(endpoint.host)"
                    @click="domainToRemove = endpoint.host"
                  >
                    <Spinner v-if="removingHostname === endpoint.host" :size="14" />
                    <X v-else :size="14" />
                  </Button>
                </div>
              </div>

              <div v-if="!addingDomain" class="flex justify-end">
                <Button size="sm" variant="outline" @click="openAddDomain">
                  <Plus :size="14" class="mr-1" />
                  Add custom domain
                </Button>
              </div>

              <div v-else class="space-y-3 rounded-md border p-3">
                <p class="text-sm font-medium text-foreground">Add custom domain</p>
                <div class="space-y-1.5">
                  <Input
                    ref="domainInputRef"
                    v-model="customDomainInput"
                    placeholder="www.example.com"
                    class="font-mono text-sm"
                    :class="{ 'border-destructive': hostnameError }"
                    @keyup.enter="canAddDomain && handleAddCustomDomain()"
                    @keyup.esc="closeAddDomain"
                  />
                  <p v-if="hostnameError" class="px-1 text-xs text-destructive">
                    {{ hostnameError }}
                  </p>
                  <p v-else-if="isApexInput && !redirectEnabled" class="px-1 text-xs text-muted-foreground">
                    An apex domain needs an ALIAS record at your DNS provider. If yours does not support that, redirect it to a subdomain instead.
                  </p>
                </div>
                <div class="flex flex-wrap items-center gap-3">
                  <label class="flex cursor-pointer items-center gap-2 text-sm">
                    <Checkbox v-model="redirectEnabled" />
                    Redirect to
                  </label>
                  <Select v-if="redirectEnabled" v-model="redirectTarget">
                    <SelectTrigger class="h-8 w-64 font-mono text-xs">
                      <SelectValue placeholder="Select a domain" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem
                        v-for="target in redirectTargets"
                        :key="target.host"
                        :value="target.host"
                        class="font-mono text-xs"
                      >
                        {{ target.host }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <p v-if="redirectEnabled && !redirectTargets.length" class="px-1 text-xs text-muted-foreground">
                  Add the domain to redirect to first.
                </p>
                <div class="flex justify-end gap-2">
                  <Button size="sm" variant="ghost" @click="closeAddDomain">Cancel</Button>
                  <Button size="sm" :disabled="!canAddDomain" @click="handleAddCustomDomain">
                    {{ addingCustomDomain ? 'Saving...' : isDomainUpdate ? 'Update domain' : 'Add domain' }}
                  </Button>
                </div>
              </div>
            </div>
          </CollapsibleContent>
        </div>
      </Collapsible>

      <Dialog :open="!!recordsHost" @update:open="(open) => { if (!open) recordsHost = null; }">
        <DialogContent class="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle class="font-mono">{{ recordsHost }}</DialogTitle>
            <DialogDescription>
              <template v-if="recordsEndpoint && isEndpointVerified(recordsEndpoint)">
                This domain is live. Keep these DNS records in place.
              </template>
              <template v-else>
                This domain goes live once these DNS records exist at your DNS provider. Changes can take a while to propagate.
              </template>
            </DialogDescription>
          </DialogHeader>

          <div v-if="recordsEndpoint" class="min-w-0 space-y-3">
            <div class="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
              <span>DNS <span :class="['font-medium', dnsStatusColor(recordsEndpoint.dns.status)]">{{ recordsEndpoint.dns.status }}</span></span>
              <span>TLS <span :class="['font-medium', tlsStatusColor(recordsEndpoint.tls)]">{{ recordsEndpoint.tls }}</span></span>
              <span v-if="recordsEndpoint.redirectTo" class="truncate">
                Redirects to <span class="font-mono">{{ recordsEndpoint.redirectTo }}</span>
              </span>
            </div>

            <div class="space-y-2">
              <div
                v-for="record in recordsEndpoint.dns.requiredRecords"
                :key="record.type + record.host + record.value"
                class="rounded-md border bg-muted/30 px-3 py-2 text-xs"
              >
                <p class="mb-1 font-mono font-semibold">{{ record.type }}</p>
                <div class="grid grid-cols-[3rem_minmax(0,1fr)_auto] items-center gap-x-2 gap-y-1">
                  <span class="text-muted-foreground">Name</span>
                  <span class="min-w-0 break-all font-mono">{{ record.host }}</span>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(record.host)">
                    <Copy :size="12" />
                  </Button>
                  <span class="text-muted-foreground">Value</span>
                  <span class="min-w-0 break-all font-mono">{{ record.value }}</span>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToClipboard(record.value)">
                    <Copy :size="12" />
                  </Button>
                </div>
              </div>
            </div>

            <p
              v-if="recordsEndpoint.dns.requiredRecords.some(r => r.type === DnsRecordType.Alias)"
              class="text-xs text-muted-foreground"
            >
              ALIAS, ANAME or CNAME flattening, depending on your DNS provider. Without that support, redirect the domain to a subdomain instead.
            </p>
          </div>
          <div v-else class="flex items-center justify-center py-6">
            <Spinner :size="16" />
          </div>

          <DialogFooter>
            <Button variant="outline" :disabled="refreshingDomains" @click="refreshDomains">
              <RefreshCw :size="14" class="mr-1" :class="{ 'animate-spin': refreshingDomains }" />
              {{ refreshingDomains ? 'Checking…' : 'Check again' }}
            </Button>
            <Button @click="recordsHost = null">Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog :open="!!domainToRemove">
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove domain</AlertDialogTitle>
            <AlertDialogDescription>
              Remove <strong class="font-mono">{{ domainToRemove }}</strong> from this service? This will also delete the TLS certificate.
              <template v-if="endpointToRemove?.redirectTo">
                It will no longer redirect to <strong class="font-mono">{{ endpointToRemove.redirectTo }}</strong>, which stays.
              </template>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel :disabled="!!removingHostname" @click="domainToRemove = null">Cancel</AlertDialogCancel>
            <Button
              variant="destructive"
              :disabled="!!removingHostname"
              @click="handleRemoveDomain(domainToRemove!)"
            >
              {{ removingHostname ? 'Removing...' : 'Remove' }}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <!-- Private Networking -->
      <Collapsible>
        <div class="overflow-hidden rounded-lg border">
          <CollapsibleTrigger class="flex w-full items-center gap-3 px-4 py-3 transition-colors hover:bg-muted/30">
            <Network :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1 text-left">
              <p class="text-sm font-medium text-foreground">Private Networking</p>
              <p class="max-w-55 truncate text-xs text-muted-foreground">
                {{ internalEndpoint?.host || 'Internal DNS' }}
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
              <div v-if="internalEndpoint" class="flex min-w-0 items-stretch overflow-hidden rounded-md border bg-muted/30">
                <div class="flex min-w-0 flex-1 items-center gap-2 px-3 py-2" :title="internalEndpoint.host">
                  <Network :size="14" class="shrink-0 text-muted-foreground" />
                  <span class="truncate font-mono text-sm">{{ internalEndpoint.host }}</span>
                </div>
                <Button
                  variant="ghost"
                  class="h-auto shrink-0 rounded-none border-l px-2.5 text-muted-foreground hover:text-foreground"
                  title="Copy"
                  @click="copyToClipboard(internalEndpoint.host)"
                >
                  <Copy :size="14" />
                </Button>
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
        <div class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <Cpu :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">CPU</p>
              <p class="text-xs text-muted-foreground">Limit per instance</p>
            </div>
            <Select v-model="selectedCpu">
              <SelectTrigger class="h-8 w-36">
                <SelectValue placeholder="Select" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in cpuOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div class="flex items-center gap-3 px-4 py-3">
            <MemoryStick :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">Memory</p>
              <p class="text-xs text-muted-foreground">Limit per instance</p>
            </div>
            <Select v-model="selectedMemory">
              <SelectTrigger class="h-8 w-36">
                <SelectValue placeholder="Select" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="option in memoryOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.label }}
                </SelectItem>
              </SelectContent>
            </Select>
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
          <div class="flex items-center justify-between gap-3 px-4 py-3">
            <p class="text-xs text-muted-foreground">
              Saving restarts running instances.
            </p>
            <Button
              size="sm"
              :disabled="!resourcesChanged || resourcesSaving"
              @click="handleSaveResources"
            >
              {{ resourcesSaving ? 'Saving...' : 'Save' }}
            </Button>
          </div>
        </div>
      </div>
    </section>

    <!-- Run as user (image-based services only) -->
    <section v-if="!service.sourceUrl" class="space-y-2">
      <h3 class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Run as user
      </h3>

      <div class="overflow-hidden rounded-lg border">
        <div class="divide-y">
          <div class="flex items-center gap-3 px-4 py-3">
            <UserCog :size="16" class="shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">User id</p>
              <p class="text-xs text-muted-foreground">
                Runs the container as this user. Mounted volumes are owned by the same id. Leave empty to use the image
                default.
              </p>
            </div>
            <Input v-model="userInput" type="number" min="0" max="65535" placeholder="image default" class="h-8 w-36" />
          </div>
          <div class="flex items-center justify-between gap-3 px-4 py-3">
            <p class="text-xs text-muted-foreground">
              Saving restarts running instances.
            </p>
            <Button
              size="sm"
              :disabled="!userChanged || userSaving"
              @click="handleSaveUser"
            >
              {{ userSaving ? 'Saving...' : 'Save' }}
            </Button>
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
                  <AlertDialogCancel :disabled="removing">Cancel</AlertDialogCancel>
                  <Button
                    variant="destructive"
                    :disabled="removing"
                    @click="handleRemoveService"
                  >
                    {{ removing ? 'Removing...' : 'Remove' }}
                  </Button>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
