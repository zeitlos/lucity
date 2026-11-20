<script setup lang="ts">
import { environmentRoute, projectRoute } from '@/lib/routes';
import { lastSegment } from '@/lib/id';
import SegmentedControl from '@/components/SegmentedControl.vue';
import ContextDrawer from '@/components/drawer/ContextDrawer.vue';
import QueryError from '@/components/QueryError.vue';
import { useQuery } from '@vue/apollo-composable';
import { useRoute, useRouter } from 'vue-router';
import { computed, ref, watch } from 'vue';
import { onKeyStroke, useLocalStorage } from '@vueuse/core';
import { LayoutGrid, List, Plus, Search } from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import { graphql } from '@/gql';

const ProjectsDocument = graphql(`
  query Projects {
    projects {
      id
      name
      environments {
        id
        name
        services {
          id
          name
          sourceUrl
        }
        databases {
          id
          name
        }
        keyValueStores {
          id
          name
        }
        buckets {
          id
          name
        }
        volumes {
          id
          name
        }
      }
    }
  }
`);

const GitHubConnectedDocument = graphql(`
  query GitHubConnected {
    githubConnected
  }
`);

const WorkspaceDocument = graphql(`
  query Workspace {
    workspace {
      id
      name
      personal
      suspended
      members {
        id
        email
        name
        role
      }
    }
  }
`);
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import PageToolbar from '@/components/PageToolbar.vue';
import ProjectCard from '@/components/ProjectCard.vue';
import ProjectTable from '@/components/ProjectTable.vue';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import WelcomeCard from '@/components/WelcomeCard.vue';
import { useOnboarding } from '@/composables/useOnboarding';
import { useGitHubInstall } from '@/composables/useGitHubInstall';
import { preferredEnvironment } from '@/lib/lastEnvironment';
import { workspaceNavItems } from '@/lib/contextNav';
import { useCapabilities } from '@/composables/useCapabilities';
import { useAuth } from '@/composables/useAuth';
import { usePageTitle } from '@/composables/usePageTitle';
import ContextRail from '@/components/ContextRail.vue';
import WorkspaceMembersDrawer from '@/components/drawer/WorkspaceMembersDrawer.vue';
import ApiTokensDrawer from '@/components/drawer/ApiTokensDrawer.vue';
import ConnectionsDrawer from '@/components/drawer/ConnectionsDrawer.vue';
import ConnectorDrawer from '@/components/drawer/ConnectorDrawer.vue';
import ChatAppDrawer from '@/components/drawer/ChatAppDrawer.vue';
import AddChannelLinkDialog from '@/components/AddChannelLinkDialog.vue';
import ModelProviderDrawer from '@/components/drawer/ModelProviderDrawer.vue';
import WorkspaceBillingDrawer from '@/components/drawer/WorkspaceBillingDrawer.vue';
import WorkspaceSettingsDrawer from '@/components/drawer/WorkspaceSettingsDrawer.vue';

const route = useRoute();
const router = useRouter();

const { result, loading, error, refetch } = useQuery(ProjectsDocument);
const { result: ghResult } = useQuery(GitHubConnectedDocument);

const projects = computed(() => result.value?.projects ?? []);
const githubConnected = computed(() => ghResult.value?.githubConnected ?? false);
const paletteOpen = ref(false);
const initialPaletteView = ref<'main' | 'github-repos'>('main');

const { isWorkspaceAdmin, activeWorkspace } = useAuth();
const { result: workspaceResult, refetch: refetchWorkspace } = useQuery(WorkspaceDocument);
const workspace = computed(() => workspaceResult.value?.workspace ?? null);

const drawer = computed(() => route.params.drawer);
const connection = computed(() => (route.name === 'connection'
  ? { kind: String(route.params.kind), name: String(route.params.name) }
  : null));
const drawerTitles: Record<string, string> = {
  'members': 'Members',
  'connections': 'Connections',
  'api-tokens': 'API tokens',
  'billing': 'Billing',
  'settings': 'Workspace settings',
};

const { agents } = useCapabilities();
const railItems = computed(() => workspaceNavItems(isWorkspaceAdmin.value, agents.value));
const activeKey = computed(() => {
  if (connection.value) return 'connections';
  return typeof drawer.value === 'string' && drawer.value ? drawer.value : 'projects';
});

usePageTitle(() => [
  railItems.value.find(item => item.key === activeKey.value && item.key !== 'projects')?.label,
  workspace.value?.name,
]);

function closeDrawer() {
  router.push({ name: 'projects' });
}

function handleOutsideClick(event: MouseEvent) {
  if ((drawer.value || connection.value) && !event.defaultPrevented) closeDrawer();
}

const { isWelcome, dismissWelcome } = useOnboarding();
const { openInstallPopup } = useGitHubInstall();

function handleWelcomeDismiss() {
  dismissWelcome();
}

function handleCreateProject() {
  dismissWelcome();
  paletteOpen.value = true;
}

function handleImportGitHub() {
  dismissWelcome();
  initialPaletteView.value = 'github-repos';
  paletteOpen.value = true;
}

