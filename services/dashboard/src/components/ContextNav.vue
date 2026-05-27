<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { ChevronDown, Plus, Check, User, Users, Loader2, Settings } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { useEnvironment } from '@/composables/useEnvironment';
import { graphql } from '@/gql';
import { ResourceTier } from '@/gql/graphql';

const WorkspacesDocument = graphql(`
  query Workspaces {
    workspaces {
      id
      name
      personal
    }
  }
`);

const ProjectsForNavDocument = graphql(`
  query ProjectsForNav {
    projects {
      id
      name
      environments {
        id
        name
        resourceTier
      }
    }
  }
`);
import { apolloClient } from '@/lib/apollo';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import CreateWorkspaceDialog from '@/components/CreateWorkspaceDialog.vue';
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const route = useRoute();
const router = useRouter();
const { activeWorkspace, setActiveWorkspace } = useAuth();
const { activeEnvironment, environments } = useEnvironment();

const { result: wsResult } = useQuery(WorkspacesDocument);
const workspaces = computed(() => wsResult.value?.workspaces ?? []);
const activeWs = computed(() => workspaces.value.find((w: { id: string }) => w.id === activeWorkspace.value));

const { result: projResult } = useQuery(ProjectsForNavDocument);
const projects = computed(() => projResult.value?.projects ?? []);

const isEnvironmentRoute = computed(() => route.name === 'environment');
const isProjectSettingsRoute = computed(() => route.name === 'project-settings');
const isWorkspaceSettingsRoute = computed(() => route.name === 'workspace-settings');

// Resolve the current project ID from either route shape
const activeProjectId = computed<string | undefined>(() => {
  if (isProjectSettingsRoute.value) {
    const id = route.params.id;
    return Array.isArray(id) ? id[0] : (id as string | undefined);
  }
  if (isEnvironmentRoute.value && activeEnvironment.value) {
    const parts = activeEnvironment.value.id.split('/');
    if (parts.length >= 2) {
      return `${parts[0]}/${parts[1]}`;
    }
  }
  return undefined;
});

const activeProject = computed(() =>
  projects.value.find((p: { id: string }) => p.id === activeProjectId.value),
);

const wsDialogOpen = ref(false);
const envDialogOpen = ref(false);
const switchingWorkspace = ref(false);
const switchingWorkspaceName = ref('');

async function handleWorkspaceSwitch(id: string) {
  if (id === activeWorkspace.value) return;
  const ws = workspaces.value.find((w: { id: string; name: string }) => w.id === id);
  switchingWorkspaceName.value = ws?.name ?? id;
  switchingWorkspace.value = true;
  setActiveWorkspace(id);
  await router.push('/');
  await apolloClient.resetStore();
  switchingWorkspace.value = false;
}

function handleProjectSwitch(id: string) {
  if (id === activeProjectId.value) return;
  const project = projects.value.find((p: { id: string }) => p.id === id);
  const firstEnv = project?.environments?.[0];
  if (firstEnv) {
    router.push({ name: 'environment', params: { environmentId: firstEnv.id } });
  } else {
    router.push({ name: 'project-settings', params: { id, section: 'environments' } });
  }
}

function handleEnvironmentSwitch(envId: string) {
  if (envId === activeEnvironment.value?.id) return;
  router.push({ name: 'environment', params: { environmentId: envId } });
}

function tierLabel(tier?: string | null) {
  if (tier === ResourceTier.Production) return 'Production';
  return 'Eco';
}

function handleEnvSettings(envName: string) {
  if (!activeProjectId.value) return;
  router.push({
    name: 'project-settings',
    params: { id: activeProjectId.value, section: 'environments' },
    query: { env: envName },
  });
}

const projectEnvironments = computed(() => {
  if (environments.value.length > 0) return environments.value;
  return activeProject.value?.environments ?? [];
});
</script>

