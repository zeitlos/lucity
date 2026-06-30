<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable';
import { FolderPlus, FolderGit2, Plus, Lock, Globe, ArrowLeft, Search, ChevronDown, ChevronRight, Container, Star, Award, Loader2, HardDrive, Braces, CornerDownLeft, Check } from '@lucide/vue';
import type { Component } from 'vue';
import BucketIcon from '@/components/BucketIcon.vue';
import FrameworkIcon from '@/components/FrameworkIcon.vue';
import GithubIcon from '@/components/GithubIcon.vue';
import CommandFlow from '@/components/CommandFlow.vue';
import type { CommandFlowConfig } from '@/lib/commandFlow';
import { onKeyStroke, refDebounced } from '@vueuse/core';
import { graphql } from '@/gql';
import { GitHubAccountType } from '@/gql/graphql';

const GitHubConnectedDocument = graphql(`
  query GitHubConnected {
    githubConnected
  }
`);

const GitHubSourcesDocument = graphql(`
  query GitHubSources {
    githubSources {
      accountLogin
      accountAvatarUrl
      accountType
    }
  }
`);

const GitHubRepositoriesDocument = graphql(`
  query GitHubRepositories($account: String!) {
    githubRepositories(account: $account) {
      id
      name
      fullName
      htmlUrl
      defaultBranch
      private
    }
  }
`);

const CreateProjectDocument = graphql(`
  mutation CreateProject($input: CreateProjectInput!) {
    createProject(input: $input) {
      id
      name
      environments {
        id
        name
      }
    }
  }
`);

const AddServiceDocument = graphql(`
  mutation AddService($environmentId: EnvironmentID!, $input: AddServiceInput!) {
    addService(environment: $environmentId, input: $input) {
      id
      name
    }
  }
`);

const DetectServicesDocument = graphql(`
  query DetectServices($repositoryUrl: String!) {
    detectServices(repositoryUrl: $repositoryUrl) {
      name
      language
      framework
      contextPath
      startCommand
      suggestedPort
    }
  }
`);

const SearchImagesDocument = graphql(`
  query SearchImages($query: String!) {
    searchImages(query: $query) {
      name
      description
      starCount
      pullCount
      official
    }
  }
`);

const CreateDatabaseDocument = graphql(`
  mutation CreateDatabase($input: CreateDatabaseInput!) {
    createDatabase(input: $input) {
      id
      name
      version
      size
    }
  }
`);

const CreateKeyValueStoreDocument = graphql(`
  mutation CreateKeyValueStore($input: CreateKeyValueStoreInput!) {
    createKeyValueStore(input: $input) {
      id
      name
      version
      size
    }
  }
`);

const CreateBucketDocument = graphql(`
  mutation CreateBucket($input: CreateBucketInput!) {
    createBucket(input: $input) {
      id
      name
      region
      endpoint
    }
  }
`);

const CreateVolumeDocument = graphql(`
  mutation CreateVolume($environment: EnvironmentID!, $name: String!, $size: String!) {
    createVolume(environment: $environment, name: $name, size: $size) {
      id
      name
      size
    }
  }
`);
import { useEnvironment } from '@/composables/useEnvironment';
import { useGitHubInstall } from '@/composables/useGitHubInstall';
import { errorToast } from '@/components/ui/sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { errorMessage } from '@/lib/utils';
import { isValidSlug, deriveServiceName, isValidServiceName } from '@/lib/slug';
import NameSlugField from '@/components/NameSlugField.vue';

const props = defineProps<{
  open: boolean;
  context: 'projects' | 'environment';
  environmentId?: string;
  initialView?: 'main' | 'github-repos';
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'created'): void;
}>();

const router = useRouter();
const { resolveClient } = useApolloClient();
const { activeEnvironment } = useEnvironment();

// Drill-down state
type PaletteView = 'main' | 'github-repos' | 'select-services' | 'service-wizard' | 'manual-service' | 'database' | 'keyValueStore' | 'bucket' | 'volume' | 'container-image' | 'name-project';
const view = ref<PaletteView>('main');
const search = ref('');
const inputRef = ref<HTMLInputElement>();
const focusedIndex = ref(0);

// Source picker state
const selectedSource = ref<(typeof sources)['value'][number] | null>(null);
const sourcePickerOpen = ref(false);

// Project naming state
const projectDisplayName = ref('');
const projectSlug = ref('');
const nameSlugRef = ref<InstanceType<typeof NameSlugField> | null>(null);
const pendingRepo = ref<{ fullName: string; htmlUrl: string } | null>(null);
const pendingImage = ref<string | null>(null);
const processingItemId = ref<string | null>(null);

const isProjectValid = computed(() =>
  projectDisplayName.value.trim().length > 0 && isValidSlug(projectSlug.value),
);

type DetectedService = {
  name: string;
  language: string;
  framework: string;
  contextPath: string;
  startCommand: string;
};
const detectedServices = ref<DetectedService[]>([]);
const selectedDetectedIndex = ref(0);
const activeDetected = computed(() => detectedServices.value[selectedDetectedIndex.value] ?? null);

type ServiceStep = 'name' | 'context' | 'variables';
const serviceStep = ref<ServiceStep>('name');
const serviceName = ref('');
const serviceContextPath = ref('/');
const serviceEnvVars = ref<{ key: string; value: string }[]>([]);
const envVarDraft = ref('');

const confirmEnvironmentId = ref<string | null>(null);
const confirmRepo = ref<{ fullName: string; htmlUrl: string } | null>(null);

const isServiceValid = computed(() => isValidServiceName(serviceName.value));
const serviceDetected = computed(() => activeDetected.value !== null);

const wizardInput = computed<string>({
  get() {
    switch (serviceStep.value) {
      case 'name': return serviceName.value;
      case 'context': return serviceContextPath.value;
      case 'variables': return envVarDraft.value;
    }
    return '';
  },
  set(value: string) {
    switch (serviceStep.value) {
      case 'name': serviceName.value = value; break;
      case 'context': serviceContextPath.value = value; break;
      case 'variables': envVarDraft.value = value; break;
    }
  },
});