watch(() => route.query.new, (val) => {
  if (val === 'project') {
    paletteOpen.value = true;
    router.replace({ query: {} });
  }
}, { immediate: true });

watch(() => route.query.github, (val) => {
  if (val === 'account_connected') {
    initialPaletteView.value = 'github-repos';
    paletteOpen.value = true;
    router.replace({ query: {} });
  }
}, { immediate: true });

watch(paletteOpen, (open) => {
  if (!open) initialPaletteView.value = 'main';
});

function projectLink(project: { id: string; environments: { id: string }[] }) {
  const target = preferredEnvironment(project.id, project.environments);
  if (target) {
    return environmentRoute(target.id);
  }
  return projectRoute(project.id);
}

function allServices(project: { environments: { services?: { id: string; name: string; sourceUrl: string }[] }[] }) {
  const seen = new Set<string>();
  const out: { name: string; sourceUrl: string }[] = [];
  for (const env of project.environments) {
    for (const svc of env.services ?? []) {
      const id = lastSegment(svc.id);
      if (!seen.has(id)) {
        seen.add(id);
        out.push(svc);
      }
    }
  }
  return out;
}

const view = useLocalStorage<'cards' | 'list'>('lucity_projects_view', 'cards');
const viewOptions = [
  { value: 'cards' as const, label: 'Cards', icon: LayoutGrid },
  { value: 'list' as const, label: 'List', icon: List },
];

const search = ref('');
const searchInput = ref<InstanceType<typeof Input> | null>(null);

const filteredProjects = computed(() => {
  const query = search.value.trim().toLowerCase();
  if (!query) return projects.value;
  return projects.value.filter(project =>
    project.name.toLowerCase().includes(query)
    || allServices(project).some(service => service.sourceUrl.toLowerCase().includes(query)),
  );
});

const showTable = computed(() => view.value === 'list' && !error.value && (loading.value || filteredProjects.value.length > 0));

const skeletonCards = [
  { name: 'w-24', repository: 'w-40', tiles: 3 },
  { name: 'w-32', repository: 'w-36', tiles: 2 },
  { name: 'w-20', repository: 'w-44', tiles: 4 },
];

onKeyStroke('/', (event) => {
  const target = event.target as HTMLElement | null;
  if (target?.closest('input, textarea, select, [contenteditable="true"]')) return;
  if (paletteOpen.value || !searchInput.value) return;
  event.preventDefault();
  searchInput.value.$el.focus();
});

</script>

