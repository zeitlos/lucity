<script setup lang="ts">
import { environmentRoute, projectRoute, routeProjectId } from '@/lib/routes';
import QueryError from '@/components/QueryError.vue';
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Plus } from '@lucide/vue';
import { graphql } from '@/gql';

const ProjectDocument = graphql(`
  query ProjectPage($id: ProjectID!) {
    project(id: $id) {
      id
      name
      environments {
        id
        name
        services {
          id
          name
          status
          sourceUrl
        }
        databases {
          id
          name
          status
        }
        keyValueStores {
          id
          name
          status
        }
        buckets {
          id
          name
          status
        }
        volumes {
          id
          name
        }
      }
    }
  }
`);
import ContextRail from '@/components/ContextRail.vue';
import { projectNavItems } from '@/lib/contextNav';
import PageToolbar from '@/components/PageToolbar.vue';
import ProjectSettingsDrawer from '@/components/drawer/ProjectSettingsDrawer.vue';
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue';
import EmptyState from '@/components/EmptyState.vue';
import EnvironmentCard from '@/components/EnvironmentCard.vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';
import { usePageTitle } from '@/composables/usePageTitle';

const route = useRoute();
const router = useRouter();

const projectId = computed(() => routeProjectId(route) ?? '');

const drawer = computed(() => route.params.drawer);

const { result, loading, error, refetch } = useQuery(ProjectDocument, () => ({
  id: projectId.value,
}));

const project = computed(() => result.value?.project ?? null);

const { setEnvironments } = useEnvironment();

watch(
  () => project.value?.environments,
  (envs) => {
    if (envs) {
      const shells: Environment[] = envs.map(e => ({
        id: e.id,
        name: e.name,
        services: [],
        databases: [],
        keyValueStores: [],
        buckets: [],
        volumes: [],
      }));
      setEnvironments(shells);
    }
  },
  { immediate: true },
);

const railItems = computed(() => projectNavItems(projectId.value));

const activeKey = computed(() => (drawer.value === 'settings' ? 'settings' : 'environments'));

usePageTitle(() => [
  railItems.value.find(item => item.key === drawer.value)?.label,
  project.value?.name,
]);

function closeDrawer() {
  router.push(projectRoute(projectId.value));
}

function handleOutsideClick(event: MouseEvent) {
  if (drawer.value && !event.defaultPrevented) closeDrawer();
}

const envDialogOpen = ref(false);

const skeletonCards = [
  { name: 'w-28', tiles: 3 },
  { name: 'w-20', tiles: 2 },
];

function handleEnvironmentCreated(environmentId: string) {
  router.push(environmentRoute(environmentId));
}

</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div v-if="loading && !project" class="relative flex min-h-0 flex-1 py-3 max-md:py-0">
      <ContextRail
        :items="railItems"
        :active-key="activeKey"
        up-label="Up to all projects"
        :up-to="{ name: 'projects' }"
        :project-id="projectId"
      />

      <div class="relative min-w-0 flex-1">
        <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm max-md:rounded-none max-md:border-0 max-md:shadow-none">
          <PageToolbar title="Environments">
            <Button class="max-sm:w-9 max-sm:px-0" @click="envDialogOpen = true">
              <Plus />
              <span class="max-sm:sr-only">New environment</span>
            </Button>
          </PageToolbar>

          <div class="@container flex-1 overflow-y-auto p-4 max-md:pb-22">
            <div class="grid gap-4 @xl:grid-cols-2 @6xl:grid-cols-3">
              <div
                v-for="card in skeletonCards"
                :key="card.name"
                class="flex flex-col gap-3.5 rounded-lg border bg-card px-5 py-4.5 shadow-sm"
              >
                <div class="flex h-6 items-center">
                  <Skeleton :class="['h-4', card.name]" />
                </div>
                <div class="canvas-preview grid min-h-30 grid-cols-3 content-start gap-2 rounded-lg border border-border/60 p-3.5">
                  <span
                    v-for="tile in card.tiles"
                    :key="tile"
                    class="flex h-8 items-center gap-1.5 rounded-md border bg-card px-2 shadow-2xs"
                  >
                    <Skeleton class="size-3 rounded-sm" />
                    <Skeleton class="h-3 w-14" />
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else-if="error && !project" class="flex flex-1 items-center justify-center p-8">
      <QueryError
        class="max-w-md"
        title="Couldn't load this project"
        :error="error"
        @retry="refetch()"
      />
    </div>

    <div v-else-if="project" class="relative flex min-h-0 flex-1 py-3 max-md:py-0">
      <ContextRail
        :items="railItems"
        :active-key="activeKey"
        up-label="Up to all projects"
        :up-to="{ name: 'projects' }"
        :project-id="project.id"
      />

      <div class="relative min-w-0 flex-1">
        <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm max-md:rounded-none max-md:border-0 max-md:shadow-none" @click="handleOutsideClick">
          <PageToolbar title="Environments" :count="project.environments.length">
            <Button class="max-sm:w-9 max-sm:px-0" @click="envDialogOpen = true">
              <Plus />
              <span class="max-sm:sr-only">New environment</span>
            </Button>
          </PageToolbar>

          <div class="@container flex-1 overflow-y-auto p-4 max-md:pb-22">
            <div
              v-if="project.environments.length > 0"
              class="grid gap-4 @xl:grid-cols-2 @6xl:grid-cols-3"
            >
              <EnvironmentCard
                v-for="env in project.environments"
                :key="env.id"
                :environment="env"
                :to="environmentRoute(env.id)"
              />
            </div>

            <EmptyState
              v-else
              title="No environments yet"
              description="Create an environment to start deploying services."
            >
              <template #action>
                <Button @click="envDialogOpen = true">
                  <Plus />
                  New environment
                </Button>
              </template>
            </EmptyState>
          </div>
        </div>

        <Transition
          enter-active-class="transition-opacity duration-200"
          leave-active-class="transition-opacity duration-150"
          enter-from-class="opacity-0"
          leave-to-class="opacity-0"
        >
          <div v-if="drawer === 'settings'" class="absolute inset-0 z-10 bg-foreground/20 md:hidden [.dark_&]:bg-background/60" @click="closeDrawer" />
        </Transition>

        <Transition name="slide-drawer">
          <div
            v-if="drawer === 'settings'"
            class="absolute inset-y-0 left-0 z-10 w-110 max-w-full max-md:top-3"
          >
            <ProjectSettingsDrawer :project="project" @close="closeDrawer" />
          </div>
        </Transition>
      </div>
    </div>

    <CreateEnvironmentDialog
      v-model:open="envDialogOpen"
      :project-id="projectId"
      @created="handleEnvironmentCreated"
    />
  </div>
</template>

<style scoped>
.canvas-preview {
  background-color: color-mix(in oklab, var(--color-muted) 40%, transparent);
  background-image: radial-gradient(var(--color-border) 1px, transparent 1px);
  background-size: 14px 14px;
}

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
