<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { graphql } from '@/gql';
import { ServiceStatus, DatabaseStatus, ReleaseStatus, RolloutStatus, ScanStatus } from '@/gql/graphql';

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
        branch
        autoDeploy
        contextPath
        resources {
          cpu
          memory
        }
        command
        activeDeployment {
          id
          image
          imageDigest
          commit
          commitMessage
          ref
          status
          createdAt
          replicas {
            desired
            ready
          }
          rollout {
            status
            reason
            message
            restarts
            startedAt
          }
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
        }
        builds {
          id
          status
          startedAt
          finishedAt
        }
        releases {
          id
          status
          createdAt
          trigger {
            kind
            actor
          }
          source {
            provider
            repository
            url
            ref
            contextPath
            commit {
              sha
              message
              url
            }
          }
          build {
            id
            status
            startedAt
            finishedAt
          }
          deploy {
            id
            status
            startedAt
            finishedAt
          }
          scan {
            id
            status
            findingsCount
            verifiedCount
            startedAt
            finishedAt
          }
          deployment {
            id
            image
            imageDigest
            commit
            commitMessage
            ref
            status
            createdAt
            replicas {
              desired
              ready
            }
            rollout {
              status
              reason
              message
              restarts
              startedAt
            }
          }
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
      keyValueStores {
        id
        name
        version
        status
        size
        createdAt
      }
      buckets {
        id
        name
        region
        endpoint
        status
        sizeBytes
        objectCount
        createdAt
      }
      volumes {
        id
        name
        size
        mount {
          service
          path
        }
      }
    }
  }
`);

const EnvironmentVolumeUsageDocument = graphql(`
  query EnvironmentVolumeUsage($environment: EnvironmentID!) {
    environment(environment: $environment) {
      id
      volumes {
        id
        size
        metrics(metrics: [STORAGE_USED], range: { window: LAST_1H }) {
          points {
            value
          }
        }
      }
    }
  }
`);

const EnvironmentDefaultCommandsDocument = graphql(`
  query EnvironmentDefaultCommands($environment: EnvironmentID!) {
    environment(environment: $environment) {
      id
      services {
        id
        defaultCommand
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
import KeyValueStorePanel from '@/components/panel/KeyValueStorePanel.vue';
import BucketPanel from '@/components/panel/BucketPanel.vue';
import VolumePanel from '@/components/panel/VolumePanel.vue';
import EmptyState from '@/components/EmptyState.vue';
import CreateCommandPalette from '@/components/CreateCommandPalette.vue';
import MountVolumeDialog from '@/components/MountVolumeDialog.vue';
import BuildLogsPanel from '@/components/panel/BuildLogsPanel.vue';
import ServiceLogsPanel from '@/components/panel/ServiceLogsPanel.vue';
import { useEnvironment, type Environment } from '@/composables/useEnvironment';
import { usePanel } from '@/composables/usePanel';
import { useBuildLogsPanel } from '@/composables/useBuildLogsPanel';
import { useServiceLogsPanel } from '@/composables/useServiceLogsPanel';
import { parseStorageSize } from '@/lib/utils';

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

const SERVICE_TRANSIENT_STATUSES = new Set<ServiceStatus>([
  ServiceStatus.Deploying,
  ServiceStatus.Degraded,
]);
const DATABASE_TRANSIENT_STATUSES = new Set<DatabaseStatus>([
  DatabaseStatus.Pending,
  DatabaseStatus.Updating,
  DatabaseStatus.Degraded,
]);
const RELEASE_TRANSIENT_STATUSES = new Set<ReleaseStatus>([
  ReleaseStatus.Queued,
  ReleaseStatus.Building,
  ReleaseStatus.Deploying,
]);
const RELEASE_POLL_MAX_AGE_MS = 60 * 60 * 1000;
const SCAN_TRANSIENT_STATUSES = new Set<ScanStatus>([ScanStatus.Queued, ScanStatus.Running]);

function isReleaseInFlight(release: { status: ReleaseStatus; createdAt: string; scan?: { status: ScanStatus } | null }): boolean {
  const active = RELEASE_TRANSIENT_STATUSES.has(release.status)
    || (release.scan != null && SCAN_TRANSIENT_STATUSES.has(release.scan.status));

  return active && Date.now() - new Date(release.createdAt).getTime() < RELEASE_POLL_MAX_AGE_MS;
}

const isReconciling = ref(false);

const { result, loading, error, refetch } = useQuery(
  EnvironmentDocument,
  () => ({ environment: environmentId.value }),
  () => ({ pollInterval: isReconciling.value ? 3000 : 0 }),
);

const { result: volumeUsageResult } = useQuery(
  EnvironmentVolumeUsageDocument,
  () => ({ environment: environmentId.value }),
  { pollInterval: 30000 },
);

const { result: defaultCommandsResult, refetch: refetchDefaultCommands } = useQuery(
  EnvironmentDefaultCommandsDocument,
  () => ({ environment: environmentId.value }),
);

const activeImageDigests = computed(() =>
  (result.value?.environment?.services ?? [])
    .map(s => `${s.id}=${s.activeDeployment?.imageDigest ?? s.activeDeployment?.image ?? ''}`)
    .join('|'),
);

watch(activeImageDigests, (current, previous) => {
  if (previous && current !== previous) {
    refetchDefaultCommands();
  }
});

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
  activeEnvKeyValueStores,
  activeEnvBuckets,
  activeEnvVolumes,
} = useEnvironment();
const { isOpen, currentPanel, closePanel } = usePanel();
const logsPanel = useBuildLogsPanel();
const serviceLogsPanel = useServiceLogsPanel();

watch(currentPanel, (panel, oldPanel) => {
  if (panel?.id !== oldPanel?.id || panel?.type !== oldPanel?.type) {
    serviceLogsPanel.close();
    logsPanel.close();
  }
});

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
        keyValueStores: [],
        buckets: [],
        volumes: [],
      }));
      setEnvironments(shells, environmentId.value);
    }
  },
  { immediate: true },
);

