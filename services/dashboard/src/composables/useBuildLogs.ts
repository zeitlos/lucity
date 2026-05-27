import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const BuildLogsDocument = graphql(`
  subscription BuildLogs($id: String!) {
    buildLogs(id: $id)
  }
`);

export function useBuildLogs(buildId: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const isActive = ref(false);

  const { onResult, onError, stop, restart } = useSubscription(
    BuildLogsDocument,
    () => ({ id: buildId.value! }),
    () => ({ enabled: !!buildId.value }),
  );

  onResult(({ data }) => {
    if (data?.buildLogs) {
      lines.value.push(data.buildLogs);
      isActive.value = true;
    }
  });

  onError(() => {
    isActive.value = false;
  });

  // Reset when buildId changes.
  watch(buildId, (newId, oldId) => {
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
