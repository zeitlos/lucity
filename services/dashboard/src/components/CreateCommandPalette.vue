<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { Github, FolderPlus, Plus, Lock, Globe, ArrowLeft, Search, X } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { GitHubRepositoriesQuery } from '@/graphql/github';
import { CreateProjectMutation } from '@/graphql/projects';
import { AddServiceMutation } from '@/graphql/services';
import { toast } from '@/components/ui/sonner';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  context: 'projects' | 'project';
  projectId?: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
  (e: 'created'): void;
}>();

const router = useRouter();

// Drill-down state
type PaletteView = 'main' | 'github-repos' | 'manual-service';
const view = ref<PaletteView>('main');
const search = ref('');
const inputRef = ref<HTMLInputElement>();

// Reset when palette opens
watch(() => props.open, (open) => {
  if (open) {
    view.value = 'main';
    search.value = '';
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
  if (view.value !== 'main') {
    view.value = 'main';
  } else {
    close();
  }
});

function close() {
  emit('update:open', false);
}

// GitHub repos
const { result: reposResult, loading: reposLoading } = useQuery(GitHubRepositoriesQuery, null, () => ({
  enabled: props.open && view.value === 'github-repos' && props.context === 'projects',
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
  try {
    const res = await createProject({
      input: {
        name: repo.fullName,
        sourceUrl: repo.htmlUrl,
      },
    });

    if (res?.errors?.length) {
      toast.error('Failed to create project', {
        description: res.errors.map(e => e.message).join(', '),
      });
      return;
    }

    close();
    if (res?.data?.createProject) {
      router.push({ name: 'project', params: { id: res.data.createProject.id } });
    }
  } catch (e: unknown) {
    toast.error('Failed to create project', { description: errorMessage(e) });
  }
}

// Add service (within project context)
const { mutate: addServiceMutate, loading: addingService } = useMutation(AddServiceMutation);

const newServiceName = ref('web');
const newServicePort = ref(3000);

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

// Main menu items filtering
const mainItems = computed(() => {
  const items = props.context === 'projects'
    ? [
        { id: 'github-repo', label: 'GitHub Repository', icon: Github, action: () => { view.value = 'github-repos'; } },
        { id: 'empty-project', label: 'Empty Project', icon: FolderPlus, action: () => { router.push('/'); close(); } },
      ]
    : [
        { id: 'manual-service', label: 'Manual Service', icon: Plus, action: () => { view.value = 'manual-service'; } },
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
        <div class="relative z-10 w-full max-w-lg overflow-hidden rounded-xl border bg-popover shadow-2xl">
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
            <ScrollArea class="max-h-[320px]">
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
                    :disabled="creating"
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
            </ScrollArea>
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
