import { ref, watch, onUnmounted, computed, type Ref } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';
import { DeploymentStatus, ReleaseStatus } from '@/gql/graphql';

const CanvasServiceReleasesDocument = graphql(`
  query CanvasServiceReleases($id: ServiceID!) {
    service(id: $id) {
      id
      releases {
        id
        status
        createdAt
        deployment {
          id
          status
        }
      }
    }
  }
`);

const IN_FLIGHT_STATUSES = new Set<ReleaseStatus>([
  ReleaseStatus.Queued,
  ReleaseStatus.Building,
  ReleaseStatus.Deploying,
]);

const IN_FLIGHT_MAX_AGE_MS = 60 * 60 * 1000;

export type CanvasReleasePhase = 'queued' | 'building' | 'deploying' | 'rollout';

export interface CanvasReleaseInfo {
  phase: CanvasReleasePhase;
  startedAt: number;
}

interface CanvasService {
  id: string;
}

export function useCanvasReleaseStatus(
  services: Ref<CanvasService[]>,
  onCompleted?: () => void,
) {
  const statusMap = ref<Record<string, CanvasReleaseInfo>>({});
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  async function pollAll() {
    const prev = statusMap.value;
    const results: Record<string, CanvasReleaseInfo> = {};

    await Promise.allSettled(
      services.value.map(async (svc) => {
        try {
          const { data } = await apolloClient.query({
            query: CanvasServiceReleasesDocument,
            variables: { id: svc.id },
            fetchPolicy: 'network-only',
          });
          const releases = data?.service?.releases ?? [];
          const inFlight = releases.find(r =>
            IN_FLIGHT_STATUSES.has(r.status)
            && Date.now() - new Date(r.createdAt).getTime() < IN_FLIGHT_MAX_AGE_MS,
          );
          if (inFlight) {
            results[svc.id] = {
              phase: releasePhase(inFlight.status, inFlight.deployment?.status),
              startedAt: inFlight.createdAt ? new Date(inFlight.createdAt).getTime() : Date.now(),
            };
          }
        } catch {
          // Service query failed — ignore for this tick.
        }
      }),
    );

    const completed = Object.keys(prev).some(id => !(id in results));

    statusMap.value = results;

    if (completed && onCompleted) {
      onCompleted();
    }
  }

  function releasePhase(status: ReleaseStatus, deploymentStatus?: DeploymentStatus): CanvasReleasePhase {
    switch (status) {
      case ReleaseStatus.Queued:
        return 'queued';
      case ReleaseStatus.Building:
        return 'building';
      default:
        return deploymentStatus === DeploymentStatus.Deploying ? 'rollout' : 'deploying';
    }
  }

  function startPolling() {
    stopPolling();
    pollAll();
    pollTimer = setInterval(pollAll, 3000);
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  }

  const hasServices = computed(() => services.value.length > 0);

  watch(hasServices, (active) => {
    if (active) {
      startPolling();
    } else {
      stopPolling();
      statusMap.value = {};
    }
  }, { immediate: true });

  onUnmounted(stopPolling);

  return { statusMap };
}