const wizardPlaceholder = computed(() => {
  switch (serviceStep.value) {
    case 'name': return 'Service name';
    case 'context': return 'Root directory, e.g. /';
    case 'variables': return 'KEY=value';
  }
  return '';
});

// Selectable command rows shown below the input. Navigated with arrows, the
// focused row is activated with Enter (green ↵ badge) or clicked.
type WizardItem = { id: string; label: string; icon: Component; actionLabel: string; disabled?: boolean; mono?: boolean; action: () => void };

const wizardItems = computed<WizardItem[]>(() => {
  const createItem: WizardItem = {
    id: 'create',
    label: creating.value || addingService.value ? 'Creating service...' : 'Create service',
    icon: Check,
    actionLabel: 'create',
    disabled: !isServiceValid.value || creating.value || addingService.value,
    action: () => { handleCreateService(); },
  };
  const variablesItem: WizardItem = {
    id: 'add-variables',
    label: serviceEnvVars.value.length > 0 ? `Variables (${serviceEnvVars.value.length})` : 'Add variables',
    icon: Braces,
    actionLabel: 'open',
    action: openVariablesStep,
  };

  if (serviceStep.value === 'name' && !serviceDetected.value) {
    return [{
      id: 'continue',
      label: 'Continue',
      icon: ChevronRight,
      actionLabel: 'next',
      disabled: !isServiceValid.value,
      action: goToContextStep,
    }];
  }

  if (serviceStep.value === 'variables') {
    const draft = envVarDraft.value.trim();
    const varRows: WizardItem[] = serviceEnvVars.value.map((row, index) => ({
      id: `var-${row.key}`,
      label: row.value ? `${row.key}=${row.value}` : row.key,
      icon: Braces,
      mono: true,
      actionLabel: 'remove',
      action: () => removeEnvVarRow(index),
    }));
    return [
      ...varRows,
      {
        id: 'add-variable',
        label: draft ? `Add ${draft}` : 'Add variable',
        icon: Plus,
        actionLabel: 'add',
        action: addEnvVarFromDraft,
      },
      createItem,
    ];
  }

  return [variablesItem, createItem];
});

const wizardPrimaryIndex = computed(() => {
  const index = wizardItems.value.findIndex(i => i.id === 'create' || i.id === 'continue');
  return index >= 0 ? index : 0;
});

// Reset when palette opens
watch(() => props.open, (open) => {
  if (open) {
    view.value = props.initialView || 'main';
    search.value = '';
    sourcePickerOpen.value = false;
    containerImageRef.value = '';
    focusedIndex.value = 0;
    projectDisplayName.value = '';
    projectSlug.value = '';
    pendingRepo.value = null;
    pendingImage.value = null;
    processingItemId.value = null;
    volumeStep.value = 'name';
    detectedServices.value = [];
    selectedDetectedIndex.value = 0;
    serviceStep.value = 'name';
    serviceName.value = '';
    serviceContextPath.value = '/';
    serviceEnvVars.value = [];
    envVarDraft.value = '';
    confirmEnvironmentId.value = null;
    confirmRepo.value = null;
    selectedSource.value = sources.value[0] ?? null;
    nextTick(() => inputRef.value?.focus());
  }
});

watch(view, (newView) => {
  search.value = '';
  focusedIndex.value = newView === 'service-wizard' ? wizardPrimaryIndex.value : 0;
  nextTick(() => inputRef.value?.focus());
});

// Stepping through the wizard keeps the input focused and the primary command
// row (Create / Continue) pre-selected so Enter does the expected thing.
watch(serviceStep, () => {
  focusedIndex.value = wizardPrimaryIndex.value;
  nextTick(() => inputRef.value?.focus());
});

onKeyStroke('Escape', () => {
  if (!props.open) return;
  if (sourcePickerOpen.value) {
    sourcePickerOpen.value = false;
  } else if (view.value === 'service-wizard') {
    wizardBack();
  } else if (view.value === 'select-services') {
    selectServicesBack();
  } else if (view.value === 'volume' && volumeStep.value === 'size') {
    volumeStep.value = 'name';
  } else if (view.value !== 'main') {
    view.value = 'main';
  } else {
    close();
  }
});

function close() {
  emit('update:open', false);
}

// GitHub connected check
const { result: connectedResult, loading: connectedLoading } = useQuery(GitHubConnectedDocument, null, () => ({
  enabled: props.open && view.value === 'github-repos',
}));

const githubConnected = computed(() => connectedResult.value?.githubConnected ?? false);

// GitHub sources (installations)
const { result: sourcesResult, loading: sourcesLoading } = useQuery(GitHubSourcesDocument, null, () => ({
  enabled: props.open && view.value === 'github-repos' && githubConnected.value,
}));

const sources = computed(() => sourcesResult.value?.githubSources ?? []);

watch(sources, (s) => {
  if (s.length > 0 && !selectedSource.value) {
    selectedSource.value = s[0] ?? null;
  }
});

// GitHub repos for selected source
const { result: reposResult, loading: reposLoading } = useQuery(GitHubRepositoriesDocument, () => ({
  account: selectedSource.value?.accountLogin ?? '',
}), () => ({
  enabled: props.open && view.value === 'github-repos' && !!selectedSource.value,
}));

const repos = computed(() => {
  const all = reposResult.value?.githubRepositories ?? [];
  if (!search.value) return all;
  const q = search.value.toLowerCase();
  return all.filter((r: { fullName: string }) => r.fullName.toLowerCase().includes(q));
});

// Create project
const { mutate: createProject, loading: creating } = useMutation(CreateProjectDocument);

const detectingServices = ref(false);

async function handleSelectRepo(repo: { fullName: string; htmlUrl: string }) {
  if (creating.value || detectingServices.value) return;
  if (props.context === 'projects') {
    showProjectNaming(repo);
  } else {
    if (!props.environmentId) return;
    confirmEnvironmentId.value = props.environmentId;
    await detectAndConfirm(repo);
  }
}

