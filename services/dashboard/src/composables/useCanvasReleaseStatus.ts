import { computed, watch, type Ref } from 'vue';
import { BuildStatus, DeploymentStatus, DeployStatus, ReleaseStatus, RolloutStatus, ScanStatus } from '@/gql/graphql';
import type { Release, Service } from '@/composables/useEnvironment';
import type { LogsPanelKind } from '@/composables/useBuildLogsPanel';

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

export interface CanvasReleaseFailure {
  label: string;
  at: string;
  logs: { kind: LogsPanelKind; id: string } | null;
}

function describeFailure(release: Release): CanvasReleaseFailure {
  const { build, scan, deploy, deployment } = release;
  if (build?.status === BuildStatus.Failed) {
    return { label: 'Build failed', at: build.finishedAt ?? release.createdAt, logs: { kind: 'build', id: build.id } };
  }
  if (scan?.status === ScanStatus.Findings && scan.verifiedCount) {
    return { label: 'Leaked credentials found', at: scan.finishedAt ?? release.createdAt, logs: { kind: 'scan', id: scan.id } };
  }
  if (scan?.status === ScanStatus.Failed) {
    return { label: 'Secret scan failed', at: scan.finishedAt ?? release.createdAt, logs: { kind: 'scan', id: scan.id } };
  }
  if (deploy?.status === DeployStatus.Failed) {
    return { label: 'Deploy failed', at: deploy.finishedAt ?? release.createdAt, logs: { kind: 'deploy', id: deploy.id } };
  }
  if (deployment?.rollout?.status === RolloutStatus.Failed) {
    return { label: 'Rollout failed', at: deployment.rollout.startedAt ?? release.createdAt, logs: null };
  }
  return { label: 'Release failed', at: release.createdAt, logs: null };
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

  const failureMap = computed<Record<string, CanvasReleaseFailure>>(() => {
    const result: Record<string, CanvasReleaseFailure> = {};

    for (const service of services.value) {
      const latest = [...(service.releases ?? [])].sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))[0];
      if (latest?.status === ReleaseStatus.Failed) {
        result[service.id] = describeFailure(latest);
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

  return { statusMap, failureMap };
}
