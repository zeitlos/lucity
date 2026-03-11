<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable';
import { Github, FolderPlus, Plus, Lock, Globe, ArrowLeft, Search, X, Database, ChevronDown, Container, Star, Award, Loader2 } from 'lucide-vue-next';
import { onKeyStroke, refDebounced } from '@vueuse/core';
import { GitHubConnectedQuery, GitHubSourcesQuery, GitHubRepositoriesQuery } from '@/graphql/github';
import { CreateProjectMutation } from '@/graphql/projects';
import { AddServiceMutation, DetectServicesQuery, DeployMutation } from '@/graphql/services';
import { SearchImagesQuery } from '@/graphql/registry';
import { CreateDatabaseMutation } from '@/graphql/databases';
import { toast } from '@/components/ui/sonner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { errorMessage } from '@/lib/utils';
import { generateName } from '@/lib/names';

const props = defineProps<{
  open: boolean;
  context: 'projects' | 'project';
  projectId?: string;
  initialView?: 'main' | 'github-repos';
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'created'): void;
}>();

const router = useRouter();
const { resolveClient } = useApolloClient();

// Drill-down state
type PaletteView = 'main' | 'github-repos' | 'manual-service' | 'database' | 'container-image';
const view = ref<PaletteView>('main');
const search = ref('');
const inputRef = ref<HTMLInputElement>();

// Source picker state
const selectedSource = ref<{ id: string; accountLogin: string; accountAvatarUrl: string; accountType: string } | null>(null);
const sourcePickerOpen = ref(false);

// Reset when palette opens
watch(() => props.open, (open) => {
  if (open) {
    view.value = props.initialView || 'main';
    search.value = '';
    selectedSource.value = null;
    sourcePickerOpen.value = false;
    containerImageRef.value = '';
    containerServiceName.value = '';
    containerPort.value = 80;
    imageSelected.value = false;
    nextTick(() => inputRef.value?.focus());
  }
});

// Focus input when view changes
watch(view, () => {
  search.value = '';
  nextTick(() => inputRef.value?.focus());
});

// Close on Escape
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
const { result: connectedResult, loading: connectedLoading } = useQuery(GitHubConnectedQuery, null, () => ({
  enabled: props.open && view.value === 'github-repos',
}));

const githubConnected = computed(() => connectedResult.value?.githubConnected ?? false);

// GitHub sources (installations)
const { result: sourcesResult, loading: sourcesLoading } = useQuery(GitHubSourcesQuery, null, () => ({
  enabled: props.open && view.value === 'github-repos' && githubConnected.value,
}));

const sources = computed(() => sourcesResult.value?.githubSources ?? []);

// Auto-select first source
watch(sources, (s) => {
  if (s.length > 0 && !selectedSource.value) {
    selectedSource.value = s[0];
  }
});

