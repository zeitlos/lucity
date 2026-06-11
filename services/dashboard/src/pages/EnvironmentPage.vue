<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const EnvironmentDocument = graphql(`
  query Environment($environment: EnvironmentID!) {
    environment(environment: $environment) {
      id
      name
      resourceTier
      services {
        id
        name
        status
        replicas {
          desired
          ready
        }
        autoscaling {
          minReplicas
          maxReplicas
          targetCpu
        }
        port
        endpoints {
          host
          port
          protocol
          type
          dns {
            status
            requiredRecords {
              type
              host
              value
            }
          }
          tls
        }
        sourceUrl
        contextPath
        resources {
          cpu
          memory
        }
        command
        defaultCommand
        activeDeployment {
          id
          image
          imageDigest
          commit
          commitMessage
          ref
          status
          createdAt
          deployedBy
        }
        deployments {
          id
          image
          imageDigest
          commit
          commitMessage
          ref
          status
          createdAt
          deployedBy
        }
        builds {
          id
          status
          startedAt
          finishedAt
        }
        lastDeployedAt
        createdAt
      }
      databases {
        id
        name
        version
        instances
        status
        size
        createdAt
      }
    }
  }
`);

const ProjectEnvironmentsDocument = graphql(`
  query ProjectEnvironments($id: ProjectID!) {
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
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import ServiceCanvas from '@/components/canvas/ServiceCanvas.vue';
import ServicePanel from '@/components/panel/ServicePanel.vue';
import DatabasePanel from '@/components/panel/DatabasePanel.vue';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import BuildLogsPanel from '@/components/panel/BuildLogsPanel.vue';
import ServiceLogsPanel from '@/components/panel/ServiceLogsPanel.vue';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import { useBuildLogsPanel } from '@/composables/useBuildLogsPanel';
import { useServiceLogsPanel } from '@/composables/useServiceLogsPanel';

const route = useRoute();
const router = useRouter();

const environmentId = computed(() => {
  const id = route.params.environmentId;
  return Array.isArray(id) ? id[0]! : (id as string);
});

const projectId = computed(() => {
  // EnvironmentID format: workspace/project/env-name. ProjectID is the first two segments.
  const parts = environmentId.value.split('/');
  if (parts.length >= 2) {
    return `${parts[0]}/${parts[1]}`;
  }
  return '';
});

const { result, loading, error, refetch } = useQuery(EnvironmentDocument, () => ({
  environment: environmentId.value,
}));

const environment = computed(() => result.value?.environment ?? null);

// Load sibling environments for the env switcher
const { result: projectResult } = useQuery(
  ProjectEnvironmentsDocument,
  () => ({ id: projectId.value }),
  () => ({ enabled: !!projectId.value }),
);

const {
  setEnvironments,
  setEnvironment,
  activeEnvServices,
  activeEnvDatabases,
} = useEnvironment();
const { isOpen, currentPanel, closePanel } = usePanel();
const logsPanel = useBuildLogsPanel();
const serviceLogsPanel = useServiceLogsPanel();

// Sync sibling environments into the global composable when the project loads
watch(
  () => projectResult.value?.project?.environments,
  (envs) => {
    if (envs) {
      const shells: Environment[] = envs.map(e => ({
        id: e.id,
        name: e.name,
        resourceTier: e.resourceTier,
        services: [],
        databases: [],
      }));
      setEnvironments(shells, environmentId.value);
    }
  },
  { immediate: true },
);

// When the env detail loads, replace the active environment shell with the full payload
watch(
  () => result.value?.environment,
  (env) => {
    if (!env) return;
    const full: Environment = {
      id: env.id,
      name: env.name,
      resourceTier: env.resourceTier,
      services: env.services.map(s => ({
        id: s.id,
        name: s.name,
        status: s.status,
        replicas: s.replicas,
        autoscaling: s.autoscaling ?? null,
        port: s.port,
        endpoints: s.endpoints,
        sourceUrl: s.sourceUrl,
        contextPath: s.contextPath,
        resources: s.resources,
        command: s.command,
        defaultCommand: s.defaultCommand,
        activeDeployment: s.activeDeployment
          ? {
            id: s.activeDeployment.id,
            image: s.activeDeployment.image,
            imageDigest: s.activeDeployment.imageDigest ?? null,
            commit: s.activeDeployment.commit,
            commitMessage: s.activeDeployment.commitMessage,
            ref: s.activeDeployment.ref,
            status: s.activeDeployment.status,
            createdAt: s.activeDeployment.createdAt,
            deployedBy: s.activeDeployment.deployedBy,
          }
          : null,
        deployments: s.deployments.map(d => ({
          id: d.id,
          image: d.image,
          imageDigest: d.imageDigest ?? null,
          commit: d.commit,
          commitMessage: d.commitMessage,
          ref: d.ref,
          status: d.status,
          createdAt: d.createdAt,
          deployedBy: d.deployedBy,
        })),
        builds: s.builds.map(b => ({
          id: b.id,
          status: b.status,
          startedAt: b.startedAt,
          finishedAt: b.finishedAt ?? null,
        })),
        lastDeployedAt: s.lastDeployedAt ?? null,
        createdAt: s.createdAt,
      })),
      databases: env.databases.map(d => ({
        id: d.id,
        name: d.name,
        version: d.version,
        instances: d.instances,
        status: d.status,
        size: d.size,
        createdAt: d.createdAt,
      })),
    };
    setEnvironment(full);
  },
  { immediate: true },
);

// Selected service for the panel
const selectedService = computed(() => {
  if (!currentPanel.value || currentPanel.value.type !== 'service') return null;
  return activeEnvServices.value.find(s => s.id === currentPanel.value!.id) ?? null;
});

// Selected database for the panel
const selectedDatabase = computed(() => {
  if (!currentPanel.value || currentPanel.value.type !== 'database') return null;
  return activeEnvDatabases.value.find(d => d.id === currentPanel.value!.id) ?? null;
});

// Command palette
const paletteOpen = ref(false);

function handleResourceRemoved() {
  closePanel();
  refetch();
}

function handleCreateFromPalette() {
  refetch();
}

const hasResources = computed(() => activeEnvServices.value.length > 0 || activeEnvDatabases.value.length > 0);

// If the env returns an error, bounce to the projects list
watch(error, (err) => {
  if (err) {
    router.replace({ name: 'projects' });
  }
});
</script>

<template>
  <div class="flex h-[calc(100vh-52px-0.75rem-0.5rem)] flex-col">
    <!-- Loading -->
    <div v-if="loading && !environment" class="flex flex-1 items-center justify-center">
      <div class="space-y-4 text-center">
        <Skeleton class="mx-auto h-8 w-48" />
        <Skeleton class="mx-auto h-4 w-64" />
      </div>
    </div>

    <!-- Error -->
    <div
      v-else-if="error"
      class="flex flex-1 items-center justify-center p-8"
    >
      <div class="rounded-lg border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
        Failed to load environment: {{ error.message }}
      </div>
    </div>

    <template v-else-if="environment">
      <div class="relative flex-1 py-3">
        <div class="h-full w-full overflow-hidden rounded-lg border bg-card shadow-sm">
          <template v-if="hasResources">
            <ServiceCanvas
              :services="activeEnvServices"
              :databases="activeEnvDatabases"
              @create="paletteOpen = true"
              @deploy-completed="refetch()"
            />
          </template>
          <template v-else>
            <div class="flex h-full items-center justify-center">
              <EmptyState
                title="No resources yet"
                description="Create a service or database to get started."
                pattern="crosshatch"
              >
                <template #action>
                  <Button @click="paletteOpen = true">
                    Create Resource
                  </Button>
                </template>
              </EmptyState>
            </div>
          </template>
        </div>

        <!-- Service Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedService"
            class="absolute inset-y-3 right-0 w-[55%]"
          >
            <ServicePanel
              :service="selectedService"
              @close="closePanel"
              @service-removed="handleResourceRemoved"
              @refetch="refetch()"
            />
          </div>
        </Transition>

        <!-- Database Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedDatabase"
            class="absolute inset-y-3 right-0 w-[55%]"
          >
            <DatabasePanel
              :database="selectedDatabase"
              @close="closePanel"
              @database-removed="handleResourceRemoved"
            />
          </div>
        </Transition>

        <!-- Service Runtime Logs Panel -->
        <Transition name="slide-panel">
          <div
            v-if="serviceLogsPanel.isOpen.value"
            class="absolute inset-y-3 right-0 z-10"
            style="left: calc(45% + 12px + 2rem)"
          >
            <ServiceLogsPanel
              :service-id="serviceLogsPanel.serviceId.value!"
              :service-name="serviceLogsPanel.serviceName.value"
              @close="serviceLogsPanel.close()"
            />
          </div>
        </Transition>

        <!-- Build Logs Panel -->
        <Transition name="slide-panel">
          <div
            v-if="logsPanel.isOpen.value"
            class="absolute -top-1 -right-1 bottom-6 z-10 shadow-2xl"
            style="left: calc(45% + 12px + 2rem)"
          >
            <BuildLogsPanel
              :build-id="logsPanel.buildId.value!"
              :service-name="logsPanel.serviceName.value"
              @close="logsPanel.close()"
            />
          </div>
        </Transition>
      </div>
    </template>

    <!-- Command Palette -->
    <CreateCommandPalette
      v-model:open="paletteOpen"
      context="environment"
      :environment-id="environmentId"
      @created="handleCreateFromPalette"
    />
  </div>
</template>

<style scoped>
.slide-panel-enter-active {
  transition: transform 0.3s ease, opacity 0.2s ease;
}

.slide-panel-leave-active {
  transition: transform 0.2s ease, opacity 0.15s ease;
}

.slide-panel-enter-from,
.slide-panel-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