<template>
  <nav class="flex items-center gap-1.5">
    <!-- Workspace Dropdown -->
    <DropdownMenu>
      <DropdownMenuTrigger
        class="inline-flex items-center gap-1.5 rounded px-1.5 py-0.5 text-sm font-medium text-foreground transition-colors hover:bg-accent"
      >
        <component
          :is="activeWs?.personal ? User : Users"
          :size="14"
          class="text-muted-foreground"
        />
        <span class="max-w-[140px] truncate">{{ activeWs?.name ?? activeWorkspace }}</span>
        <ChevronDown :size="14" class="text-muted-foreground" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" class="w-56">
        <DropdownMenuItem
          v-for="ws in workspaces"
          :key="ws.id"
          @select="handleWorkspaceSwitch(ws.id)"
        >
          <div class="flex w-full items-center gap-2">
            <component
              :is="ws.personal ? User : Users"
              :size="14"
              class="shrink-0 text-muted-foreground"
            />
            <span class="truncate">{{ ws.name }}</span>
            <Check
              v-if="ws.id === activeWorkspace"
              :size="14"
              class="ml-auto shrink-0"
            />
          </div>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem @select="wsDialogOpen = true">
          <Plus :size="14" class="mr-2" />
          New Workspace
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>

    <!-- Project Dropdown -->
    <template v-if="activeProjectId">
      <span class="text-sm text-border">/</span>
      <DropdownMenu>
        <DropdownMenuTrigger
          class="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-sm font-medium text-foreground transition-colors hover:bg-accent"
        >
          <span class="max-w-[160px] truncate">{{ activeProject?.name ?? activeProjectId }}</span>
          <ChevronDown :size="14" class="text-muted-foreground" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="w-56">
          <DropdownMenuItem
            v-for="project in projects"
            :key="project.id"
            @select="handleProjectSwitch(project.id)"
          >
            <div class="flex w-full items-center gap-2">
              <span class="truncate">{{ project.name }}</span>
              <Check
                v-if="project.id === activeProjectId"
                :size="14"
                class="ml-auto shrink-0"
              />
            </div>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </template>

    <!-- Environment Dropdown (only on environment route) -->
    <template v-if="isEnvironmentRoute && activeEnvironment">
      <span class="text-sm text-border">/</span>
      <DropdownMenu>
        <DropdownMenuTrigger
          class="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-sm font-medium text-foreground transition-colors hover:bg-accent"
        >
          {{ activeEnvironment.name }}
          <ChevronDown :size="14" class="text-muted-foreground" />
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="w-64">
          <div
            v-for="env in projectEnvironments"
            :key="env.id"
            class="flex items-center"
          >
            <DropdownMenuItem
              class="flex-1"
              @select="handleEnvironmentSwitch(env.id)"
            >
              <div class="flex items-center gap-2">
                <Check
                  v-if="env.id === activeEnvironment?.id"
                  :size="14"
                  class="shrink-0"
                />
                <div v-else class="w-3.5" />
                <div class="flex flex-col">
                  <span class="truncate text-sm">{{ env.name }}</span>
                  <span class="text-[11px] text-muted-foreground">{{ tierLabel(env.resourceTier) }}</span>
                </div>
              </div>
            </DropdownMenuItem>
            <button
              class="mr-1 rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
              @click="handleEnvSettings(env.name)"
            >
              <Settings :size="13" />
            </button>
          </div>
          <DropdownMenuSeparator />
          <DropdownMenuItem @select="envDialogOpen = true">
            <Plus :size="14" class="mr-2" />
            New Environment
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </template>

    <template v-if="isWorkspaceSettingsRoute">
      <span class="text-sm text-border">/</span>
      <span class="px-1.5 py-0.5 text-sm text-muted-foreground">Settings</span>
    </template>

    <template v-if="isProjectSettingsRoute">
      <span class="text-sm text-border">/</span>
      <span class="px-1.5 py-0.5 text-sm text-muted-foreground">Settings</span>
    </template>
  </nav>

  <CreateWorkspaceDialog v-model:open="wsDialogOpen" />
  <CreateEnvironmentDialog
    v-if="activeProjectId"
    v-model:open="envDialogOpen"
    :project-id="activeProjectId"
    @created="(id) => router.push({ name: 'environment', params: { environmentId: id } })"
  />

  <AlertDialog :open="switchingWorkspace">
    <AlertDialogContent class="flex flex-col items-center gap-4 sm:max-w-sm">
      <AlertDialogTitle class="sr-only">Switching workspace</AlertDialogTitle>
      <Loader2 :size="32" class="animate-spin text-muted-foreground" />
      <AlertDialogDescription class="text-center text-sm">
        Switching to <span class="font-medium text-foreground">{{ switchingWorkspaceName }}</span>... hang tight
      </AlertDialogDescription>
    </AlertDialogContent>
  </AlertDialog>
</template>
