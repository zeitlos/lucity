import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { DeployLogsDocument } from '@/gql/graphql';

export function useDeployLogs(deployId: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const isActive = ref(false);

  const { onResult, onError, stop, restart } = useSubscription(
    DeployLogsDocument,
    () => ({ id: deployId.value! }),
    () => ({ enabled: !!deployId.value }),
  );

  onResult(({ data }) => {
    if (data?.deployLogs) {
      lines.value.push(data.deployLogs);
      isActive.value = true;
    }
  });

  onError(() => {
    isActive.value = false;
  });

  // Reset when deployId changes.
  watch(deployId, (newId, oldId) => {
    if (newId !== oldId) {
      lines.value = [];
      isActive.value = !!newId;
      if (newId) {
        restart();
      }
    }
  });

  function clear() {
    lines.value = [];
  }

  function reset() {
    lines.value = [];
    isActive.value = false;
    stop();
  }

  return { lines, isActive, clear, stop, restart, reset };
}
