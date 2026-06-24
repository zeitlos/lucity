<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable';
import { FolderPlus, Plus, Lock, Globe, ArrowLeft, Search, X, ChevronDown, Container, Star, Award, Loader2 } from 'lucide-vue-next';
import type { Component } from 'vue';
import BucketIcon from '@/components/BucketIcon.vue';
import GithubIcon from '@/components/GithubIcon.vue';
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
      instances
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
import { useEnvironment } from '@/composables/useEnvironment';
import { useGitHubInstall } from '@/composables/useGitHubInstall';
import { toast, errorToast } from '@/components/ui/sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { errorMessage } from '@/lib/utils';
import { isValidSlug } from '@/lib/slug';
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
type PaletteView = 'main' | 'github-repos' | 'manual-service' | 'database' | 'keyValueStore' | 'bucket' | 'container-image' | 'name-project';
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
    selectedSource.value = sources.value[0] ?? null;
    nextTick(() => inputRef.value?.focus());
  }
});

watch(view, () => {
  search.value = '';
  focusedIndex.value = 0;
  nextTick(() => inputRef.value?.focus());
});

onKeyStroke('Escape', () => {
  if (!props.open) return;
  if (sourcePickerOpen.value) {
    sourcePickerOpen.value = false;
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

async function handleSelectRepo(repo: { fullName: string; htmlUrl: string }) {
  if (creating.value || detectingServices.value) return;
  if (props.context === 'projects') {
    showProjectNaming(repo);
  } else {
    await handleAddServicesFromRepo(repo);
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

async function handleConfirmProjectCreation() {
  if (!isProjectValid.value || creating.value) return;

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

    const firstEnv = project.environments?.[0];
    const targetEnvId = firstEnv?.id;

    // Add services from pending source into the first env
    if (targetEnvId) {
      if (pendingRepo.value) {
        detectingServices.value = true;
        try {
          await detectAndAddServices(targetEnvId, pendingRepo.value);
        } finally {
          detectingServices.value = false;
        }
      } else if (pendingImage.value) {
        await addImageService(targetEnvId, pendingImage.value);
      }
    }

    close();
    if (targetEnvId) {
      router.push({ name: 'environment', params: { projectId: project.id, environmentId: targetEnvId } });
    }
  } catch (e: unknown) {
    errorToast('Failed to create project', { description: errorMessage(e) });
  }
}

// Detect services from a repo and add them to an environment
const detectingServices = ref(false);

async function detectAndAddServices(environmentId: string, repo: { fullName: string; htmlUrl: string }) {
  const client = resolveClient();
  const { data } = await client.query({
    query: DetectServicesDocument,
    variables: {
      repositoryUrl: repo.htmlUrl,
    },
  });

  const detected = data?.detectServices ?? [];
  if (detected.length === 0) {
    toast.info('No services detected', { description: `No services found in ${repo.fullName}` });
    return;
  }

  // TODO: We should probably show a pre-filled input field where the user can customize the name themselves.
  const repoName = repo.fullName.split('/').pop()!.replace(/[._]/g, '-');

  const addedNames: string[] = [];
  for (const svc of detected) {
    const name = detected.length === 1 ? repoName : `${repoName}-${svc.name}`;
    try {
      await addServiceMutate({
        environmentId: environmentId,
        input: {
          name,
          repository: repo.fullName,
        },
      });
      addedNames.push(name);
    } catch (e: unknown) {
      errorToast(`Failed to add service ${name}`, { description: errorMessage(e) });
    }
  }

  if (addedNames.length > 0) {
    toast.success(`Added ${addedNames.length} service${addedNames.length !== 1 ? 's' : ''}`, {
      description: `from ${repo.fullName}`,
    });
  }
}

async function handleAddServicesFromRepo(repo: { fullName: string; htmlUrl: string }) {
  if (!props.environmentId) return;

  processingItemId.value = repo.fullName;
  detectingServices.value = true;
  try {
    await detectAndAddServices(props.environmentId, repo);
    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to detect services', { description: errorMessage(e) });
  } finally {
    detectingServices.value = false;
    processingItemId.value = null;
  }
}

// Add service (within environment context)
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceDocument);

const newServiceName = ref('web');
const newServicePort = ref<number | null>(null);

// Create database (within environment context)
const { mutate: createDatabaseMutate, loading: creatingDatabase } = useMutation(CreateDatabaseDocument);
const newDatabaseName = ref('main');

async function handleCreateDatabase() {
  if (!props.environmentId) return;

  try {
    const res = await createDatabaseMutate({
      input: {
        environment: props.environmentId,
        name: newDatabaseName.value,
      },
    });

    if (res?.errors?.length) {
      errorToast('Failed to create database', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Database created');
    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create database', { description: errorMessage(e) });
  }
}

// Create Redis store (within environment context)
const { mutate: createKeyValueStoreMutate, loading: creatingKeyValueStore } = useMutation(CreateKeyValueStoreDocument);
const newKeyValueStoreName = ref('cache');

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

    toast.success('Redis store created');
    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create Redis store', { description: errorMessage(e) });
  }
}

// Create object storage bucket (within environment context)
const { mutate: createBucketMutate, loading: creatingBucket } = useMutation(CreateBucketDocument);
const newBucketName = ref('uploads');

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

    toast.success('Bucket created');
    close();
    emit('created');
  } catch (e: unknown) {
    errorToast('Failed to create bucket', { description: errorMessage(e) });
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

    toast.success('Service added');
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

  toast.success('Service added', { description: imageRef });
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
    default: return 0;
  }
});

