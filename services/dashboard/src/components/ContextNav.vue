<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Check, ChevronDown, Plus } from '@lucide/vue';
import Spinner from '@/components/LoadingSpinner.vue';
import { useAuth } from '@/composables/useAuth';
import { useEnvironment } from '@/composables/useEnvironment';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';
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
import { preferredEnvironment } from '@/lib/lastEnvironment';
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
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { darkTooltipClass } from '@/lib/tooltip';

const route = useRoute();
const router = useRouter();
const { activeWorkspace, setActiveWorkspace } = useAuth();
const { activeEnvironment, environments } = useEnvironment();
const { workspaceColor } = useWorkspaceColor();

const { result: wsResult } = useQuery(WorkspacesDocument);
const workspaces = computed(() => wsResult.value?.workspaces ?? []);
const activeWs = computed(() => workspaces.value.find(w => w.id === activeWorkspace.value));

const { result: projResult } = useQuery(ProjectsForNavDocument);
const projects = computed(() => projResult.value?.projects ?? []);

const isEnvironmentRoute = computed(() => route.name === 'environment');

const activeProjectId = computed<string | undefined>(() => {
  const id = route.params.projectId;
  return Array.isArray(id) ? id[0] : (id as string | undefined);
});

const activeProject = computed(() => projects.value.find(p => p.id === activeProjectId.value));
const activeProjectName = computed(() => activeProject.value?.name ?? activeProjectId.value ?? '');

const projectEnvironments = computed(() => {
  if (environments.value.length > 0) return environments.value;
  return activeProject.value?.environments ?? [];
});

const wsDialogOpen = ref(false);
const envDialogOpen = ref(false);
const switchingWorkspace = ref(false);
const switchingWorkspaceName = ref('');

async function handleWorkspaceSwitch(id: string) {
  if (id === activeWorkspace.value) return;
  const ws = workspaces.value.find(w => w.id === id);
  switchingWorkspaceName.value = ws?.name ?? id;
  switchingWorkspace.value = true;
  setActiveWorkspace(id);
  await router.push('/');
  await apolloClient.resetStore();
  switchingWorkspace.value = false;
}

function handleProjectSwitch(id: string) {
  if (id === activeProjectId.value) return;
  const project = projects.value.find(p => p.id === id);
  const target = preferredEnvironment(id, project?.environments ?? []);
  if (target) {
    router.push({ name: 'environment', params: { projectId: id, environmentId: target.id } });
  } else {
    router.push({ name: 'project', params: { projectId: id } });
  }
}

function handleEnvironmentSwitch(envId: string) {
  if (envId === activeEnvironment.value?.id) return;
  if (!activeProjectId.value) return;
  router.push({
    name: 'environment',
    params: { projectId: activeProjectId.value, environmentId: envId, drawer: route.params.drawer },
  });
}

function handleEnvironmentCreated(environmentId: string) {
  if (!activeProjectId.value) return;
  router.push({
    name: 'environment',
    params: { projectId: activeProjectId.value, environmentId },
  });
}

function handleNewProject() {
  router.push({ name: 'projects', query: { new: 'project' } });
}

function tierLabel(tier?: string | null) {
  if (tier === ResourceTier.Production) return 'Production';
  return 'Eco';
}

function initial(name: string) {
  return name.charAt(0).toUpperCase();
}

const titleClass = 'inline-flex h-7 items-center gap-1.5 rounded-l-md px-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted';
const arrowClass = 'flex h-7 w-5 items-center justify-center rounded-r-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground data-[state=open]:bg-muted data-[state=open]:text-foreground';
const contentClass = 'w-65 rounded-xl p-1';
const rowClass = 'h-8 gap-2 rounded-lg px-1 text-sm';
const letterClass = 'flex size-6 shrink-0 items-center justify-center rounded bg-background text-xs font-semibold text-foreground ring-1 ring-inset ring-border';
const workspaceLetterClass = 'flex size-6 shrink-0 items-center justify-center rounded text-xs font-semibold';
const plainRowClass = 'h-8 gap-2 rounded-lg px-2 text-sm';
const newBoxClass = 'flex size-6 shrink-0 items-center justify-center rounded border border-dashed border-border text-muted-foreground';
</script>

