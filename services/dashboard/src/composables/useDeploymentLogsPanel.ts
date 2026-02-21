import { ref, computed } from 'vue';

interface LogsPanelState {
  deployId: string;
  serviceName: string;
}

const panelState = ref<LogsPanelState | null>(null);

export function useDeploymentLogsPanel() {
  const isOpen = computed(() => panelState.value !== null);
  const deployId = computed(() => panelState.value?.deployId ?? null);
  const serviceName = computed(() => panelState.value?.serviceName ?? '');

  function open(deployIdValue: string, serviceNameValue: string) {
    panelState.value = { deployId: deployIdValue, serviceName: serviceNameValue };
  }

  function close() {
    panelState.value = null;
  }

  return { isOpen, deployId, serviceName, open, close };
}
