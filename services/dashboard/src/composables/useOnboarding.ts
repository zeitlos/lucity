import { ref, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuth } from './useAuth';

const welcomeDismissed = ref(false);

export function useOnboarding() {
  const route = useRoute();
  const router = useRouter();
  const { activeWorkspace } = useAuth();

  const isWelcome = computed(() => route.query.welcome === 'true' && !welcomeDismissed.value);

  // Checklist dismissed via localStorage per workspace
  const checklistDismissed = computed(() => {
    if (!activeWorkspace.value) return false;
    return localStorage.getItem(`lucity_onboarding_${activeWorkspace.value}_dismissed`) === 'true';
  });

  function dismissWelcome() {
    welcomeDismissed.value = true;
    router.replace({ query: {} });
  }

  function dismissChecklist() {
    if (activeWorkspace.value) {
      localStorage.setItem(`lucity_onboarding_${activeWorkspace.value}_dismissed`, 'true');
    }
  }

  interface ChecklistState {
    githubConnected: boolean;
    hasProjects: boolean;
    hasDeployments: boolean;
  }

  function checklistComplete(state: ChecklistState): boolean {
    return state.githubConnected && state.hasProjects && state.hasDeployments;
  }

  function completedCount(state: ChecklistState): number {
    return [state.githubConnected, state.hasProjects, state.hasDeployments].filter(Boolean).length;
  }

  // Auto-dismiss welcome query param after first render cycle
  watch(() => route.query.welcome, (val) => {
    if (val === 'true') {
      welcomeDismissed.value = false;
    }
  });

  return {
    isWelcome,
    checklistDismissed,
    dismissWelcome,
    dismissChecklist,
    checklistComplete,
    completedCount,
  };
}