function showProjectNaming(repo: { fullName: string; htmlUrl: string }) {
  pendingRepo.value = repo;
  pendingImage.value = null;
  const repoShortName = repo.fullName.split('/').pop() || '';
  projectDisplayName.value = repoShortName;
  projectSlug.value = '';
  view.value = 'name-project';
  nextTick(() => nameSlugRef.value?.focusName());
}

// Submit handler for the name-project view. Repo-backed projects defer creation
// until the service is confirmed (so an abandoned flow leaves no empty project);
// empty and image-backed projects are created immediately.
async function handleProjectNamingNext() {
  if (!isProjectValid.value || creating.value || detectingServices.value) return;
  if (pendingRepo.value) {
    confirmEnvironmentId.value = null;
    await detectAndConfirm(pendingRepo.value);
  } else {
    await finishEmptyOrImageProject();
  }
}

async function finishEmptyOrImageProject() {
  if (creating.value) return;

  try {
    const res = await createProject({
      input: {
        name: projectDisplayName.value.trim(),
        id: projectSlug.value,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const project = res?.data?.createProject;
    if (!project) return;

    const targetEnvId = project.environments?.[0]?.id;

    if (targetEnvId && pendingImage.value) {
      await addImageService(targetEnvId, pendingImage.value);
    }

    close();
    if (targetEnvId) {
      router.push({ name: 'environment', params: { projectId: project.id, environmentId: targetEnvId } });
    }
  } catch (e: unknown) {
    errorToast('Failed to create project', { description: errorMessage(e) });
  }
}

// Detect services in a repo, then route to the confirm/overwrite step.
async function detectAndConfirm(repo: { fullName: string; htmlUrl: string }) {
  confirmRepo.value = repo;
  processingItemId.value = repo.fullName;
  detectingServices.value = true;

  try {
    const client = resolveClient();
    const { data } = await client.query({
      query: DetectServicesDocument,
      variables: { repositoryUrl: repo.htmlUrl },
    });
    detectedServices.value = data?.detectServices ?? [];
    proceedAfterDetection(repo);
  } catch (e: unknown) {
    errorToast('Failed to detect services', { description: errorMessage(e) });
  } finally {
    detectingServices.value = false;
    processingItemId.value = null;
  }
}

function proceedAfterDetection(repo: { fullName: string; htmlUrl: string }) {
  const detected = detectedServices.value;

  if (detected.length === 0) {
    // Nothing detected — fall through to manual setup (name + root directory).
    startManualService(repo);
    return;
  }

  if (detected.length > 1) {
    view.value = 'select-services';
    return;
  }

  selectDetectedService(0);
}

function enterServiceWizard() {
  serviceEnvVars.value = [];
  envVarDraft.value = '';
  serviceStep.value = 'name';
  view.value = 'service-wizard';
  nextTick(() => inputRef.value?.focus());
}

function selectDetectedService(index: number) {
  selectedDetectedIndex.value = index;
  const detected = detectedServices.value[index];
  if (!detected) return;
  serviceName.value = deriveServiceName(detected.name);
  serviceContextPath.value = detected.contextPath || '/';
  enterServiceWizard();
}

function startManualService(repo: { fullName: string; htmlUrl: string }) {
  detectedServices.value = [];
  selectedDetectedIndex.value = 0;
  serviceName.value = deriveServiceName(repo.fullName.split('/').pop() || '');
  serviceContextPath.value = '/';
  enterServiceWizard();
}

function addEnvVarFromDraft() {
  const raw = envVarDraft.value.trim();
  if (!raw) return;
  const eq = raw.indexOf('=');
  const key = (eq === -1 ? raw : raw.slice(0, eq)).trim();
  const value = eq === -1 ? '' : raw.slice(eq + 1).trim();
  if (!key) return;
  const existing = serviceEnvVars.value.findIndex(v => v.key === key);
  if (existing >= 0) {
    serviceEnvVars.value[existing]!.value = value;
  } else {
    serviceEnvVars.value.push({ key, value });
  }
  envVarDraft.value = '';
  focusedIndex.value = wizardPrimaryIndex.value;
}

function removeEnvVarRow(index: number) {
  serviceEnvVars.value.splice(index, 1);
  focusedIndex.value = Math.min(focusedIndex.value, wizardItems.value.length - 1);
}

// Enter on the wizard: while typing a KEY=value draft it adds the variable,
// otherwise it activates the focused command row.
function wizardEnter() {
  if (creating.value || addingService.value || detectingServices.value) return;
  if (serviceStep.value === 'variables' && envVarDraft.value.trim()) {
    addEnvVarFromDraft();
    return;
  }
  const item = wizardItems.value[focusedIndex.value];
  if (item && !item.disabled) item.action();
}

function goToContextStep() {
  if (!isServiceValid.value) return;
  serviceStep.value = 'context';
}

function openVariablesStep() {
  serviceStep.value = 'variables';
}

function wizardBack() {
  if (serviceStep.value === 'variables') {
    serviceStep.value = serviceDetected.value ? 'name' : 'context';
    nextTick(() => inputRef.value?.focus());
  } else if (serviceStep.value === 'context') {
    serviceStep.value = 'name';
    nextTick(() => inputRef.value?.focus());
  } else if (detectedServices.value.length > 1) {
    view.value = 'select-services';
  } else {
    view.value = props.context === 'projects' ? 'name-project' : 'github-repos';
  }
}

function selectServicesBack() {
  view.value = props.context === 'projects' ? 'name-project' : 'github-repos';
}

// Create the confirmed service. In the project flow the project is created
// here (lazily) so the confirm step is the single point of no return.
async function handleCreateService() {
  if (!isServiceValid.value || creating.value || addingService.value || detectingServices.value) return;
  if (!confirmRepo.value) return;

  try {
    let environmentId = confirmEnvironmentId.value;
    let navigate: { projectId: string; environmentId: string } | null = null;

    if (props.context === 'projects' && !environmentId) {
      const res = await createProject({
        input: {
          name: projectDisplayName.value.trim(),
          id: projectSlug.value,
        },
      });

      if (res?.errors?.length) {
        errorToast('Failed to create project', {
          description: res.errors.map(e => e.message).join(', '),
        });
        return;
      }

      const project = res?.data?.createProject;
      const envId = project?.environments?.[0]?.id;
      if (!project || !envId) return;

      environmentId = envId;
      navigate = { projectId: project.id, environmentId: envId };
    }

    if (!environmentId) return;

    addEnvVarFromDraft();
    const variables = serviceEnvVars.value
      .map(v => ({ key: v.key.trim(), value: v.value }))
      .filter(v => v.key.length > 0);

    const res = await addServiceMutate({
      environmentId,
      input: {
        name: serviceName.value,
        repository: confirmRepo.value.fullName,
        contextPath: serviceContextPath.value || undefined,
        variables: variables.length > 0 ? variables : undefined,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    if (navigate) {
      router.push({ name: 'environment', params: navigate });
    } else {
      emit('created');
    }
  } catch (e: unknown) {
    errorToast('Failed to create service', { description: errorMessage(e) });
  }
}

// Add service (within environment context)
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceDocument);

const newServiceName = ref('');
const newServicePort = ref<number | null>(null);

// Create database (within environment context) — first flow on the generic
// config-driven CommandFlow renderer. Other create flows still use bespoke
// views below until they are migrated.
const databaseFlow = computed<CommandFlowConfig>(() => ({
  title: 'Database',
  iconSrc: 'https://devicons.railway.com/i/postgresql.svg',
  submitLabel: 'Create database',
  mutation: CreateDatabaseDocument,
  variables: (values) => ({ input: { environment: props.environmentId, name: values.name } }),
  steps: [
    {
      id: 'name',
      placeholder: 'Database name',
      validate: isValidServiceName,
      invalidHint: 'Lowercase letters, digits and hyphens only (2–16 characters).',
    },
  ],
}));

function handleFlowCreated() {
  emit('created');
  close();
}

// Create Redis store (within environment context)
const { mutate: createKeyValueStoreMutate, loading: creatingKeyValueStore } = useMutation(CreateKeyValueStoreDocument);
const newKeyValueStoreName = ref('');

async function handleCreateKeyValueStore() {
  if (!props.environmentId) return;

  try {
    const res = await createKeyValueStoreMutate({
      input: {
        environment: props.environmentId,
        name: newKeyValueStoreName.value,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create Redis store', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create Redis store', { description: errorMessage(e) });
  }
}

// Create object storage bucket (within environment context)
const { mutate: createBucketMutate, loading: creatingBucket } = useMutation(CreateBucketDocument);
const newBucketName = ref('');

async function handleCreateBucket() {
  if (!props.environmentId) return;

  try {
    const res = await createBucketMutate({
      input: {
        environment: props.environmentId,
        name: newBucketName.value,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create bucket', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create bucket', { description: errorMessage(e) });
  }
}

// Create volume (within environment context)
const { mutate: createVolumeMutate, loading: creatingVolume } = useMutation(CreateVolumeDocument);
const newVolumeName = ref('');
const volumeSizes = ['10Gi', '16Gi', '32Gi', '64Gi', '128Gi', '256Gi', '512Gi', '1Ti'];
const newVolumeSize = ref('10Gi');
const volumeStep = ref<'name' | 'size'>('name');

async function handleCreateVolume() {
  if (!props.environmentId) return;

  try {
    const res = await createVolumeMutate({
      environment: props.environmentId,
      name: newVolumeName.value,
      size: newVolumeSize.value,
    });

    if (res?.errors?.length) {
      errorToast('Failed to create volume', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create volume', { description: errorMessage(e) });
  }
}

async function handleAddManualService() {
  if (!props.environmentId) return;

  try {
    const res = await addServiceMutate({
      environmentId: props.environmentId,
      input: {
        name: newServiceName.value,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to add service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to add service', { description: errorMessage(e) });
  }
}

// Open GitHub App install in a popup, refetch sources on completion
const { openInstallPopup } = useGitHubInstall(() => {
  const client = resolveClient();
  client.refetchQueries({ include: [GitHubSourcesDocument] });
});

// Container image state
const containerImageRef = ref('');
const containerImageDebounced = refDebounced(containerImageRef, 300);

const shouldSearchImages = computed(() => {
  const q = containerImageDebounced.value;
  if (!q) return false;
  if (q.includes('.')) return false;
  return true;
});

const { result: imageSearchResult, loading: searchingImages } = useQuery(SearchImagesDocument, () => ({
  query: containerImageDebounced.value,
}), () => ({
  enabled: props.open && view.value === 'container-image' && shouldSearchImages.value,
}));

const imageResults = computed(() => imageSearchResult.value?.searchImages ?? []);

function formatPullCount(count: number): string {
  if (count >= 1_000_000_000) return `${(count / 1_000_000_000).toFixed(1)}B`;
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return String(count);
}

async function handleSelectImage(imageRef: string) {
  if (!imageRef || creating.value || addingService.value) return;
  containerImageRef.value = imageRef;

  if (props.context === 'projects') {
    pendingImage.value = imageRef;
    pendingRepo.value = null;
    const imageName = imageRef.split('/').pop() || imageRef;
    projectDisplayName.value = imageName;
    projectSlug.value = '';
    view.value = 'name-project';
    nextTick(() => nameSlugRef.value?.focusName());
    return;
  } else {
    if (!props.environmentId) return;
    processingItemId.value = imageRef;
    try {
      await addImageService(props.environmentId, imageRef);
      close();
      emit('created');
    } catch (e: unknown) {
      errorToast('Failed to add service', { description: errorMessage(e) });
    } finally {
      processingItemId.value = null;
    }
  }
}

async function addImageService(environmentId: string, imageRef: string) {
  const res = await addServiceMutate({
    environmentId: environmentId,
    input: {
      image: imageRef,
    },
  });

  if (res?.errors?.length) {
    errorToast('Failed to add service', {
      description: res.errors.map(e => e.message).join(', '),
    });
    return;
  }
}

watch([search, imageResults, repos], () => {
  focusedIndex.value = 0;
});

watch(sourcePickerOpen, (isOpen) => {
  if (isOpen) focusedIndex.value = 0;
});

const currentItemCount = computed(() => {
  if (sourcePickerOpen.value) return sources.value.length + 1;
  switch (view.value) {
    case 'main': return mainItems.value.length;
    case 'github-repos': return repos.value.length;
    case 'container-image': return imageResults.value.length;
    case 'select-services': return detectedServices.value.length;
    case 'service-wizard': return wizardItems.value.length;
    default: return 0;
  }
});

function scrollFocusedIntoView() {
  document.querySelector('[data-focused="true"]')?.scrollIntoView({ block: 'nearest' });
}

onKeyStroke('ArrowDown', (e) => {
  if (!props.open) return;
  if (view.value === 'manual-service' || view.value === 'database' || view.value === 'keyValueStore' || view.value === 'volume') return;
  if (currentItemCount.value === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value + 1) % currentItemCount.value;
  nextTick(() => scrollFocusedIntoView());
});

onKeyStroke('ArrowUp', (e) => {
  if (!props.open) return;
  if (view.value === 'manual-service' || view.value === 'database' || view.value === 'keyValueStore' || view.value === 'volume') return;
  if (currentItemCount.value === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value - 1 + currentItemCount.value) % currentItemCount.value;
  nextTick(() => scrollFocusedIntoView());
});

onKeyStroke('Enter', (e) => {
  if (!props.open) return;

  if (sourcePickerOpen.value) {
    e.preventDefault();
    if (focusedIndex.value < sources.value.length) {
      selectedSource.value = sources.value[focusedIndex.value] ?? null;
    } else {
      openInstallPopup();
    }
    sourcePickerOpen.value = false;
    focusedIndex.value = 0;
    return;
  }

  switch (view.value) {
    case 'main':
      if (mainItems.value.length > 0 && focusedIndex.value < mainItems.value.length) {
        e.preventDefault();
        mainItems.value[focusedIndex.value]?.action();
      }
      break;
    case 'github-repos': {
      const repo = repos.value[focusedIndex.value];
      if (repo && !creating.value && !detectingServices.value) {
        e.preventDefault();
        handleSelectRepo(repo);
      }
      break;
    }
    case 'select-services':
      if (focusedIndex.value < detectedServices.value.length) {
        e.preventDefault();
        selectDetectedService(focusedIndex.value);
      }
      break;
    case 'service-wizard':
      e.preventDefault();
      wizardEnter();
      break;
    case 'container-image':
      if (!containerImageRef.value || creating.value || addingService.value) break;
      e.preventDefault();
      if (imageResults.value.length > 0 && focusedIndex.value < imageResults.value.length) {
        handleSelectImage(imageResults.value[focusedIndex.value]!.name);
      } else {
        handleSelectImage(containerImageRef.value);
      }
      break;
    case 'manual-service':
      if (!addingService.value && newServiceName.value) {
        e.preventDefault();
        handleAddManualService();
      }
      break;
    case 'keyValueStore':
      if (!creatingKeyValueStore.value && newKeyValueStoreName.value) {
        e.preventDefault();
        handleCreateKeyValueStore();
      }
      break;
    case 'bucket':
      if (!creatingBucket.value && newBucketName.value) {
        e.preventDefault();
        handleCreateBucket();
      }
      break;
    case 'volume':
      e.preventDefault();
      if (volumeStep.value === 'name') {
        if (newVolumeName.value) volumeStep.value = 'size';
      } else if (!creatingVolume.value && newVolumeSize.value) {
        handleCreateVolume();
      }
      break;
    case 'name-project':
      if (isProjectValid.value && !creating.value && !detectingServices.value) {
        e.preventDefault();
        handleProjectNamingNext();
      }
      break;
  }
});

type PaletteItem = {
  id: string;
  label: string;
  icon?: Component;
  iconSrc?: string;
  action: () => void;
};

const mainItems = computed(() => {
  const items: PaletteItem[] = props.context === 'projects'
    ? [
        { id: 'github-repo', label: 'GitHub Repository', icon: GithubIcon, action: () => { view.value = 'github-repos'; } },
        { id: 'container-image', label: 'Container Image', iconSrc: 'https://devicons.railway.com/i/docker.svg', action: () => { view.value = 'container-image'; } },
        { id: 'empty-project', label: 'Empty Project', icon: FolderPlus, action: () => {
          pendingRepo.value = null;
          pendingImage.value = null;
          projectDisplayName.value = '';
          projectSlug.value = '';
          view.value = 'name-project';
          nextTick(() => nameSlugRef.value?.focusName());
        } },
      ]
    : [
        { id: 'github-repo', label: 'GitHub Repository', icon: GithubIcon, action: () => { view.value = 'github-repos'; } },
        { id: 'container-image', label: 'Container Image', iconSrc: 'https://devicons.railway.com/i/docker.svg', action: () => { view.value = 'container-image'; } },
        { id: 'database', label: 'PostgreSQL', iconSrc: 'https://devicons.railway.com/i/postgresql.svg', action: () => { view.value = 'database'; } },
        { id: 'keyValueStore', label: 'Redis', iconSrc: 'https://devicons.railway.com/i/redis.svg', action: () => { view.value = 'keyValueStore'; } },
        { id: 'bucket', label: 'Bucket', icon: BucketIcon, action: () => { view.value = 'bucket'; } },
        { id: 'volume', label: 'Volume', icon: HardDrive, action: () => { volumeStep.value = 'name'; view.value = 'volume'; } },
        { id: 'manual-service', label: 'Empty Service', icon: Plus, action: () => { view.value = 'manual-service'; } },
      ];

  if (!search.value) return items;
  const q = search.value.toLowerCase();
  return items.filter(i => i.label.toLowerCase().includes(q));
});

// Suppress unused warning — kept for future env-context UI
void activeEnvironment;
</script>

<template>
  <Teleport to="body">
    <Transition name="palette">
      <div
        v-if="open"
        class="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]"
      >
        <!-- Backdrop -->
        <div
          class="absolute inset-0 bg-background/80 backdrop-blur-sm"
          @click="close"
        />

        <!-- Palette -->
        <div class="relative z-10 w-full max-w-lg rounded-xl border bg-popover shadow-2xl">
          <!-- Main view -->
          <template v-if="view === 'main'">
            <div class="flex items-center border-b px-3">
              <Search :size="18" class="shrink-0 text-muted-foreground" />
              <input
                ref="inputRef"
                v-model="search"
                placeholder="What would you like to create?"
                class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
              />            </div>
            <div class="p-1">
              <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">Create</p>
              <button
                v-for="(item, index) in mainItems"
                :key="item.id"
                :data-focused="focusedIndex === index"
                class="flex w-full items-center gap-2 rounded-lg px-2 py-2.5 text-sm text-popover-foreground transition-colors"
                :class="focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                @click="item.action()"
                @mouseenter="focusedIndex = index"
              >
                <img v-if="item.iconSrc" :src="item.iconSrc" :width="20" :height="20" class="shrink-0" alt="" />
                <component v-else-if="item.icon" :is="item.icon" :size="20" class="text-muted-foreground" />
                {{ item.label }}
              </button>
              <p v-if="mainItems.length === 0" class="px-2 py-6 text-center text-sm text-muted-foreground">
                No results found.
              </p>
            </div>
          </template>

          <!-- GitHub repos view -->
          <template v-if="view === 'github-repos'">
            <div class="flex items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <Search :size="16" class="shrink-0 text-muted-foreground" />
              <input
                ref="inputRef"
                v-model="search"
                placeholder="Search repositories..."
                class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
              />            </div>

            <template v-if="connectedLoading || sourcesLoading">
              <div class="px-2 py-6 text-center text-sm text-muted-foreground">Loading...</div>
            </template>
            <template v-else-if="!githubConnected">
              <div class="px-4 py-6 text-center">
                <GithubIcon :size="24" class="mx-auto mb-3" />
                <p class="text-sm font-medium text-foreground">Connect your GitHub account</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  Link your GitHub account to browse and import repositories.
                </p>
                <Button size="sm" class="mt-3" @click="openInstallPopup()">
                  <GithubIcon :size="14" class="mr-1.5" primary />
                  Connect GitHub
                </Button>
              </div>
            </template>
            <template v-else>
              <div
                v-if="sources.length > 0"
                class="relative border-b px-3 py-2"
              >
                <button
                  class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent"
                  @click="sourcePickerOpen = !sourcePickerOpen"
                >
                  <img
                    v-if="selectedSource?.accountAvatarUrl"
                    :src="selectedSource.accountAvatarUrl"
                    :alt="selectedSource.accountLogin"
                    class="size-5 rounded-full"
                  />
                  <span class="flex-1 text-left font-medium">{{ selectedSource?.accountLogin }}</span>
                  <Badge
                    v-if="selectedSource?.accountType === GitHubAccountType.Organization"
                    variant="outline"
                    class="text-[10px]"
                  >Org</Badge>
                  <ChevronDown :size="14" class="text-muted-foreground" />
                </button>
                <div
                  v-if="sourcePickerOpen"
                  class="absolute left-0 right-0 top-full z-20 rounded-b-xl border border-t-0 bg-popover shadow-lg"
                >
                  <div class="p-1">
                    <button
                      v-for="(source, index) in sources"
                      :key="source.accountLogin"
                      :data-focused="sourcePickerOpen && focusedIndex === index"
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm"
                      :class="sourcePickerOpen && focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                      @click="selectedSource = source; sourcePickerOpen = false"
                      @mouseenter="focusedIndex = Number(index)"
                    >
                      <img
                        :src="source.accountAvatarUrl"
                        :alt="source.accountLogin"
                        class="size-5 rounded-full"
                      />
                      <span class="flex-1 text-left">{{ source.accountLogin }}</span>
                      <Badge
                        v-if="source.accountType === GitHubAccountType.Organization"
                        variant="outline"
                        class="text-[10px]"
                      >Org</Badge>
                    </button>
                    <button
                      :data-focused="sourcePickerOpen && focusedIndex === sources.length"
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm text-muted-foreground"
                      :class="sourcePickerOpen && focusedIndex === sources.length ? 'bg-accent text-foreground' : 'hover:bg-accent hover:text-foreground'"
                      @click="openInstallPopup(); sourcePickerOpen = false"
                      @mouseenter="focusedIndex = sources.length"
                    >
                      <Plus :size="14" />
                      Add GitHub Account
                    </button>
                  </div>
                </div>
              </div>

              <div v-if="sources.length === 0" class="px-4 py-6 text-center">
                <GithubIcon :size="24" class="mx-auto mb-3" />
                <p class="text-sm font-medium text-foreground">No GitHub App installations found</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  Install the Lucity GitHub App on your account or organization.
                </p>
                <Button size="sm" class="mt-3" @click="openInstallPopup()">
                  <Plus :size="14" class="mr-1.5" />
                  Add GitHub Account
                </Button>
              </div>

              <div v-else class="max-h-[320px] overflow-y-auto">
                <div class="p-1">
                  <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">Repositories</p>
                  <template v-if="reposLoading">
                    <p class="px-2 py-6 text-center text-sm text-muted-foreground">Loading repositories...</p>
                  </template>
                  <template v-else-if="repos.length === 0">
                    <p class="px-2 py-6 text-center text-sm text-muted-foreground">No repositories found.</p>
                  </template>
                  <template v-else>
                    <button
                      v-for="(repo, index) in repos"
                      :key="repo.id"
                      :data-focused="focusedIndex === index"
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2.5 text-sm text-popover-foreground transition-colors disabled:opacity-50"
                      :class="focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                      :disabled="creating || detectingServices"
                      @click="handleSelectRepo(repo)"
                      @mouseenter="focusedIndex = Number(index)"
                    >
                      <component
                        :is="repo.private ? Lock : Globe"
                        :size="14"
                        class="shrink-0 text-muted-foreground"
                      />
                      <span class="flex-1 truncate text-left">{{ repo.fullName }}</span>
                      <Loader2
                        v-if="processingItemId === repo.fullName"
                        :size="14"
                        class="shrink-0 animate-spin text-muted-foreground"
                      />
                      <Badge
                        v-else
                        variant="outline"
                        class="shrink-0 text-[10px]"
                      >{{ repo.defaultBranch }}</Badge>
                    </button>
                  </template>
                </div>
              </div>
            </template>
          </template>

          <!-- Container image view -->
          <template v-if="view === 'container-image'">
            <div class="flex items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <Search :size="16" class="shrink-0 text-muted-foreground" />
              <input
                ref="inputRef"
                v-model="containerImageRef"
                placeholder="Search Docker Hub or enter image..."
                class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
              />
              <Loader2
                v-if="searchingImages || addingService"
                :size="14"
                class="shrink-0 animate-spin text-muted-foreground"
              />            </div>

            <div
              v-if="imageResults.length > 0"
              class="max-h-[320px] overflow-y-auto"
            >
              <div class="p-1">
                <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">Docker Hub</p>
                <button
                  v-for="(img, index) in imageResults"
                  :key="img.name"
                  :data-focused="focusedIndex === index"
                  class="flex w-full items-start gap-2 rounded-lg px-2 py-2.5 text-left text-sm text-popover-foreground transition-colors disabled:opacity-50"
                  :class="focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                  :disabled="creating || addingService"
                  @click="handleSelectImage(img.name)"
                  @mouseenter="focusedIndex = Number(index)"
                >
                  <Container :size="14" class="mt-0.5 shrink-0 text-muted-foreground" />
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-1.5">
                      <span class="font-medium">{{ img.name }}</span>
                      <Badge v-if="img.official" variant="outline" class="text-[10px]">
                        <Award :size="10" class="mr-0.5" />
                        Official
                      </Badge>
                    </div>
                    <p
                      v-if="img.description"
                      class="mt-0.5 truncate text-xs text-muted-foreground"
                    >{{ img.description }}</p>
                  </div>
                  <Loader2
                    v-if="processingItemId === img.name"
                    :size="14"
                    class="shrink-0 animate-spin text-muted-foreground"
                  />
                  <div
                    v-else
                    class="flex shrink-0 items-center gap-1 text-xs text-muted-foreground"
                  >
                    <Star :size="10" />
                    {{ formatPullCount(img.starCount) }}
                  </div>
                </button>
              </div>
            </div>

            <div
              v-if="!containerImageRef && imageResults.length === 0"
              class="px-4 py-6 text-center text-sm text-muted-foreground"
            >
              Search Docker Hub or type any image reference and press Enter.
            </div>

            <div
              v-if="containerImageRef && containerImageRef.includes('.') && !addingService"
              class="px-4 py-6 text-center text-sm text-muted-foreground"
            >
              Press Enter to deploy <span class="font-medium text-foreground">{{ containerImageRef }}</span>
            </div>
          </template>

          <!-- Manual service view -->
          <template v-if="view === 'manual-service'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Empty Service</span>
              <div class="flex-1" />            </div>
            <div class="space-y-4 p-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Service Name</label>
                <input
                  v-model="newServiceName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="web"
                />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Port (optional)</label>
                <input
                  v-model.number="newServicePort"
                  type="number"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="3000"
                />
                <p class="text-xs text-muted-foreground">
                  The port your app listens on. Leave empty if it doesn't serve traffic.
                </p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="addingService || !newServiceName"
                @click="handleAddManualService"
              >
                {{ addingService ? 'Adding...' : 'Add Service' }}
              </button>
            </div>
          </template>

          <!-- Database view (migrated to the generic CommandFlow renderer) -->
          <CommandFlow
            v-if="view === 'database'"
            :flow="databaseFlow"
            @back="view = 'main'"
            @created="handleFlowCreated"
          />

          <!-- Redis view -->
          <template v-if="view === 'keyValueStore'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Redis</span>
              <div class="flex-1" />            </div>
            <div class="space-y-4 p-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Store Name</label>
                <input
                  v-model="newKeyValueStoreName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="cache"
                />
                <p class="text-xs text-muted-foreground">Redis 8 &middot; powered by Valkey &middot; 1Gi storage</p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="creatingKeyValueStore || !newKeyValueStoreName"
                @click="handleCreateKeyValueStore"
              >
                {{ creatingKeyValueStore ? 'Creating...' : 'Create Redis' }}
              </button>
            </div>
          </template>

          <!-- Object storage view -->
          <template v-if="view === 'bucket'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Bucket</span>
              <div class="flex-1" />            </div>
            <div class="space-y-4 p-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Bucket Name</label>
                <input
                  v-model="newBucketName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="uploads"
                />
                <p class="text-xs text-muted-foreground">S3-compatible &middot; works with any S3 SDK or the AWS CLI</p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="creatingBucket || !newBucketName"
                @click="handleCreateBucket"
              >
                {{ creatingBucket ? 'Creating...' : 'Create Bucket' }}
              </button>
            </div>
          </template>

          <!-- Volume view -->
          <template v-if="view === 'volume'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="volumeStep === 'size' ? (volumeStep = 'name') : (view = 'main')"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Volume</span>
              <div class="flex-1" />            </div>

            <!-- Step: name -->
            <div v-if="volumeStep === 'name'" class="space-y-4 p-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Volume Name</label>
                <input
                  ref="inputRef"
                  v-model="newVolumeName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="data"
                />
                <p class="text-xs text-muted-foreground">Persistent storage you can mount to a service.</p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="!newVolumeName"
                @click="volumeStep = 'size'"
              >
                Next
              </button>
            </div>

            <!-- Step: size -->
            <div v-else class="space-y-4 p-4">
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <label class="text-sm font-medium text-foreground">Size</label>
                  <span class="font-mono text-sm font-semibold text-foreground">{{ newVolumeSize }}</span>
                </div>
                <Slider
                  :model-value="[volumeSizes.indexOf(newVolumeSize)]"
                  :min="0"
                  :max="volumeSizes.length - 1"
                  :step="1"
                  @update:model-value="newVolumeSize = volumeSizes[$event?.[0] ?? 0]!"
                />
                <div class="flex justify-between text-[10px] text-muted-foreground">
                  <span>{{ volumeSizes[0] }}</span>
                  <span>{{ volumeSizes[volumeSizes.length - 1] }}</span>
                </div>
                <p class="text-xs text-muted-foreground">Volumes can be grown later, but not shrunk.</p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="creatingVolume || !newVolumeSize"
                @click="handleCreateVolume"
              >
                {{ creatingVolume ? 'Creating...' : 'Create Volume' }}
              </button>
            </div>
          </template>

          <!-- Name project view -->
          <template v-if="view === 'name-project'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">New Project</span>
              <div class="flex-1" />            </div>
            <form
              class="space-y-4 p-4"
              @submit.prevent="handleProjectNamingNext"
            >
              <div
                v-if="pendingRepo || pendingImage"
                class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm text-muted-foreground"
              >
                <Loader2
                  v-if="creating || detectingServices"
                  :size="14"
                  class="shrink-0 animate-spin"
                />
                <GithubIcon
                  v-else-if="pendingRepo"
                  :size="14"
                  class="shrink-0"
                />
                <Container
                  v-else
                  :size="14"
                  class="shrink-0"
                />
                <span class="truncate">{{ pendingRepo?.fullName || pendingImage }}</span>
              </div>
              <NameSlugField
                ref="nameSlugRef"
                v-model:name="projectDisplayName"
                v-model:slug="projectSlug"
                :disabled="creating"
                name-placeholder="e.g. My API"
              />
              <button
                type="submit"
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="!isProjectValid || creating || detectingServices"
              >
                <template v-if="pendingRepo">
                  {{ detectingServices ? 'Detecting services...' : 'Continue' }}
                </template>
                <template v-else>
                  {{ creating ? 'Creating...' : 'Create Project' }}
                </template>
              </button>
            </form>
          </template>

          <!-- Select detected service view (shown when a repo yields more than one) -->
          <template v-if="view === 'select-services'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="selectServicesBack"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">Select a service</span>
              <div class="flex-1" />            </div>
            <div class="max-h-[320px] overflow-y-auto p-1">
              <p class="px-2 py-1.5 text-xs font-medium text-muted-foreground">
                Detected in {{ confirmRepo?.fullName }}
              </p>
              <button
                v-for="(svc, index) in detectedServices"
                :key="`${svc.name}-${svc.contextPath}`"
                :data-focused="focusedIndex === index"
                class="flex w-full items-center gap-2.5 rounded-lg px-2 py-2.5 text-left text-sm text-popover-foreground transition-colors"
                :class="focusedIndex === index ? 'bg-accent' : 'hover:bg-accent'"
                @click="selectDetectedService(index)"
                @mouseenter="focusedIndex = index"
              >
                <FrameworkIcon :framework="svc.framework" :language="svc.language" :size="20" />
                <div class="min-w-0 flex-1">
                  <div class="font-medium">{{ svc.name }}</div>
                  <div class="truncate text-xs text-muted-foreground">
                    {{ svc.framework || svc.language || 'Service' }} &middot;
                    <span class="font-mono">{{ svc.contextPath }}</span>
                  </div>
                </div>
                <ChevronRight :size="14" class="shrink-0 text-muted-foreground" />
              </button>
            </div>
          </template>

          <!-- Service wizard view (single-input, step-driven) -->
          <template v-if="view === 'service-wizard'">
            <div class="flex items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="wizardBack"
              >
                <ArrowLeft :size="16" />
              </button>
              <FrameworkIcon
                v-if="serviceStep === 'name'"
                :framework="activeDetected?.framework"
                :language="activeDetected?.language"
                :size="18"
              />
              <FolderGit2
                v-else-if="serviceStep === 'context'"
                :size="16"
                class="shrink-0 text-muted-foreground"
              />
              <Braces v-else :size="16" class="shrink-0 text-muted-foreground" />
              <input
                ref="inputRef"
                v-model="wizardInput"
                :placeholder="wizardPlaceholder"
                class="flex h-12 w-full bg-transparent px-3 text-sm outline-none placeholder:text-muted-foreground"
                autocomplete="off"
                data-1p-ignore
                spellcheck="false"
              />
              <Loader2
                v-if="creating || addingService"
                :size="14"
                class="shrink-0 animate-spin text-muted-foreground"
              />            </div>

            <div class="max-h-[50vh] overflow-y-auto">
              <p
                v-if="serviceStep === 'name' && serviceName && !isServiceValid"
                class="px-3 py-1.5 text-xs text-destructive"
              >
                Lowercase letters, digits and hyphens only (2&ndash;16 characters).
              </p>

              <button
                v-for="(item, index) in wizardItems"
                :key="item.id"
                :data-focused="focusedIndex === index"
                :disabled="item.disabled"
                class="group flex h-12 w-full items-center gap-2 px-3 text-sm text-popover-foreground transition-colors disabled:opacity-50"
                :class="[
                  focusedIndex === index ? 'bg-accent' : 'hover:bg-accent/40',
                  index === wizardItems.length - 1 ? 'rounded-b-xl' : 'rounded-none',
                ]"
                @click="item.action()"
              >
                <component :is="item.icon" :size="16" class="shrink-0 text-muted-foreground" />
                <span class="min-w-0 flex-1 truncate text-left" :class="{ 'font-mono': item.mono }">{{ item.label }}</span>
                <span
                  v-if="focusedIndex === index && !item.disabled"
                  class="flex shrink-0 items-center gap-1 text-xs font-medium text-green-500"
                >
                  {{ item.actionLabel }}
                  <CornerDownLeft :size="14" />
                </span>
                <span
                  v-else-if="!item.disabled"
                  class="shrink-0 text-xs text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
                >
                  {{ item.actionLabel }}
                </span>
              </button>
            </div>
          </template>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.palette-enter-active,
.palette-leave-active {
  transition: opacity 0.15s ease;
}

.palette-enter-active .relative,
.palette-leave-active .relative {
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.palette-enter-from,
.palette-leave-to {
  opacity: 0;
}

.palette-enter-from .relative,
.palette-leave-to .relative {
  transform: scale(0.96);
  opacity: 0;
}
</style>
