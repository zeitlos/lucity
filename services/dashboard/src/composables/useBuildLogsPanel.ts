import { ref, computed } from 'vue';

interface LogsPanelState {
  buildId: string;
  serviceName: string;
}

const panelState = ref<LogsPanelState | null>(null);

export function useBuildLogsPanel() {
  const isOpen = computed(() => panelState.value !== null);
  const buildId = computed(() => panelState.value?.buildId ?? null);
  const serviceName = computed(() => panelState.value?.serviceName ?? '');

  function open(buildIdValue: string, serviceNameValue: string) {
    panelState.value = { buildId: buildIdValue, serviceName: serviceNameValue };
  }

  function close() {
    panelState.value = null;
  }

  return { isOpen, buildId, serviceName, open, close };
}
