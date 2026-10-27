<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { ArrowRight, Layers, Plus, SlidersHorizontal } from '@lucide/vue';
import { graphql } from '@/gql';
import { ResourceTier } from '@/gql/graphql';

const ProjectDocument = graphql(`
  query ProjectPage($id: ProjectID!) {
    project(id: $id) {
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
import ContextRail from '@/components/ContextRail.vue';
import ProjectSettingsDrawer from '@/components/drawer/ProjectSettingsDrawer.vue';
import CreateEnvironmentDialog from '@/components/CreateEnvironmentDialog.vue';
import EmptyState from '@/components/EmptyState.vue';
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

const { result, loading, error } = useQuery(ProjectDocument, () => ({
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

watch(error, (err) => {
  if (err) {
    router.replace({ name: 'projects' });
  }
});
</script>

<template>
  <div class="flex h-[calc(100vh-var(--banner-height)-52px-0.75rem-0.5rem)] flex-col">
    <div v-if="loading && !project" class="flex flex-1 items-center justify-center">
      <div class="space-y-4 text-center">
        <Skeleton class="mx-auto h-8 w-48" />
        <Skeleton class="mx-auto h-4 w-64" />
      </div>
    </div>

    <div v-else-if="project" class="relative flex flex-1 py-3">
      <ContextRail
        :items="railItems"
        :active-key="activeKey"
        up-label="Up to all projects"
        :up-to="{ name: 'projects' }"
        :project-id="project.id"
        :project-name="project.name"
      />

      <div class="relative min-w-0 flex-1">
        <div class="h-full overflow-y-auto rounded-lg border bg-card shadow-sm">
          <div class="max-w-5xl space-y-6 px-10 py-8">
            <div class="flex items-start justify-between gap-6">
              <div>
                <h1 class="text-xl font-semibold text-foreground">Environments</h1>
                <p class="mt-1.5 text-sm text-muted-foreground">
                  Each environment runs its own copy of {{ project.name }}, with its own variables, resources and domains.
                </p>
              </div>
              <Button variant="outline" size="sm" class="shrink-0" @click="envDialogOpen = true">
                <Plus :size="14" class="mr-1" />
                New environment
              </Button>
            </div>

            <div
              v-if="project.environments.length > 0"
              class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3"
            >
              <RouterLink
                v-for="env in project.environments"
                :key="env.id"
                :to="{ name: 'environment', params: { projectId: project.id, environmentId: env.id } }"
                class="flex flex-col gap-3 rounded-lg border bg-card p-5 shadow-sm transition-colors hover:border-primary/50"
              >
                <div class="flex items-center justify-between gap-3">
                  <span class="truncate text-base font-semibold text-foreground">{{ env.name }}</span>
                  <span class="shrink-0 rounded-full border px-2 py-px text-[11px] text-muted-foreground">
                    {{ env.resourceTier === ResourceTier.Production ? 'Production' : 'Eco' }}
                  </span>
                </div>
                <div class="flex items-center gap-1.5 border-t border-border/50 pt-3 text-[13px] font-medium text-foreground">
                  Open canvas
                  <ArrowRight :size="14" />
                </div>
              </RouterLink>
            </div>

            <div v-else class="rounded-lg border border-dashed">
              <EmptyState
                title="No environments yet"
                description="Create an environment to start deploying services."
                pattern="crosshatch"
              >
                <template #action>
                  <Button @click="envDialogOpen = true">New environment</Button>
                </template>
              </EmptyState>
            </div>
          </div>
        </div>

        <Transition name="slide-drawer">
          <div
            v-if="drawer === 'settings'"
            class="absolute inset-y-0 left-0 z-10 w-[440px] max-w-full"
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