// GitHub repos for selected source
const { result: reposResult, loading: reposLoading } = useQuery(GitHubRepositoriesQuery, () => ({
  installationId: selectedSource.value?.id,
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
const { mutate: createProject, loading: creating } = useMutation(CreateProjectMutation);

async function handleSelectRepo(repo: { fullName: string; htmlUrl: string }) {
  if (props.context === 'projects') {
    await handleCreateProjectFromRepo(repo);
  } else {
    await handleAddServicesFromRepo(repo);
  }
}

async function handleCreateProjectFromRepo(repo: { fullName: string; htmlUrl: string }) {
  try {
    const projectName = generateName();

    const res = await createProject({
      input: {
        name: projectName,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to create project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const projectId = res?.data?.createProject?.id;
    if (!projectId) return;

    // Detect and add services from the selected repo
    detectingServices.value = true;
    try {
      await detectAndAddServices(projectId, repo);
    } finally {
      detectingServices.value = false;
    }

    close();
    router.push({ name: 'project', params: { id: projectId } });
  } catch (e: unknown) {
    toast.error('Failed to create project', { description: errorMessage(e) });
  }
}

// Detect services from a repo and add them to a project
const detectingServices = ref(false);

async function detectAndAddServices(projectId: string, repo: { fullName: string; htmlUrl: string }) {
  const client = resolveClient();
  const { data } = await client.query({
    query: DetectServicesQuery,
    variables: {
      sourceUrl: repo.htmlUrl,
      installationId: selectedSource.value?.id,
    },
  });

  const detected = data?.detectServices ?? [];
  if (detected.length === 0) {
    toast.info('No services detected', { description: `No services found in ${repo.fullName}` });
    return;
  }

  // Use repo short name as the service name (e.g., "cblaettl/beast-website" -> "beast-website")
  const repoName = repo.fullName.split('/').pop()!;

  const addedNames: string[] = [];
  for (const svc of detected) {
    // For single-service repos, use the repo name directly.
    // For multi-service repos (monorepos), suffix with the detected name.
    const name = detected.length === 1 ? repoName : `${repoName}-${svc.name}`;
    try {
      await addServiceMutate({
        input: {
          projectId,
          name,
          port: svc.suggestedPort,
          framework: svc.framework || undefined,
          sourceUrl: repo.htmlUrl,
          installationId: selectedSource.value?.id,
        },
      });
      addedNames.push(name);
    } catch (e: unknown) {
      toast.error(`Failed to add service ${name}`, { description: errorMessage(e) });
    }
  }

  if (addedNames.length > 0) {
    toast.success(`Added ${addedNames.length} service${addedNames.length !== 1 ? 's' : ''}`, {
      description: `from ${repo.fullName}`,
    });

    // Trigger initial deploy for each added service
    for (const name of addedNames) {
      try {
        await client.mutate({
          mutation: DeployMutation,
          variables: {
            input: {
              projectId,
              service: name,
              environment: 'development',
            },
          },
        });
      } catch {
        // Initial deploy is best-effort
      }
    }
  }
}

async function handleAddServicesFromRepo(repo: { fullName: string; htmlUrl: string }) {
  if (!props.projectId) return;

  detectingServices.value = true;
  try {
    await detectAndAddServices(props.projectId, repo);
    close();
    emit('created');
  } catch (e: unknown) {
    toast.error('Failed to detect services', { description: errorMessage(e) });
  } finally {
    detectingServices.value = false;
  }
}

// Add service (within project context)
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceMutation);

const newServiceName = ref('web');
const newServicePort = ref(3000);

// Create database (within project context)
const { mutate: createDatabaseMutate, loading: creatingDatabase } = useMutation(CreateDatabaseMutation);
const newDatabaseName = ref('main');

async function handleCreateDatabase() {
  if (!props.projectId) return;

  try {
    const res = await createDatabaseMutate({
      input: {
        projectId: props.projectId,
        name: newDatabaseName.value,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to create database', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Database created');
    close();
    emit('created');
  } catch (e: unknown) {
    toast.error('Failed to create database', { description: errorMessage(e) });
  }
}

async function handleAddManualService() {
  if (!props.projectId) return;

  try {
    const res = await addServiceMutate({
      input: {
        projectId: props.projectId,
        name: newServiceName.value,
        port: newServicePort.value,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to add service', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    toast.success('Service added');
    close();
    emit('created');
  } catch (e: unknown) {
    toast.error('Failed to add service', { description: errorMessage(e) });
  }
}

// Open GitHub App install in a popup
function openInstallPopup() {
  const w = 600;
  const h = 700;
  const left = window.screenX + (window.outerWidth - w) / 2;
  const top = window.screenY + (window.outerHeight - h) / 2;
  window.open('/auth/github/install', 'github-install', `width=${w},height=${h},left=${left},top=${top}`);
}

// Listen for postMessage from popup after GitHub App install
if (typeof window !== 'undefined') {
  window.addEventListener('message', (event) => {
    if (event.data === 'github-app-installed') {
      // Refetch sources to pick up the new installation
      const client = resolveClient();
      client.refetchQueries({ include: [GitHubSourcesQuery] });
    }
  });
}

// Container image state
const containerImageRef = ref('');
const containerImageDebounced = refDebounced(containerImageRef, 300);
const containerServiceName = ref('');
const containerPort = ref(80);
const imageSelected = ref(false); // true after user picks from autocomplete or stops typing

// Search Docker Hub based on the image input (skip if it looks like a full registry path)
const shouldSearchImages = computed(() => {
  const q = containerImageDebounced.value;
  if (!q || imageSelected.value) return false;
  // Don't search if it looks like a full registry reference (contains dots like ghcr.io/...)
  if (q.includes('.')) return false;
  return true;
});

const { result: imageSearchResult, loading: searchingImages } = useQuery(SearchImagesQuery, () => ({
  query: containerImageDebounced.value,
}), () => ({
  enabled: props.open && view.value === 'container-image' && shouldSearchImages.value,
}));

const imageResults = computed(() => {
  if (imageSelected.value) return [];
  return imageSearchResult.value?.searchImages ?? [];
});

// Reset imageSelected when user edits the input
watch(containerImageRef, () => {
  imageSelected.value = false;
  // Auto-derive service name as user types
  deriveServiceName(containerImageRef.value);
});

function selectImage(image: { name: string }) {
  containerImageRef.value = image.name;
  deriveServiceName(image.name);
  imageSelected.value = true;
}

function deriveServiceName(ref: string) {
  if (!ref) {
    containerServiceName.value = '';
    return;
  }
  // Strip tag (e.g., "nginx:1.25" -> "nginx")
  const withoutTag = ref.includes(':') ? ref.split(':')[0] ?? ref : ref;
  // Use last path segment (e.g., "bitnami/redis" -> "redis", "ghcr.io/foo/bar" -> "bar")
  const parts = withoutTag.split('/');
  containerServiceName.value = parts[parts.length - 1] ?? ref;
}

function formatPullCount(count: number): string {
  if (count >= 1_000_000_000) return `${(count / 1_000_000_000).toFixed(1)}B`;
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K`;
  return String(count);
}

async function handleAddContainerImage() {
  if (props.context === 'projects') {
    await handleCreateProjectFromImage();
  } else {
    await handleAddImageToProject();
  }
}

async function handleCreateProjectFromImage() {
  try {
    const projectName = generateName();
    const res = await createProject({ input: { name: projectName } });

    if (res?.errors?.length) {
      toast.error('Failed to create project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    const projectId = res?.data?.createProject?.id;
    if (!projectId) return;

    await addImageService(projectId);
    close();
    router.push({ name: 'project', params: { id: projectId } });
  } catch (e: unknown) {
    toast.error('Failed to create project', { description: errorMessage(e) });
  }
}

async function handleAddImageToProject() {
  if (!props.projectId) return;

  try {
    await addImageService(props.projectId);
    close();
    emit('created');
  } catch (e: unknown) {
    toast.error('Failed to add service', { description: errorMessage(e) });
  }
}

async function addImageService(projectId: string) {
  const res = await addServiceMutate({
    input: {
      projectId,
      name: containerServiceName.value,
      port: containerPort.value,
      image: containerImageRef.value,
    },
  });

  if (res?.errors?.length) {
    toast.error('Failed to add service', {
      description: res.errors.map(e => e.message).join(', '),
    });
    return;
  }

  toast.success('Service added', { description: containerImageRef.value });
}

// Main menu items filtering
const mainItems = computed(() => {
  const items = props.context === 'projects'
    ? [
        { id: 'github-repo', label: 'GitHub Repository', icon: Github, action: () => { view.value = 'github-repos'; } },
        { id: 'container-image', label: 'Container Image', icon: Container, action: () => { view.value = 'container-image'; } },
        { id: 'empty-project', label: 'Empty Project', icon: FolderPlus, action: () => { router.push('/'); close(); } },
      ]
    : [
        { id: 'github-repo', label: 'GitHub Repository', icon: Github, action: () => { view.value = 'github-repos'; } },
        { id: 'container-image', label: 'Container Image', icon: Container, action: () => { view.value = 'container-image'; } },
        { id: 'manual-service', label: 'Manual Service', icon: Plus, action: () => { view.value = 'manual-service'; } },
        { id: 'database', label: 'PostgreSQL Database', icon: Database, action: () => { view.value = 'database'; } },
      ];

  if (!search.value) return items;
  const q = search.value.toLowerCase();
  return items.filter(i => i.label.toLowerCase().includes(q));
});
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
                v-for="item in mainItems"
                :key="item.id"
                class="flex w-full items-center gap-2 rounded-lg px-2 py-2.5 text-sm text-popover-foreground transition-colors hover:bg-accent"
                @click="item.action()"
              >
                <component :is="item.icon" :size="16" class="text-muted-foreground" />
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
              <Badge variant="secondary" class="mr-2 shrink-0">GitHub</Badge>
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

            <!-- Not connected state -->
            <template v-if="connectedLoading || sourcesLoading">
              <div class="px-2 py-6 text-center text-sm text-muted-foreground">Loading...</div>
            </template>
            <template v-else-if="!githubConnected">
              <div class="px-4 py-6 text-center">
                <Github :size="24" class="mx-auto mb-3 text-muted-foreground" />
                <p class="text-sm font-medium text-foreground">Connect your GitHub account</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  Link your GitHub account to browse and import repositories.
                </p>
                <a href="/auth/github/connect" class="mt-3 inline-flex">
                  <Button size="sm">
                    <Github :size="14" class="mr-1.5" />
                    Connect GitHub
                  </Button>
                </a>
              </div>
            </template>
            <template v-else>
              <!-- Source picker -->
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
                    v-if="selectedSource?.accountType === 'ORGANIZATION'"
                    variant="outline"
                    class="text-[10px]"
                  >Org</Badge>
                  <ChevronDown :size="14" class="text-muted-foreground" />
                </button>
                <!-- Source dropdown -->
                <div
                  v-if="sourcePickerOpen"
                  class="absolute left-0 right-0 top-full z-20 rounded-b-xl border border-t-0 bg-popover shadow-lg"
                >
                  <div class="p-1">
                    <button
                      v-for="source in sources"
                      :key="source.id"
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm hover:bg-accent"
                      @click="selectedSource = source; sourcePickerOpen = false"
                    >
                      <img
                        :src="source.accountAvatarUrl"
                        :alt="source.accountLogin"
                        class="size-5 rounded-full"
                      />
                      <span class="flex-1 text-left">{{ source.accountLogin }}</span>
                      <Badge
                        v-if="source.accountType === 'ORGANIZATION'"
                        variant="outline"
                        class="text-[10px]"
                      >Org</Badge>
                    </button>
                    <button
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2 text-sm text-muted-foreground hover:bg-accent hover:text-foreground"
                      @click="openInstallPopup(); sourcePickerOpen = false"
                    >
                      <Plus :size="14" />
                      Add GitHub Account
                    </button>
                  </div>
                </div>
              </div>

              <!-- No sources state -->
              <div v-if="sources.length === 0" class="px-4 py-6 text-center">
                <Github :size="24" class="mx-auto mb-3 text-muted-foreground" />
                <p class="text-sm font-medium text-foreground">No GitHub App installations found</p>
                <p class="mt-1 text-xs text-muted-foreground">
                  Install the Lucity GitHub App on your account or organization.
                </p>
                <Button size="sm" class="mt-3" @click="openInstallPopup()">
                  <Plus :size="14" class="mr-1.5" />
                  Add GitHub Account
                </Button>
              </div>

              <!-- Repo list -->
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
                      v-for="repo in repos"
                      :key="repo.id"
                      class="flex w-full items-center gap-2 rounded-lg px-2 py-2.5 text-sm text-popover-foreground transition-colors hover:bg-accent disabled:opacity-50"
                      :disabled="creating || detectingServices"
                      @click="handleSelectRepo(repo)"
                    >
                      <component
                        :is="repo.private ? Lock : Globe"
                        :size="14"
                        class="shrink-0 text-muted-foreground"
                      />
                      <span class="flex-1 truncate text-left">{{ repo.fullName }}</span>
                      <Badge variant="outline" class="shrink-0 text-[10px]">{{ repo.defaultBranch }}</Badge>
                    </button>
                  </template>
                </div>
              </div>
            </template>
          </template>

          <!-- Container image view -->
          <template v-if="view === 'container-image'">
            <div class="flex h-12 items-center border-b px-3">
              <button
                class="mr-1 shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="view = 'main'"
              >
                <ArrowLeft :size="16" />
              </button>
              <Badge variant="secondary">Container Image</Badge>
              <div class="flex-1" />
              <button
                class="shrink-0 rounded p-1 text-muted-foreground hover:text-foreground"
                @click="close"
              >
                <X :size="16" />
              </button>
            </div>

            <div class="space-y-4 p-4">
              <!-- Image input with autocomplete -->
              <div class="relative space-y-2">
                <label class="text-sm font-medium text-foreground">Image</label>
                <div class="relative">
                  <input
                    ref="inputRef"
                    v-model="containerImageRef"
                    class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                    placeholder="nginx:1.25"
                  />
                  <Loader2
                    v-if="searchingImages"
                    :size="14"
                    class="absolute right-2.5 top-2.5 animate-spin text-muted-foreground"
                  />
                </div>

                <!-- Autocomplete dropdown -->
                <div
                  v-if="imageResults.length > 0"
                  class="absolute left-0 right-0 z-10 mt-1 max-h-[220px] overflow-y-auto rounded-md border bg-popover shadow-lg"
                >
                  <div class="p-1">
                    <button
                      v-for="img in imageResults"
                      :key="img.name"
                      class="flex w-full items-start gap-2 rounded px-2 py-1.5 text-left text-sm text-popover-foreground transition-colors hover:bg-accent"
                      @click="selectImage(img)"
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
                      <div class="flex shrink-0 items-center gap-1 text-xs text-muted-foreground">
                        <Star :size="10" />
                        {{ formatPullCount(img.starCount) }}
                      </div>
                    </button>
                  </div>
                </div>

                <p class="text-xs text-muted-foreground">
                  Searches Docker Hub. You can also enter any image from ghcr.io, quay.io, or other public registries.
                </p>
              </div>

              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Service Name</label>
                <input
                  v-model="containerServiceName"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="nginx"
                />
              </div>
              <div class="space-y-2">
                <label class="text-sm font-medium text-foreground">Port</label>
                <input
                  v-model.number="containerPort"
                  type="number"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="80"
                />
              </div>
              <button
                class="inline-flex h-9 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="addingService || !containerImageRef || !containerServiceName"
                @click="handleAddContainerImage"
              >
                {{ addingService ? 'Adding...' : 'Add Service' }}
              </button>
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
              <Badge variant="secondary">Add Service</Badge>
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
                <label class="text-sm font-medium text-foreground">Port</label>
                <input
                  v-model.number="newServicePort"
                  type="number"
                  class="flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  placeholder="3000"
                />
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
              <Badge variant="secondary">PostgreSQL Database</Badge>
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
