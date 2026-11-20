<script setup lang="ts">
import { environmentRoute, projectRoute } from '@/lib/routes';
import { ref } from 'vue';
import { Check, ChevronDown, Plus } from '@lucide/vue';
import { useContextSwitch } from '@/composables/useContextSwitch';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';
import WorkspaceSwitchDialog from '@/components/WorkspaceSwitchDialog.vue';
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

const {
  activeWorkspace,
  activeEnvironment,
  workspaces,
  activeWs,
  projects,
  isEnvironmentRoute,
  activeProjectId,
  activeProjectName,
  projectEnvironments,
  workspaceSwitch,
  handleWorkspaceSwitch,
  finishWorkspaceSwitch,
  handleProjectSwitch,
  handleEnvironmentSwitch,
  handleEnvironmentCreated,
  handleNewProject,
} = useContextSwitch();
const { workspaceColor } = useWorkspaceColor();

const wsDialogOpen = ref(false);
const envDialogOpen = ref(false);

function initial(name: string) {
  return name.charAt(0).toUpperCase();
}

const titleClass = 'inline-flex h-7 min-w-0 items-center gap-1.5 rounded-l-md px-1.5 text-sm font-medium text-foreground transition-colors hover:bg-muted';
const arrowClass = 'flex h-7 w-5 shrink-0 items-center justify-center rounded-r-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground data-[state=open]:bg-muted data-[state=open]:text-foreground';
const contentClass = 'w-65 rounded-xl p-1';
const rowClass = 'h-8 gap-2 rounded-lg px-1 text-sm';
const letterClass = 'flex size-6 shrink-0 items-center justify-center rounded bg-background text-xs font-semibold text-foreground ring-1 ring-inset ring-border';
const workspaceLetterClass = 'flex size-6 shrink-0 items-center justify-center rounded text-xs font-semibold';
const plainRowClass = 'h-8 gap-2 rounded-lg px-2 text-sm';
const newBoxClass = 'flex size-6 shrink-0 items-center justify-center rounded border border-dashed border-border text-muted-foreground';
</script>

<template>
    <nav class="flex min-w-0 items-center gap-1">
      <span class="inline-flex min-w-0 items-center">
        <RouterLink :to="{ name: 'projects' }" :class="titleClass">
          <span class="flex size-4 shrink-0 items-center justify-center rounded text-xs font-semibold" :style="workspaceColor(activeWorkspace)">
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
        <span class="inline-flex min-w-0 items-center">
          <RouterLink :to="projectRoute(activeProjectId)" :class="titleClass">
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
        <span class="inline-flex min-w-0 items-center">
          <RouterLink
            :to="environmentRoute(activeEnvironment.id)"
            :class="titleClass"
          >
            <span class="truncate">{{ activeEnvironment.name }}</span>
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

  <WorkspaceSwitchDialog :workspace-switch="workspaceSwitch" @finished="finishWorkspaceSwitch" />
</template>
