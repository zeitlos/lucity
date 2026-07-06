import { computed, watch, type Ref } from 'vue';
import { DeploymentStatus, ReleaseStatus } from '@/gql/graphql';
import type { Service } from '@/composables/useEnvironment';

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

export function useCanvasReleaseStatus(
  services: Ref<Service[]>,
  onCompleted?: () => void,
) {
  const statusMap = computed<Record<string, CanvasReleaseInfo>>(() => {
    const result: Record<string, CanvasReleaseInfo> = {};

    for (const service of services.value) {
      const inFlight = (service.releases ?? []).find(r =>
        IN_FLIGHT_STATUSES.has(r.status)
        && Date.now() - new Date(r.createdAt).getTime() < IN_FLIGHT_MAX_AGE_MS,
      );

      if (inFlight) {
        result[service.id] = {
          phase: releasePhase(inFlight.status, inFlight.deployment?.status),
          startedAt: new Date(inFlight.createdAt).getTime(),
        };
      }
    }

    return result;
  });

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

  watch(statusMap, (current, previous) => {
    if (onCompleted && Object.keys(previous).some(id => !(id in current))) {
      onCompleted();
    }
  });

  return { statusMap };
}