// When the env detail loads, replace the active environment shell with the full payload
watch(
  () => [
    result.value?.environment,
    defaultCommandsResult.value?.environment,
    volumeUsageResult.value?.environment,
  ] as const,
  ([env, commands, usage]) => {
    if (!env) return;

    const defaultCommands = new Map((commands?.services ?? []).map(s => [s.id, s.defaultCommand]));
    const volumeUsage = new Map((usage?.volumes ?? []).map(v => [v.id, volumeUsagePercent(v)]));

    isReconciling.value =
      env.services.some(s => SERVICE_TRANSIENT_STATUSES.has(s.status)) ||
      env.services.some(s => s.activeDeployment?.rollout?.status === RolloutStatus.Progressing) ||
      env.services.some(s => s.releases.some(isReleaseInFlight)) ||
      env.databases.some(d => DATABASE_TRANSIENT_STATUSES.has(d.status)) ||
      env.keyValueStores.some(v => DATABASE_TRANSIENT_STATUSES.has(v.status));

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
        branch: s.branch ?? null,
        autoDeploy: s.autoDeploy,
        contextPath: s.contextPath,
        resources: s.resources,
        command: s.command,
        defaultCommand: defaultCommands.get(s.id) ?? '',
        activeDeployment: s.activeDeployment
          ? {
            id: s.activeDeployment.id,
            image: s.activeDeployment.image,
            imageDigest: s.activeDeployment.imageDigest ?? null,
            commit: s.activeDeployment.commit,
            commitMessage: s.activeDeployment.commitMessage,
            ref: s.activeDeployment.ref,
            status: s.activeDeployment.status,
            replicas: s.activeDeployment.replicas,
            rollout: s.activeDeployment.rollout ?? null,
            createdAt: s.activeDeployment.createdAt,
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
        })),
        builds: s.builds.map(b => ({
          id: b.id,
          status: b.status,
          startedAt: b.startedAt,
          finishedAt: b.finishedAt ?? null,
        })),
        releases: s.releases.map(r => ({
          id: r.id,
          status: r.status,
          createdAt: r.createdAt,
          trigger: {
            kind: r.trigger.kind,
            actor: r.trigger.actor ?? null,
          },
          source: r.source
            ? {
              provider: r.source.provider,
              repository: r.source.repository,
              url: r.source.url,
              ref: r.source.ref,
              contextPath: r.source.contextPath,
              commit: {
                sha: r.source.commit.sha,
                message: r.source.commit.message,
                url: r.source.commit.url ?? null,
              },
            }
            : null,
          build: r.build
            ? {
              id: r.build.id,
              status: r.build.status,
              startedAt: r.build.startedAt,
              finishedAt: r.build.finishedAt ?? null,
            }
            : null,
          deploy: r.deploy
            ? {
              id: r.deploy.id,
              status: r.deploy.status,
              startedAt: r.deploy.startedAt,
              finishedAt: r.deploy.finishedAt ?? null,
            }
            : null,
          scan: r.scan
            ? {
              id: r.scan.id,
              status: r.scan.status,
              findingsCount: r.scan.findingsCount ?? null,
              verifiedCount: r.scan.verifiedCount ?? null,
              startedAt: r.scan.startedAt,
              finishedAt: r.scan.finishedAt ?? null,
            }
            : null,
          deployment: r.deployment
            ? {
              id: r.deployment.id,
              image: r.deployment.image,
              imageDigest: r.deployment.imageDigest ?? null,
              commit: r.deployment.commit,
              commitMessage: r.deployment.commitMessage,
              ref: r.deployment.ref,
              status: r.deployment.status,
              replicas: r.deployment.replicas,
              rollout: r.deployment.rollout ?? null,
              createdAt: r.deployment.createdAt,
            }
            : null,
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
      keyValueStores: env.keyValueStores.map(v => ({
        id: v.id,
        name: v.name,
        version: v.version,
        status: v.status,
        size: v.size,
        createdAt: v.createdAt,
      })),
      buckets: env.buckets.map(b => ({
        id: b.id,
        name: b.name,
        region: b.region,
        endpoint: b.endpoint,
        status: b.status,
        sizeBytes: b.sizeBytes,
        objectCount: b.objectCount,
        createdAt: b.createdAt,
      })),
      volumes: env.volumes.map(v => ({
        id: v.id,
        name: v.name,
        size: v.size,
        mount: v.mount ? { service: v.mount.service, path: v.mount.path } : null,
        usagePercent: volumeUsage.get(v.id) ?? null,
      })),
    };
    setEnvironment(full);
  },
  { immediate: true },
);

