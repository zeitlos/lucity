import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { graphql } from '@/gql';

const ServiceLogsDocument = graphql(`
  subscription ServiceLogs($service: ServiceID!, $tailLines: Int) {
    serviceLogs(service: $service, tailLines: $tailLines) {
      line
      pod
    }
  }
`);

export interface LogLine {
  line: string;
  pod: string;
}

export function useServiceLogs(
  serviceId: Ref<string>,
  enabled: Ref<boolean>,
) {
  const lines = ref<LogLine[]>([]);
  const isActive = ref(false);
  const error = ref<unknown>(null);

  const { onResult, onError, stop, restart } = useSubscription(
    ServiceLogsDocument,
    () => ({
      service: serviceId.value,
      tailLines: 1000,
    }),
    () => ({ enabled: enabled.value && !!serviceId.value }),
  );

  onResult(({ data }) => {
    if (data?.serviceLogs) {
      lines.value.push(data.serviceLogs);
      isActive.value = true;
      error.value = null;
    }
  });

  onError((e) => {
    isActive.value = false;
    error.value = e;
  });

  watch(serviceId, () => {
    lines.value = [];
    isActive.value = false;
  });

  function clear() {
    lines.value = [];
  }

  function retry() {
    error.value = null;
    restart();
  }

  return { lines, isActive, error, clear, stop, restart, retry };
}