<template>
    <nav class="flex items-center gap-1">
      <span class="inline-flex items-center">
        <RouterLink :to="{ name: 'projects' }" :class="titleClass">
          <span class="flex size-4 items-center justify-center rounded text-xs font-semibold" :style="workspaceColor(activeWorkspace)">
            {{ initial(activeWs?.name ?? activeWorkspace) }}
          </span>
          <span class="max-w-35 truncate">{{ activeWs?.name ?? activeWorkspace }}</span>
        </RouterLink>
        <Tooltip>
          <TooltipTrigger as-child>
            <span class="inline-flex">
              <DropdownMenu>
                <DropdownMenuTrigger :class="arrowClass" aria-label="Switch workspace">
                  <ChevronDown :size="14" />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" :class="contentClass">
                  <DropdownMenuItem
                    v-for="ws in workspaces"
                    :key="ws.id"
                    :class="[rowClass, ws.id === activeWorkspace ? 'bg-accent/60' : '']"
                    @select="handleWorkspaceSwitch(ws.id)"
                  >
                    <span :class="workspaceLetterClass" :style="workspaceColor(ws.id)">{{ initial(ws.name) }}</span>
                    <span class="truncate" :class="ws.id === activeWorkspace ? 'font-medium' : ''">{{ ws.name }}</span>
                    <span v-if="ws.personal" class="text-xs text-muted-foreground">Personal</span>
                    <Check v-if="ws.id === activeWorkspace" class="ml-auto" />
                  </DropdownMenuItem>
                  <DropdownMenuItem :class="[rowClass, 'text-muted-foreground']" @select="wsDialogOpen = true">
                    <span :class="newBoxClass"><Plus class="!size-3.5" /></span>
                    New workspace
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </span>
          </TooltipTrigger>
          <TooltipContent side="bottom" :class="darkTooltipClass">Switch workspace</TooltipContent>
        </Tooltip>
      </span>

      <template v-if="activeProjectId">
        <span class="text-sm text-border">/</span>
        <span class="inline-flex items-center">
          <RouterLink :to="{ name: 'project', params: { projectId: activeProjectId } }" :class="titleClass">
            <span class="max-w-40 truncate">{{ activeProjectName }}</span>
          </RouterLink>
          <Tooltip>
            <TooltipTrigger as-child>
              <span class="inline-flex">
                <DropdownMenu>
                  <DropdownMenuTrigger :class="arrowClass" aria-label="Switch project">
                    <ChevronDown :size="14" />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" :class="contentClass">
                    <DropdownMenuItem
                      v-for="project in projects"
                      :key="project.id"
                      :class="[rowClass, project.id === activeProjectId ? 'bg-accent/60' : '']"
                      @select="handleProjectSwitch(project.id)"
                    >
                      <span :class="letterClass">{{ initial(project.name) }}</span>
                      <span class="truncate" :class="project.id === activeProjectId ? 'font-medium' : ''">{{ project.name }}</span>
                      <Check v-if="project.id === activeProjectId" class="ml-auto" />
                    </DropdownMenuItem>
                    <DropdownMenuItem :class="[rowClass, 'text-muted-foreground']" @select="handleNewProject">
                      <span :class="newBoxClass"><Plus class="!size-3.5" /></span>
                      New project
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </span>
            </TooltipTrigger>
            <TooltipContent side="bottom" :class="darkTooltipClass">Switch project</TooltipContent>
          </Tooltip>
        </span>
      </template>

      <template v-if="isEnvironmentRoute && activeEnvironment && activeProjectId">
        <span class="text-sm text-border">/</span>
        <span class="inline-flex items-center">
          <RouterLink
            :to="{ name: 'environment', params: { projectId: activeProjectId, environmentId: activeEnvironment.id } }"
            :class="titleClass"
          >
            {{ activeEnvironment.name }}
          </RouterLink>
          <Tooltip>
            <TooltipTrigger as-child>
              <span class="inline-flex">
                <DropdownMenu>
                  <DropdownMenuTrigger :class="arrowClass" aria-label="Switch environment">
                    <ChevronDown :size="14" />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start" :class="contentClass">
                    <DropdownMenuItem
                      v-for="env in projectEnvironments"
                      :key="env.id"
                      :class="[plainRowClass, env.id === activeEnvironment.id ? 'bg-accent/60' : '']"
                      @select="handleEnvironmentSwitch(env.id)"
                    >
                      <span class="truncate" :class="env.id === activeEnvironment.id ? 'font-medium' : ''">{{ env.name }}</span>
                      <span class="text-xs text-muted-foreground">{{ tierLabel(env.resourceTier) }}</span>
                      <Check v-if="env.id === activeEnvironment.id" class="ml-auto" />
                    </DropdownMenuItem>
                    <DropdownMenuItem :class="[plainRowClass, 'text-muted-foreground']" @select="envDialogOpen = true">
                      <Plus class="!size-3.5" />
                      New environment
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </span>
            </TooltipTrigger>
            <TooltipContent side="bottom" :class="darkTooltipClass">Switch environment</TooltipContent>
          </Tooltip>
        </span>
      </template>

    </nav>

  <CreateWorkspaceDialog v-model:open="wsDialogOpen" />
  <CreateEnvironmentDialog
    v-if="activeProjectId"
    v-model:open="envDialogOpen"
    :project-id="activeProjectId"
    @created="handleEnvironmentCreated"
  />

  <AlertDialog :open="switchingWorkspace">
    <AlertDialogContent class="flex flex-col items-center gap-4 sm:max-w-sm">
      <AlertDialogTitle class="sr-only">Switching workspace</AlertDialogTitle>
      <Spinner :size="32" class="animate-spin text-muted-foreground" />
      <AlertDialogDescription class="text-center text-sm">
        Switching to <span class="font-medium text-foreground">{{ switchingWorkspaceName }}</span>... hang tight
      </AlertDialogDescription>
    </AlertDialogContent>
  </AlertDialog>
</template>
