import { ref, computed } from 'vue';

export interface PanelEntry {
  type: 'service' | 'database' | 'keyValueStore' | 'bucket' | 'volume';
  id: string;
  label: string;
}

const panelStack = ref<PanelEntry[]>([]);

export function usePanel() {
  const currentPanel = computed(() =>
    panelStack.value.length > 0
      ? panelStack.value[panelStack.value.length - 1]
      : null,
  );

  const isOpen = computed(() => panelStack.value.length > 0);

  function openPanel(entry: PanelEntry) {
    panelStack.value = [entry];
  }

  function closePanel() {
    panelStack.value = [];
  }

  function popPanel() {
    panelStack.value.pop();
  }

  return {
    panelStack,
    currentPanel,
    isOpen,
    openPanel,
    closePanel,
    popPanel,
  };
}
