import { ref, computed } from 'vue';

export type LogsPanelKind = 'build' | 'deploy';

interface LogsPanelState {
  id: string;
  kind: LogsPanelKind;
  serviceName: string;
}

const panelState = ref<LogsPanelState | null>(null);

export function useBuildLogsPanel() {
  const isOpen = computed(() => panelState.value !== null);
  const id = computed(() => panelState.value?.id ?? null);
  const kind = computed(() => panelState.value?.kind ?? 'build');
  const serviceName = computed(() => panelState.value?.serviceName ?? '');

  function open(idValue: string, serviceNameValue: string, kindValue: LogsPanelKind = 'build') {
    panelState.value = { id: idValue, kind: kindValue, serviceName: serviceNameValue };
  }

  function close() {
    panelState.value = null;
  }

  return { isOpen, id, kind, serviceName, open, close };
}
