import { ref, watch, onUnmounted, computed, type Ref } from 'vue';
import { apolloClient } from '@/lib/apollo';
import { ActiveDeploymentQuery } from '@/graphql/services';

export interface CanvasDeployInfo {
  phase: string;
  startedAt: number;
}

export function useCanvasDeployStatus(
  projectId: Ref<string>,
  environment: Ref<string | null>,
  serviceNames: Ref<string[]>,
  onCompleted?: () => void,
) {
  const statusMap = ref<Record<string, CanvasDeployInfo>>({});
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  async function pollAll() {
    const envName = environment.value;
    if (!envName) return;

    const prev = statusMap.value;
    const results: Record<string, CanvasDeployInfo> = {};

    await Promise.allSettled(
      serviceNames.value.map(async (svc) => {
        try {
          const { data } = await apolloClient.query({
            query: ActiveDeploymentQuery,
            variables: { projectId: projectId.value, service: svc, environment: envName },
            fetchPolicy: 'network-only',
          });
          const active = data?.activeDeployment;
          if (active?.id && active.phase !== 'SUCCEEDED' && active.phase !== 'FAILED') {
            results[svc] = {
              phase: active.phase,
              startedAt: active.startedAt ? new Date(active.startedAt).getTime() : Date.now(),
            };
          }
        } catch {
          // No active deployment — ignore
        }
      }),
    );

    // Detect completed deploys: was active, now gone
    const completed = Object.keys(prev).some(svc => !(svc in results));

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

  const hasServices = computed(() => environment.value && serviceNames.value.length > 0);

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
