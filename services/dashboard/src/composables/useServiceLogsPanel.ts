import { ref, computed } from 'vue';

interface ServiceLogsPanelState {
  serviceId: string;
  serviceName: string;
}

const panelState = ref<ServiceLogsPanelState | null>(null);

export function useServiceLogsPanel() {
  const isOpen = computed(() => panelState.value !== null);
  const serviceId = computed(() => panelState.value?.serviceId ?? null);
  const serviceName = computed(() => panelState.value?.serviceName ?? '');

  function open(serviceIdValue: string, serviceNameValue: string) {
    panelState.value = {
      serviceId: serviceIdValue,
      serviceName: serviceNameValue,
    };
  }

  function close() {
    panelState.value = null;
  }

  return { isOpen, serviceId, serviceName, open, close };
}
