import { computed } from 'vue';
import { useAuth } from '@/composables/useAuth';
import { fallbackColorIndex, workspaceColorIndexes, workspaceColorStyle } from '@/lib/workspaceColor';

export function useWorkspaceColor() {
  const { user } = useAuth();

  const indexes = computed(() => workspaceColorIndexes(user.value?.workspaces.map(membership => membership.workspace) ?? []));

  function workspaceColor(workspaceId: string) {
    return workspaceColorStyle(indexes.value.get(workspaceId) ?? fallbackColorIndex(workspaceId));
  }

  return { workspaceColor };
}
