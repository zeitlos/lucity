import { ref, computed } from 'vue';

export interface PanelEntry {
  type: 'service' | 'database' | 'deployment';
  id: string;
  label: string;
}

const panelStack = ref<PanelEntry[]>([]);

export function usePanel() {
  const currentPanel = computed(() =>
    panelStack.value.length > 0
      ? panelStack.value[panelStack.value.length - 1]
      : null
  );

  const isOpen = computed(() => panelStack.value.length > 0);

  function openPanel(entry: PanelEntry) {
    // Service-level panels replace the stack (selecting a different service)
    // Sub-views (deployments, etc.) push onto the stack for breadcrumb navigation
    if (entry.type === 'service' || entry.type === 'database') {
      panelStack.value = [entry];
    } else {
      panelStack.value.push(entry);
    }
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
