import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const BuildLogsDocument = graphql(`
  subscription BuildLogs($id: BuildID!) {
    buildLogs(id: $id)
  }
`);

export function useBuildLogs(buildId: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const isActive = ref(false);
  const error = ref<string | null>(null);

  const { onResult, onError, stop, restart } = useSubscription(
    BuildLogsDocument,
    () => ({ id: buildId.value! }),
    () => ({ enabled: !!buildId.value }),
  );

  onResult(({ data }) => {
    if (data?.buildLogs) {
      lines.value.push(data.buildLogs);
      isActive.value = true;
      error.value = null;
    }
  });

  onError((err) => {
    isActive.value = false;
    error.value = err.message || 'Failed to load build logs';
  });

  // Reset when buildId changes.
  watch(buildId, (newId, oldId) => {
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
