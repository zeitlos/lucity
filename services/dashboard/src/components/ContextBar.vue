<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { onKeyStroke } from '@vueuse/core';
import { BookOpen, Check, ChevronDown, Download, LogOut, Plus } from '@lucide/vue';
import IconButton from '@/components/IconButton.vue';
import MenuIcon from '@/components/MenuIcon.vue';
import ThemeToggle from '@/components/ThemeToggle.vue';
import TrialIndicator from '@/components/TrialIndicator.vue';
import Spinner from '@/components/LoadingSpinner.vue';
import CreateWorkspaceDialog from '@/components/CreateWorkspaceDialog.vue';
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue';
import ProjectEjectDialog from '@/components/ProjectEjectDialog.vue';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useAuth } from '@/composables/useAuth';
import { useCapabilities } from '@/composables/useCapabilities';
import { tierLabel, useContextSwitch } from '@/composables/useContextSwitch';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';
import { environmentNavItems, projectNavItems, workspaceNavItems } from '@/lib/contextNav';
import { preferredEnvironment } from '@/lib/lastEnvironment';

defineProps<{
  suspended: boolean;
}>();

const BAR_KEYS = new Set(['projects', 'members', 'billing']);

const route = useRoute();
const router = useRouter();
const { user, logout, isWorkspaceAdmin } = useAuth();
const { agents } = useCapabilities();
const { workspaceColor } = useWorkspaceColor();
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
  switchingWorkspace,
  switchingWorkspaceName,
  handleWorkspaceSwitch,
  handleEnvironmentCreated,
  handleNewProject,
} = useContextSwitch();

const environmentId = computed<string | undefined>(() => {
  const id = route.params.environmentId;
  return Array.isArray(id) ? id[0] : (id as string | undefined);
});

const tabs = computed(() => {
  if (activeProjectId.value && isEnvironmentRoute.value && environmentId.value) {
    return environmentNavItems(activeProjectId.value, environmentId.value);
  }
  if (activeProjectId.value) return projectNavItems(activeProjectId.value);
  return workspaceNavItems(isWorkspaceAdmin.value, agents.value).filter(item => BAR_KEYS.has(item.key));
});

const sheetItems = computed(() =>
  workspaceNavItems(isWorkspaceAdmin.value, agents.value).filter(item => !BAR_KEYS.has(item.key)),
);

const activeKey = computed(() => {
  if (route.name === 'connection') return 'connections';
  const drawer = route.params.drawer;
  return typeof drawer === 'string' && drawer ? drawer : tabs.value[0]?.key;
});

const navLabel = computed(() => {
  if (isEnvironmentRoute.value) return 'Environment';
  return activeProjectId.value ? 'Project' : 'Workspace';
});

const workspaceName = computed(() => activeWs.value?.name ?? activeWorkspace.value);
const environmentName = computed(() => activeEnvironment.value?.name ?? environmentId.value?.split('/').pop() ?? '');
const parentName = computed(() => (isEnvironmentRoute.value ? activeProjectName.value : workspaceName.value));
const leafName = computed(() => (isEnvironmentRoute.value ? environmentName.value : activeProjectName.value));

const otherWorkspaces = computed(() => workspaces.value.filter(ws => ws.id !== activeWorkspace.value));
const otherProjects = computed(() => projects.value.filter(project => project.id !== activeProjectId.value));

const open = ref(false);
const workspacesOpen = ref(false);
const wsDialogOpen = ref(false);
const envDialogOpen = ref(false);
const ejectOpen = ref(false);

watch(() => route.fullPath, () => {
  open.value = false;
});

watch(open, (value) => {
  if (!value) workspacesOpen.value = false;
});

onKeyStroke('Escape', () => {
  open.value = false;
});

function closeOnLink(event: MouseEvent) {
  if ((event.target as HTMLElement).closest('a')) open.value = false;
}

function switchWorkspace(id: string) {
  open.value = false;
  handleWorkspaceSwitch(id);
}

function openWorkspaceDialog() {
  open.value = false;
  wsDialogOpen.value = true;
}

function openEnvironmentDialog() {
  open.value = false;
  envDialogOpen.value = true;
}

function openEject() {
  open.value = false;
  ejectOpen.value = true;
}

function newProject() {
  open.value = false;
  handleNewProject();
}

async function handleLogout() {
  await logout();
  router.push('/login');
}

function projectLink(project: { id: string; environments: { id: string }[] }) {
  const target = preferredEnvironment(project.id, project.environments);
  if (target) {
    return { name: 'environment', params: { projectId: project.id, environmentId: target.id } };
  }
  return { name: 'project', params: { projectId: project.id } };
}

function environmentLink(id: string) {
  return {
    name: 'environment',
    params: {
      projectId: activeProjectId.value,
      environmentId: id,
      drawer: isEnvironmentRoute.value ? route.params.drawer : undefined,
    },
  };
}