function scrollFocusedIntoView() {
  document.querySelector('[data-focused="true"]')?.scrollIntoView({ block: 'nearest' });
}

onKeyStroke('ArrowDown', (e) => {
  if (!props.open) return;
  if (view.value === 'manual-service' || view.value === 'database' || view.value === 'keyValueStore') return;
  if (currentItemCount.value === 0) return;
  e.preventDefault();
  focusedIndex.value = (focusedIndex.value + 1) % currentItemCount.value;
  nextTick(() => scrollFocusedIntoView());
});

onKeyStroke('ArrowUp', (e) => {
  if (!props.open) return;
  if (view.value === 'manual-service' || view.value === 'database' || view.value === 'keyValueStore') return;
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
    case 'database':
      if (!creatingDatabase.value && newDatabaseName.value) {
        e.preventDefault();
        handleCreateDatabase();
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
    case 'name-project':
      if (isProjectValid.value && !creating.value) {
        e.preventDefault();
        handleConfirmProjectCreation();
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
        { id: 'github-repo', label: 'GitHub Repository', icon: Github, action: () => { view.value = 'github-repos'; } },
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
        { id: 'github-repo', label: 'GitHub Repository', icon: Github, action: () => { view.value = 'github-repos'; } },
        { id: 'container-image', label: 'Container Image', iconSrc: 'https://devicons.railway.com/i/docker.svg', action: () => { view.value = 'container-image'; } },
        { id: 'database', label: 'PostgreSQL', iconSrc: 'https://devicons.railway.com/i/postgresql.svg', action: () => { view.value = 'database'; } },
        { id: 'keyValueStore', label: 'Redis', iconSrc: 'https://devicons.railway.com/i/redis.svg', action: () => { view.value = 'keyValueStore'; } },
        { id: 'bucket', label: 'Bucket', icon: BucketIcon, action: () => { view.value = 'bucket'; } },
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
              />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
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
                <img v-if="item.iconSrc" :src="item.iconSrc" :width="16" :height="16" class="shrink-0" alt="" />
                <component v-else-if="item.icon" :is="item.icon" :size="16" class="text-muted-foreground" />
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
              />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>

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
              />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>

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
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
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

          <!-- Database view -->
          <template v-if="view === 'database'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <span class="ml-1 text-sm font-medium text-foreground">PostgreSQL</span>
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
            <div class="space-y-4 p-4">
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Database Name</label>
                <input
                  v-model="newDatabaseName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="main"
                />
                <p class="text-xs text-muted-foreground">PostgreSQL 16 &middot; 1 instance &middot; 10Gi storage</p>
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="creatingDatabase || !newDatabaseName"
                @click="handleCreateDatabase"
              >
                {{ creatingDatabase ? 'Creating...' : 'Create Database' }}
              </button>
            </div>
          </template>

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
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
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
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
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
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>
            <form
              class="space-y-4 p-4"
              @submit.prevent="handleConfirmProjectCreation"
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
                {{ creating || detectingServices ? 'Creating...' : 'Create Project' }}
              </button>
            </form>
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
