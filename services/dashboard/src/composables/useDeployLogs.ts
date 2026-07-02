import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const DeployLogsDocument = graphql(`
  subscription DeployLogs($id: DeployID!) {
    deployLogs(id: $id)
  }
`);

export function useDeployLogs(deployId: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const isActive = ref(false);
  const error = ref<string | null>(null);

  const { onResult, onError, stop, restart } = useSubscription(
    DeployLogsDocument,
    () => ({ id: deployId.value! }),
    () => ({ enabled: !!deployId.value }),
  );

  onResult(({ data }) => {
    if (data?.deployLogs) {
      lines.value.push(data.deployLogs);
      isActive.value = true;
      error.value = null;
    }
  });

  onError((err) => {
    isActive.value = false;
    error.value = err.message || 'Failed to load deploy logs';
  });

  watch(deployId, (newId, oldId) => {
    if (newId !== oldId) {
      lines.value = [];
      isActive.value = !!newId;
      error.value = null;
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
    error.value = null;
    stop();
  }

  return { lines, isActive, error, clear, stop, restart, reset };
}
