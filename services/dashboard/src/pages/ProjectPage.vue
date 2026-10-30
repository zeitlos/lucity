<script setup lang="ts">
import QueryError from '@/components/QueryError.vue';
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Layers, Plus, SlidersHorizontal } from '@lucide/vue';
import { graphql } from '@/gql';

const ProjectDocument = graphql(`
  query ProjectPage($id: ProjectID!) {
    project(id: $id) {
      id
      name
      environments {
        id
        name
        resourceTier
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
import PageToolbar from '@/components/PageToolbar.vue';
import ProjectSettingsDrawer from '@/components/drawer/ProjectSettingsDrawer.vue';
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue';
import EmptyState from '@/components/EmptyState.vue';
import EnvironmentCard from '@/components/EnvironmentCard.vue';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';

const route = useRoute();
const router = useRouter();

const projectId = computed(() => {
  const id = route.params.projectId;
  return Array.isArray(id) ? id[0]! : (id as string);
});

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
        resourceTier: e.resourceTier,
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

const railItems = computed(() => [
  {
    key: 'environments',
    label: 'Environments',
    icon: Layers,
    to: { name: 'project', params: { projectId: projectId.value } },
  },
  {
    key: 'settings',
    label: 'Settings',
    icon: SlidersHorizontal,
    to: { name: 'project', params: { projectId: projectId.value, drawer: 'settings' } },
  },
]);

const activeKey = computed(() => (drawer.value === 'settings' ? 'settings' : 'environments'));

function closeDrawer() {
  router.push({ name: 'project', params: { projectId: projectId.value } });
}

const envDialogOpen = ref(false);

function handleEnvironmentCreated(environmentId: string) {
  router.push({ name: 'environment', params: { projectId: projectId.value, environmentId } });
}

</script>

<template>
  <div class="flex h-[calc(100vh-var(--banner-height)-52px-0.75rem-0.5rem)] flex-col">
    <div v-if="loading && !project" class="flex flex-1 gap-3 py-3">
      <div class="w-12" />
      <div class="flex-1 space-y-4 rounded-lg border bg-card p-4">
        <Skeleton class="h-8 w-48" />
        <Skeleton class="h-24 w-full" />
        <Skeleton class="h-24 w-full" />
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

    <div v-else-if="project" class="relative flex flex-1 py-3">
      <ContextRail
        :items="railItems"
        :active-key="activeKey"
        up-label="Up to all projects"
        :up-to="{ name: 'projects' }"
        :project-id="project.id"
      />

      <div class="relative min-w-0 flex-1">
        <div class="flex h-full flex-col overflow-hidden rounded-lg border bg-card shadow-sm">
          <PageToolbar title="Environments" :count="project.environments.length">
            <Button @click="envDialogOpen = true">
              <Plus />
              New environment
            </Button>
          </PageToolbar>

          <div class="flex-1 overflow-y-auto p-4">
            <div
              v-if="project.environments.length > 0"
              class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
            >
              <EnvironmentCard
                v-for="env in project.environments"
                :key="env.id"
                :environment="env"
                :to="{ name: 'environment', params: { projectId: project.id, environmentId: env.id } }"
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

        <Transition name="slide-drawer">
          <div
            v-if="drawer === 'settings'"
            class="absolute inset-y-0 left-0 z-10 w-110 max-w-full"
          >
            <ProjectSettingsDrawer :project="project" @close="closeDrawer" />
          </div>
        </Transition>
      </div>
    </div>

    <CreateEnvironmentDialog
      v-if="project"
      v-model:open="envDialogOpen"
      :project-id="project.id"
      @created="handleEnvironmentCreated"
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
</style>
