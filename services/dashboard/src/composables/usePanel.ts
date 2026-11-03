import { ref, computed, watch } from 'vue';
import { confirmDiscard, useGuardedTab } from '@/composables/useUnsavedChanges';

export interface PanelEntry {
  type: 'service' | 'database' | 'keyValueStore' | 'bucket' | 'volume' | 'agent';
  id: string;
  label: string;
  tab?: string;
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

/**
 * A panel's guarded tab that also follows the tab an `openPanel` call asks for.
 */
export function usePanelTab(initial: string) {
  const { currentPanel } = usePanel();
  const { tab, selectTab } = useGuardedTab(currentPanel.value?.tab ?? initial);

  watch(currentPanel, (panel) => {
    if (panel?.tab) selectTab(panel.tab);
  });

  return { tab, selectTab };
}