<template>
    <div class="flex min-h-0 flex-1 flex-col">
      <div class="relative flex min-h-0 flex-1 py-3 max-md:py-0">
        <ContextRail :items="railItems" :active-key="activeKey" />

        <div class="relative min-w-0 flex-1">
          <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm max-md:rounded-none max-md:border-0 max-md:shadow-none" @click="handleOutsideClick">
            <PageToolbar title="Projects" :count="loading ? undefined : projects.length">
              <div v-if="loading || projects.length > 0" class="order-last flex w-full items-center gap-2.5 sm:order-none sm:w-auto">
                <div class="relative flex-1 sm:w-70 sm:flex-none">
                  <Search :size="14" class="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    ref="searchInput"
                    v-model="search"
                    placeholder="Search projects"
                    class="h-8 px-8 text-sm"
                    @keydown.escape="search = ''"
                  />
                  <kbd
                    v-if="!search"
                    class="pointer-events-none absolute top-1/2 right-2 -translate-y-1/2 rounded border bg-card px-1.5 font-sans text-xs leading-4.5 text-muted-foreground max-sm:hidden"
                  >/</kbd>
                </div>
                <SegmentedControl
                  v-model="view"
                  label="Layout"
                  :options="viewOptions"
                />
              </div>
              <Button class="max-sm:w-9 max-sm:px-0" @click="paletteOpen = true">
                <Plus />
                <span class="max-sm:sr-only">New project</span>
              </Button>
            </PageToolbar>

            <div :class="['flex-1 overflow-y-auto max-md:pb-22', showTable ? '' : 'p-4']">
              <ProjectTable v-if="loading && view === 'list'" :projects="[]" :link-to="projectLink" loading />
              <div v-else-if="loading" class="grid items-start gap-4 md:grid-cols-2 lg:grid-cols-3">
                <div
                  v-for="card in skeletonCards"
                  :key="card.name"
                  class="flex flex-col gap-3.5 rounded-lg border bg-card px-5 py-4.5 shadow-sm"
                >
                  <div class="space-y-1.5">
                    <div class="flex h-6 items-center">
                      <Skeleton :class="['h-4', card.name]" />
                    </div>
                    <div class="flex h-4 items-center gap-1.5">
                      <Skeleton class="size-3 rounded-sm" />
                      <Skeleton :class="['h-3', card.repository]" />
                    </div>
                  </div>
                  <div class="space-y-0.5 border-t border-border/60 pt-3">
                    <div class="grid min-h-7.5 grid-cols-[104px_minmax(0,1fr)] items-center gap-2.5 py-1">
                      <Skeleton class="h-3 w-20" />
                      <span class="flex gap-1">
                        <Skeleton v-for="tile in card.tiles" :key="tile" class="size-6 rounded-sm" />
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <QueryError
                v-else-if="error && !result"
                title="Couldn't load projects"
                :error="error"
                @retry="refetch()"
              />

              <WelcomeCard
                v-else-if="isWelcome && projects.length === 0"
                @dismiss="handleWelcomeDismiss"
                @create-project="handleCreateProject"
                @import-github="handleImportGitHub"
              />

              <template v-else-if="projects.length === 0">
                <EmptyState
                  v-if="!githubConnected"
                  title="Connect GitHub"
                  description="Connect your GitHub account to import repositories and deploy your first project."
                >
                  <template #action>
                    <Button @click="openInstallPopup()">
                      <GithubIcon :size="16" primary />
                      Connect GitHub
                    </Button>
                  </template>
                </EmptyState>

                <EmptyState
                  v-else
                  title="No projects yet"
                  description="Get started by connecting a GitHub repository."
                >
                  <template #action>
                    <Button @click="paletteOpen = true">
                      <Plus />
                      New project
                    </Button>
                  </template>
                </EmptyState>
              </template>

              <p v-else-if="filteredProjects.length === 0" class="py-12 text-center text-sm text-muted-foreground">
                No projects match "{{ search.trim() }}".
              </p>

              <template v-else>
                <div v-if="view === 'cards'" class="grid items-start gap-4 md:grid-cols-2 lg:grid-cols-3">
                  <ProjectCard
                    v-for="project in filteredProjects"
                    :key="project.id"
                    :project="project"
                    :to="projectLink(project)"
                  />
                </div>
                <ProjectTable v-else :projects="filteredProjects" :link-to="projectLink" />
              </template>
            </div>
          </div>

          <Transition
            enter-active-class="transition-opacity duration-200"
            leave-active-class="transition-opacity duration-150"
            enter-from-class="opacity-0"
            leave-to-class="opacity-0"
          >
            <div v-if="drawer || connection" class="absolute inset-0 z-10 bg-foreground/20 md:hidden [.dark_&]:bg-background/60" @click="closeDrawer" />
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="!workspace && drawer"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ContextDrawer :title="drawerTitles[String(drawer)] ?? ''" @close="closeDrawer">
                <div class="space-y-3">
                  <Skeleton class="h-9 w-full" />
                  <Skeleton class="h-16 w-full" />
                  <Skeleton class="h-16 w-full" />
                </div>
              </ContextDrawer>
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'members'"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <WorkspaceMembersDrawer
                :workspace="workspace"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
                @changed="refetchWorkspace()"
              />
            </div>
          </Transition>

          <Transition name="slide-drawer" mode="out-in">
            <div
              v-if="workspace && agents && drawer === 'connections'"
              key="connections"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ConnectionsDrawer
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
            <div
              v-else-if="agents && connection?.kind === 'model-providers'"
              :key="`model-providers/${connection.name}`"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ModelProviderDrawer
                :workspace="activeWorkspace"
                :name="connection.name"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
            <div
              v-else-if="agents && connection?.kind === 'connectors'"
              :key="`connectors/${connection.name}`"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ConnectorDrawer
                :workspace="activeWorkspace"
                :name="connection.name"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
            <div
              v-else-if="agents && connection?.kind === 'chat-apps'"
              :key="`chat-apps/${connection.name}`"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ChatAppDrawer
                :workspace="activeWorkspace"
                :name="connection.name"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && isWorkspaceAdmin && drawer === 'api-tokens'"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <ApiTokensDrawer @close="closeDrawer" />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'billing'"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <WorkspaceBillingDrawer
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'settings'"
              class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
            >
              <WorkspaceSettingsDrawer
                :workspace="workspace"
                :projects="result?.projects ?? null"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
                @changed="refetchWorkspace()"
              />
            </div>
          </Transition>
        </div>
      </div>

      <CreateCommandPalette
        v-model:open="paletteOpen"
        context="projects"
        :initial-view="initialPaletteView"
      />

      <AddChannelLinkDialog
        v-if="agents && route.name === 'add-channel'"
        :workspace="activeWorkspace"
        :chat-app="String(route.query.chatApp ?? '')"
        :channel="String(route.query.channel ?? '')"
        @close="router.push({ name: 'projects', params: { workspace: activeWorkspace } })"
      />
    </div>
</template>

<style scoped>
.slide-drawer-enter-active {
  transition: transform 0.25s ease, opacity 0.2s ease;
}

.slide-drawer-leave-active {
  transition: transform 0.15s ease, opacity 0.15s ease;
}

.slide-drawer-enter-from,
.slide-drawer-leave-to {
  transform: translateX(-16px);
  opacity: 0;
}
@media (width < 48rem) {
  .slide-drawer-enter-from,
  .slide-drawer-leave-to {
    transform: translateY(16px);
  }
}
</style>
