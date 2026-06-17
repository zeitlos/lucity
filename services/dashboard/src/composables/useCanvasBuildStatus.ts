import { ref, watch, onUnmounted, computed, type Ref } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';
import { BuildStatus } from '@/gql/graphql';

const CanvasServiceBuildsDocument = graphql(`
  query CanvasServiceBuilds($id: ServiceID!) {
    service(id: $id) {
      id
      builds {
        id
        status
        startedAt
        finishedAt
      }
    }
  }
`);

const TERMINAL_STATUSES = new Set<BuildStatus>([
  BuildStatus.Succeeded,
  BuildStatus.Failed,
  BuildStatus.Cancelled,
]);

export interface CanvasBuildInfo {
  status: BuildStatus;
  startedAt: number;
}

interface CanvasService {
  id: string;
}

export function useCanvasBuildStatus(
  services: Ref<CanvasService[]>,
  onCompleted?: () => void,
) {
  const statusMap = ref<Record<string, CanvasBuildInfo>>({});
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  async function pollAll() {
    const prev = statusMap.value;
    const results: Record<string, CanvasBuildInfo> = {};

    await Promise.allSettled(
      services.value.map(async (svc) => {
        try {
          const { data } = await apolloClient.query({
            query: CanvasServiceBuildsDocument,
            variables: { id: svc.id },
            fetchPolicy: 'network-only',
          });
          const builds = data?.service?.builds ?? [];
          const inFlight = builds.find(b => !TERMINAL_STATUSES.has(b.status));
          if (inFlight) {
            results[svc.id] = {
              status: inFlight.status,
              startedAt: inFlight.startedAt ? new Date(inFlight.startedAt).getTime() : Date.now(),
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
