import { ref, computed } from 'vue';

interface ServiceLogsPanelState {
  projectId: string;
  serviceName: string;
  environment: string;
}

const panelState = ref<ServiceLogsPanelState | null>(null);

export function useServiceLogsPanel() {
  const isOpen = computed(() => panelState.value !== null);
  const projectId = computed(() => panelState.value?.projectId ?? null);
  const serviceName = computed(() => panelState.value?.serviceName ?? '');
  const environment = computed(() => panelState.value?.environment ?? null);

  function open(projectIdValue: string, serviceNameValue: string, environmentValue: string) {
    panelState.value = {
      projectId: projectIdValue,
      serviceName: serviceNameValue,
      environment: environmentValue,
    };
  }

  function close() {
    panelState.value = null;
  }

  return { isOpen, projectId, serviceName, environment, open, close };
}
