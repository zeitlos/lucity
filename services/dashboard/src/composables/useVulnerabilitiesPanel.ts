import { ref, computed } from 'vue';

interface VulnerabilitiesPanelState {
  serviceId: string;
  serviceName: string;
}

const panelState = ref<VulnerabilitiesPanelState | null>(null);

export function useVulnerabilitiesPanel() {
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
