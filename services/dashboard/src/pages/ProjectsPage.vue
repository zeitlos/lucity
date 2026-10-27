<script setup lang="ts">
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
        resourceTier
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
import { Card, CardHeader, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import PageToolbar from '@/components/PageToolbar.vue';
import ProjectCard from '@/components/ProjectCard.vue';
import ProjectTable from '@/components/ProjectTable.vue';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import WelcomeCard from '@/components/WelcomeCard.vue';
import { useOnboarding } from '@/composables/useOnboarding';
import { useGitHubInstall } from '@/composables/useGitHubInstall';
import { preferredEnvironment } from '@/lib/lastEnvironment';
import { workspaceNavItems } from '@/lib/workspaceNav';
import { useAuth } from '@/composables/useAuth';
import ContextRail from '@/components/ContextRail.vue';
import WorkspaceMembersDrawer from '@/components/drawer/WorkspaceMembersDrawer.vue';
import ApiTokensDrawer from '@/components/drawer/ApiTokensDrawer.vue';
import WorkspaceBillingDrawer from '@/components/drawer/WorkspaceBillingDrawer.vue';
import WorkspaceSettingsDrawer from '@/components/drawer/WorkspaceSettingsDrawer.vue';

const route = useRoute();
const router = useRouter();

const { result, loading, error } = useQuery(ProjectsDocument);
const { result: ghResult } = useQuery(GitHubConnectedDocument);

const projects = computed(() => result.value?.projects ?? []);
const githubConnected = computed(() => ghResult.value?.githubConnected ?? false);
const paletteOpen = ref(false);
const initialPaletteView = ref<'main' | 'github-repos'>('main');

const { isWorkspaceAdmin } = useAuth();
const { result: workspaceResult, refetch: refetchWorkspace } = useQuery(WorkspaceDocument);
const workspace = computed(() => workspaceResult.value?.workspace ?? null);

const drawer = computed(() => route.params.drawer);
const railItems = computed(() => workspaceNavItems(isWorkspaceAdmin.value));
const activeKey = computed(() => (typeof drawer.value === 'string' && drawer.value ? drawer.value : 'projects'));

function closeDrawer() {
  router.push({ name: 'projects' });
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
    return { name: 'environment', params: { projectId: project.id, environmentId: target.id } };
  }
  return { name: 'project', params: { projectId: project.id } };
}

function allServices(project: { environments: { services?: { id: string; name: string; sourceUrl: string }[] }[] }) {
  const seen = new Set<string>();
  const out: { name: string; sourceUrl: string }[] = [];
  for (const env of project.environments) {
    for (const svc of env.services ?? []) {
      if (!seen.has(svc.name)) {
        seen.add(svc.name);
        out.push(svc);
      }
    }
  }
  return out;
}

const view = useLocalStorage<'cards' | 'list'>('lucity_projects_view', 'cards');
const viewOptions = [
  { key: 'cards', label: 'Cards', icon: LayoutGrid },
  { key: 'list', label: 'List', icon: List },
] as const;

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

const showTable = computed(() => view.value === 'list' && !loading.value && !error.value && filteredProjects.value.length > 0);

onKeyStroke('/', (event) => {
  const target = event.target as HTMLElement | null;
  if (target?.closest('input, textarea, select, [contenteditable="true"]')) return;
  if (paletteOpen.value || !searchInput.value) return;
  event.preventDefault();
  searchInput.value.$el.focus();
});

</script>

<template>
  <TooltipProvider :delay-duration="300">
    <div class="flex h-[calc(100vh-var(--banner-height)-52px-0.75rem-0.5rem)] flex-col">
      <div class="relative flex flex-1 py-3">
        <ContextRail :items="railItems" :active-key="activeKey" />

        <div class="relative min-w-0 flex-1">
          <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm">
            <PageToolbar title="Projects" :count="loading ? undefined : projects.length">
              <div v-if="projects.length > 0" class="relative w-[280px]">
                <Search :size="14" class="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  ref="searchInput"
                  v-model="search"
                  placeholder="Search projects"
                  class="h-8 px-8 text-[13px]"
                  @keydown.escape="search = ''"
                />
                <kbd
                  v-if="!search"
                  class="pointer-events-none absolute top-1/2 right-2 -translate-y-1/2 rounded-[5px] border bg-card px-1.5 font-sans text-[11px] leading-[18px] text-muted-foreground"
                >/</kbd>
              </div>
              <div v-if="projects.length > 0" class="flex rounded-[10px] border bg-background p-0.5">
                <Tooltip v-for="option in viewOptions" :key="option.key">
                  <TooltipTrigger as-child>
                    <button
                      type="button"
                      :aria-label="option.label"
                      :aria-pressed="view === option.key"
                      :class="[
                        'flex h-[26px] w-[30px] items-center justify-center rounded-[7px] transition-colors',
                        view === option.key ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground',
                      ]"
                      @click="view = option.key"
                    >
                      <component :is="option.icon" :size="15" />
                    </button>
                  </TooltipTrigger>
                  <TooltipContent side="bottom" :class="darkTooltipClass">{{ option.label }}</TooltipContent>
                </Tooltip>
              </div>
              <Button size="sm" class="h-8" @click="paletteOpen = true">
                <Plus :size="14" />
                New project
              </Button>
            </PageToolbar>

            <div :class="['flex-1 overflow-y-auto', showTable ? '' : 'p-4']">
              <div v-if="loading" class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card v-for="i in 3" :key="i">
                  <CardHeader>
                    <Skeleton class="h-5 w-32" />
                    <Skeleton class="mt-2 h-4 w-48" />
                  </CardHeader>
                  <CardContent>
                    <Skeleton class="h-4 w-full" />
                    <Skeleton class="mt-2 h-4 w-24" />
                  </CardContent>
                </Card>
              </div>

              <div
                v-else-if="error"
                class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
              >
                Failed to load projects: {{ error.message }}
              </div>

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
                  pattern="dots"
                >
                  <template #action>
                    <Button @click="openInstallPopup()">
                      <GithubIcon :size="16" class="mr-2" primary />
                      Connect GitHub
                    </Button>
                  </template>
                </EmptyState>

                <EmptyState
                  v-else
                  title="No projects yet"
                  description="Get started by connecting a GitHub repository."
                  pattern="dots"
                >
                  <template #action>
                    <Button @click="paletteOpen = true">
                      <Plus :size="16" class="mr-2" />
                      New Project
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

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'members'"
              class="absolute inset-y-0 left-0 z-10 w-[440px] max-w-full"
            >
              <WorkspaceMembersDrawer
                :workspace="workspace"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
                @changed="refetchWorkspace()"
              />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && isWorkspaceAdmin && drawer === 'api-tokens'"
              class="absolute inset-y-0 left-0 z-10 w-[440px] max-w-full"
            >
              <ApiTokensDrawer :workspace-name="workspace.name" @close="closeDrawer" />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'billing'"
              class="absolute inset-y-0 left-0 z-10 w-[440px] max-w-full"
            >
              <WorkspaceBillingDrawer
                :workspace-name="workspace.name"
                :is-admin="isWorkspaceAdmin"
                @close="closeDrawer"
              />
            </div>
          </Transition>

          <Transition name="slide-drawer">
            <div
              v-if="workspace && drawer === 'settings'"
              class="absolute inset-y-0 left-0 z-10 w-[440px] max-w-full"
            >
              <WorkspaceSettingsDrawer
                :workspace="workspace"
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
    </div>
  </TooltipProvider>
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
</style>
