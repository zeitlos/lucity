import { ref, watch, type Ref } from 'vue';
import { useSubscription } from '@vue/apollo-composable';
import { ServiceLogsSubscription } from '@/graphql/services';

export interface LogLine {
  line: string;
  pod: string;
}

export function useServiceLogs(
  projectId: Ref<string>,
  service: Ref<string>,
  environment: Ref<string | null>,
  enabled: Ref<boolean>,
) {
  const lines = ref<LogLine[]>([]);
  const isActive = ref(false);

  const { onResult, onError, stop, restart } = useSubscription(
    ServiceLogsSubscription,
    () => ({
      projectId: projectId.value,
      service: service.value,
      environment: environment.value!,
      tailLines: 100,
    }),
    () => ({ enabled: enabled.value && !!environment.value }),
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

  watch([projectId, service, environment], () => {
    lines.value = [];
    isActive.value = false;
  });

  function clear() {
    lines.value = [];
  }

  return { lines, isActive, clear, stop, restart };
}
