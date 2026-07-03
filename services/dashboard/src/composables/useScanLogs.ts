import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const ScanLogsDocument = graphql(`
  subscription ScanLogs($id: ScanID!) {
    scanLogs(id: $id)
  }
`);

export function useScanLogs(scanId: Ref<string | null>) {
  const lines = ref<string[]>([]);
  const isActive = ref(false);
  const error = ref<string | null>(null);

  const { onResult, onError, stop, restart } = useSubscription(
    ScanLogsDocument,
    () => ({ id: scanId.value! }),
    () => ({ enabled: !!scanId.value }),
  );

  onResult(({ data }) => {
    if (data?.scanLogs) {
      lines.value.push(data.scanLogs);
      isActive.value = true;
      error.value = null;
    }
  });

  onError((err) => {
    isActive.value = false;
    error.value = err.message || 'Failed to load scan logs';
  });

  watch(scanId, (newId, oldId) => {
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