function volumeUsagePercent(volume: { size: string; metrics: { points: { value?: number | null }[] }[] }): number | null {
  const points = volume.metrics[0]?.points ?? [];
  const usedBytes = [...points].reverse().find(p => p.value != null)?.value ?? null;
  const capacityBytes = parseStorageSize(volume.size);

  return usedBytes != null && capacityBytes > 0
    ? Math.min(100, Math.round((usedBytes / capacityBytes) * 100))
    : null;
}

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

// Selected key-value store for the panel
const selectedKeyValueStore = computed(() => {
  if (!currentPanel.value || currentPanel.value.type !== 'keyValueStore') return null;
  return activeEnvKeyValueStores.value.find(v => v.id === currentPanel.value!.id) ?? null;
});

// Selected bucket for the panel
const selectedBucket = computed(() => {
  if (!currentPanel.value || currentPanel.value.type !== 'bucket') return null;
  return activeEnvBuckets.value.find(b => b.id === currentPanel.value!.id) ?? null;
});

// Selected volume for the panel
const selectedVolume = computed(() => {
  if (!currentPanel.value || currentPanel.value.type !== 'volume') return null;
  return activeEnvVolumes.value.find(v => v.id === currentPanel.value!.id) ?? null;
});

// Services that already hold a mounted volume (one volume per service)
const mountedServiceIds = computed(() =>
  activeEnvVolumes.value
    .map(v => v.mount?.service)
    .filter((id): id is string => !!id),
);

// Mount volume dialog
const mountDialogOpen = ref(false);
const mountVolumeId = ref<string | null>(null);

const mountVolumeName = computed(
  () => activeEnvVolumes.value.find(v => v.id === mountVolumeId.value)?.name ?? '',
);

function openMountDialog(volumeId: string) {
  mountVolumeId.value = volumeId;
  mountDialogOpen.value = true;
}

// Command palette
const paletteOpen = ref(false);

function handleResourceRemoved() {
  closePanel();
  refetch();
}

function handleCreateFromPalette() {
  refetch();
}

const hasResources = computed(() =>
  activeEnvServices.value.length > 0 ||
  activeEnvDatabases.value.length > 0 ||
  activeEnvKeyValueStores.value.length > 0 ||
  activeEnvBuckets.value.length > 0 ||
  activeEnvVolumes.value.length > 0,
);

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
              :key-value-stores="activeEnvKeyValueStores"
              :buckets="activeEnvBuckets"
              :volumes="activeEnvVolumes"
              :environment-id="environmentId"
              @create="paletteOpen = true"
              @deploy-completed="refetch()"
              @mount-volume="openMountDialog"
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

        <!-- Redis Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedKeyValueStore"
            class="absolute inset-y-3 right-0 w-[55%]"
          >
            <KeyValueStorePanel
              :store="selectedKeyValueStore"
              @close="closePanel"
              @store-removed="handleResourceRemoved"
            />
          </div>
        </Transition>

        <!-- Object Storage Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedBucket"
            class="absolute inset-y-3 right-0 w-[55%]"
          >
            <BucketPanel
              :bucket="selectedBucket"
              @close="closePanel"
              @bucket-removed="handleResourceRemoved"
            />
          </div>
        </Transition>

        <!-- Volume Detail Panel -->
        <Transition name="slide-panel">
          <div
            v-if="isOpen && selectedVolume"
            class="absolute inset-y-3 right-0 w-[55%]"
          >
            <VolumePanel
              :volume="selectedVolume"
              :services="activeEnvServices"
              @close="closePanel"
              @volume-removed="handleResourceRemoved"
              @mount="openMountDialog(selectedVolume.id)"
              @refetch="refetch()"
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
            class="absolute inset-y-3 right-0 z-10"
            style="left: calc(45% + 12px + 2rem)"
          >
            <BuildLogsPanel
              :id="logsPanel.id.value!"
              :kind="logsPanel.kind.value"
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

    <!-- Mount Volume Dialog -->
    <MountVolumeDialog
      v-model:open="mountDialogOpen"
      :volume-id="mountVolumeId"
      :volume-name="mountVolumeName"
      :services="activeEnvServices"
      :mounted-service-ids="mountedServiceIds"
      @mounted="refetch()"
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