function initial(name: string) {
  return name.charAt(0).toUpperCase();
}

const rowClass = 'flex h-11 w-full shrink-0 items-center gap-2.5 rounded-lg px-3 text-left text-sm transition-colors hover:bg-foreground/10';
const sepClass = 'mx-2 my-1 h-px shrink-0 bg-foreground/15';
const letterClass = 'flex size-6.5 shrink-0 items-center justify-center rounded-sm bg-card text-xs font-semibold ring-1 ring-inset ring-foreground/15';
const workspaceLetterClass = 'flex size-6.5 shrink-0 items-center justify-center rounded-sm text-xs font-semibold';
const newBoxClass = 'flex size-6.5 shrink-0 items-center justify-center rounded-sm border border-dashed border-foreground/20 text-muted-foreground';
</script>

<template>
  <Transition
    enter-active-class="transition-opacity duration-200"
    leave-active-class="transition-opacity duration-150"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
  >
    <div
      v-if="open"
      aria-hidden="true"
      class="fixed inset-0 z-40 bg-foreground/20 md:hidden [.dark_&]:bg-background/60"
      @click="open = false"
    />
  </Transition>

  <div class="pointer-events-none fixed inset-x-2 top-14 bottom-3 z-40 flex flex-col justify-end md:hidden">
    <div class="dark pointer-events-auto flex max-h-full flex-col rounded-xl bg-card p-1 text-foreground shadow-2xl [.dark_&]:bg-secondary [.dark_&]:ring-1 [.dark_&]:ring-foreground/10">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        leave-active-class="transition duration-150 ease-in"
        enter-from-class="translate-y-2 opacity-0"
        leave-to-class="translate-y-2 opacity-0"
      >
        <div
          v-if="open"
          id="context-sheet"
          class="flex min-h-0 flex-col overflow-y-auto overscroll-contain"
          @click="closeOnLink"
        >
          <div v-if="!activeProjectId" class="flex h-13 shrink-0 items-center gap-3 px-3">
            <span class="flex size-8 shrink-0 items-center justify-center rounded-sm text-sm font-semibold" :style="workspaceColor(activeWorkspace)">
              {{ initial(workspaceName) }}
            </span>
            <span class="min-w-0 flex-1 truncate text-base font-semibold">{{ workspaceName }}</span>
            <span v-if="activeWs?.personal" class="shrink-0 rounded-full bg-muted px-2 py-px text-xs text-muted-foreground">Personal</span>
            <TrialIndicator v-if="!suspended" />
          </div>
          <div v-else class="flex shrink-0 items-center gap-1">
            <RouterLink :to="{ name: 'projects' }" :class="[rowClass, 'min-w-0 flex-1 font-semibold']">
              <span :class="workspaceLetterClass" :style="workspaceColor(activeWorkspace)">{{ initial(workspaceName) }}</span>
              <span class="truncate">{{ workspaceName }}</span>
            </RouterLink>
            <TrialIndicator v-if="!suspended" />
            <button
              type="button"
              aria-label="Switch workspace"
              :aria-expanded="workspacesOpen"
              class="flex size-11 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-foreground/10"
              @click="workspacesOpen = !workspacesOpen"
            >
              <ChevronDown :size="16" class="transition-transform" :class="{ 'rotate-180': workspacesOpen }" />
            </button>
          </div>

          <template v-if="!activeProjectId">
            <RouterLink v-for="item in sheetItems" :key="item.key" :to="item.to" :class="rowClass">
              <MenuIcon :icon="item.icon" />
              {{ item.label }}
            </RouterLink>
          </template>

          <template v-if="!activeProjectId || workspacesOpen">
            <div :class="sepClass" />
            <button
              v-for="ws in otherWorkspaces"
              :key="ws.id"
              type="button"
              :class="rowClass"
              @click="switchWorkspace(ws.id)"
            >
              <span :class="workspaceLetterClass" :style="workspaceColor(ws.id)">{{ initial(ws.name) }}</span>
              <span class="truncate">{{ ws.name }}</span>
              <span v-if="ws.personal" class="text-xs text-muted-foreground">Personal</span>
            </button>
            <button type="button" :class="[rowClass, 'text-muted-foreground']" @click="openWorkspaceDialog">
              <span :class="newBoxClass"><Plus :size="14" /></span>
              New workspace
            </button>
          </template>

          <template v-if="activeProjectId">
            <div :class="sepClass" />
            <RouterLink
              :to="{ name: 'project', params: { projectId: activeProjectId } }"
              class="flex h-13 shrink-0 items-center gap-3 rounded-lg px-3 transition-colors hover:bg-foreground/10"
            >
              <span class="flex size-8 shrink-0 items-center justify-center rounded-sm bg-card text-sm font-semibold ring-1 ring-inset ring-foreground/15">
                {{ initial(activeProjectName) }}
              </span>
              <span class="min-w-0 flex-1 truncate text-base font-semibold">{{ activeProjectName }}</span>
            </RouterLink>
            <button type="button" :class="rowClass" @click="openEject">
              <MenuIcon :icon="Download" />
              Eject
            </button>
            <div :class="sepClass" />
            <RouterLink
              v-for="env in projectEnvironments"
              :key="env.id"
              :to="environmentLink(env.id)"
              :aria-current="isEnvironmentRoute && env.id === environmentId ? 'page' : undefined"
              :class="[rowClass, 'pl-12']"
            >
              <span class="min-w-0 truncate" :class="{ 'font-medium': isEnvironmentRoute && env.id === environmentId }">{{ env.name }}</span>
              <span class="text-xs text-muted-foreground">{{ tierLabel(env.resourceTier) }}</span>
              <Check v-if="isEnvironmentRoute && env.id === environmentId" :size="16" class="ml-auto shrink-0" />
            </RouterLink>
            <button type="button" :class="[rowClass, 'text-muted-foreground']" @click="openEnvironmentDialog">
              <span :class="newBoxClass"><Plus :size="14" /></span>
              New environment
            </button>
            <div :class="sepClass" />
            <RouterLink v-for="project in otherProjects" :key="project.id" :to="projectLink(project)" :class="rowClass">
              <span :class="letterClass">{{ initial(project.name) }}</span>
              <span class="truncate">{{ project.name }}</span>
            </RouterLink>
            <button type="button" :class="[rowClass, 'text-muted-foreground']" @click="newProject">
              <span :class="newBoxClass"><Plus :size="14" /></span>
              New project
            </button>
          </template>

          <div :class="sepClass" />
          <div v-if="user" class="flex h-11 shrink-0 items-center gap-3 pr-1 pl-3">
            <Avatar class="size-7">
              <AvatarImage :src="user.avatarUrl" :alt="user.name || user.email || ''" />
              <AvatarFallback>{{ (user.name || user.email || '?').charAt(0).toUpperCase() }}</AvatarFallback>
            </Avatar>
            <span class="min-w-0 flex-1 truncate text-sm font-semibold">{{ user.name || user.email }}</span>
            <IconButton
              label="Docs"
              size="icon"
              class="text-muted-foreground"
              as="a"
              href="https://lucity.cloud/docs/quickstart"
              target="_blank"
              rel="noopener"
            >
              <BookOpen />
            </IconButton>
            <ThemeToggle class="mx-1" />
            <IconButton label="Sign out" size="icon" class="text-muted-foreground" @click="handleLogout">
              <LogOut />
            </IconButton>
          </div>
          <div :class="sepClass" />
        </div>
      </Transition>

      <nav :aria-label="navLabel" class="flex h-12 shrink-0 items-center gap-0.5">
        <button
          type="button"
          aria-controls="context-sheet"
          :aria-expanded="open"
          class="flex h-12 min-w-0 flex-1 items-center gap-2 rounded-lg px-2 text-left transition-colors"
          :class="{ 'bg-foreground/10': open }"
          @click="open = !open"
        >
          <span class="flex size-7 shrink-0 items-center justify-center rounded-sm text-sm font-semibold" :style="workspaceColor(activeWorkspace)">
            {{ initial(workspaceName) }}
          </span>
          <span v-if="!activeProjectId" class="truncate text-sm font-semibold">{{ workspaceName }}</span>
          <span v-else class="flex min-w-0 flex-col gap-0.5">
            <span class="truncate text-xs text-muted-foreground">{{ parentName }}</span>
            <span class="truncate text-sm font-semibold">{{ leafName }}</span>
          </span>
          <ChevronDown :size="14" class="shrink-0 text-muted-foreground" />
        </button>
        <span aria-hidden="true" class="mx-1 h-6 w-px shrink-0 bg-foreground/15" />
        <RouterLink
          v-for="tab in tabs"
          :key="tab.key"
          :to="tab.to"
          :aria-label="tab.label"
          :aria-current="tab.key === activeKey ? 'page' : undefined"
          class="flex size-12 shrink-0 items-center justify-center rounded-lg transition-colors"
          :class="tab.key === activeKey ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'"
        >
          <component :is="tab.icon" :size="18" />
        </RouterLink>
      </nav>
    </div>
  </div>

  <CreateWorkspaceDialog v-model:open="wsDialogOpen" />
  <CreateEnvironmentDialog
    v-if="activeProjectId"
    v-model:open="envDialogOpen"
    :project-id="activeProjectId"
    @created="handleEnvironmentCreated"
  />
  <ProjectEjectDialog
    v-if="activeProjectId"
    v-model:open="ejectOpen"
    :project-id="activeProjectId"
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
