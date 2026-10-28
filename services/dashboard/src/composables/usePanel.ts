import { ref, computed } from 'vue';
import { confirmDiscard } from '@/composables/useUnsavedChanges';

export interface PanelEntry {
  type: 'service' | 'database' | 'keyValueStore' | 'bucket' | 'volume' | 'agent';
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

  async function openPanel(entry: PanelEntry) {
    const current = currentPanel.value;
    const switching = current && (current.type !== entry.type || current.id !== entry.id);
    if (switching && !(await confirmDiscard())) return;
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
