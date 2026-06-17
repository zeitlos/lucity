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
    }
  });

  onError(() => {
    isActive.value = false;
  });

  watch(serviceId, () => {
    lines.value = [];
    isActive.value = false;
  });

  function clear() {
    lines.value = [];
  }

  return { lines, isActive, clear, stop, restart };
}
