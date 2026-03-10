import { ref, computed } from 'vue';

export interface WorkspaceMembership {
  workspace: string;
  role: 'user' | 'admin';
}

export interface AuthUser {
  name: string | null;
  email: string | null;
  avatarUrl: string;
  workspaces: WorkspaceMembership[];
}

const user = ref<AuthUser | null>(null);
const loading = ref(true);
const activeWorkspace = ref<string>(localStorage.getItem('lucity_workspace') || '');

export function useAuth() {
  const isAuthenticated = computed(() => user.value !== null);

  async function fetchUser() {
    try {
      const res = await fetch('/auth/me', { credentials: 'include' });
      if (res.ok) {
        user.value = await res.json();

        // If JWT has no workspace claims (e.g., minted before workspace support),
        // refresh the token to pick up current Rauthy groups.
        if (user.value && user.value.workspaces.length === 0) {
          const refreshRes = await fetch('/auth/refresh', {
            method: 'POST',
            credentials: 'include',
          });
          if (refreshRes.ok) {
            const meRes = await fetch('/auth/me', { credentials: 'include' });
            if (meRes.ok) {
              user.value = await meRes.json();
            }
          }
        }

        // Default to first workspace if none selected or stale
        if (user.value && (!activeWorkspace.value || !user.value.workspaces.some(w => w.workspace === activeWorkspace.value))) {
          setActiveWorkspace(user.value.workspaces[0]?.workspace || '');
        }
      } else {
        user.value = null;
      }
    } catch {
      user.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function logout() {
    await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
    user.value = null;
  }

  function login() {
    window.location.href = '/auth/login';
  }

  function setActiveWorkspace(ws: string) {
    activeWorkspace.value = ws;
    localStorage.setItem('lucity_workspace', ws);
  }

  async function refreshToken() {
    try {
      const res = await fetch('/auth/refresh', {
        method: 'POST',
        credentials: 'include',
      });
      if (res.ok) {
        await fetchUser();
      }
    } catch {
      // Token refresh failed — user will need to re-login on next protected action
    }
  }

  return {
    user,
    loading,
    isAuthenticated,
    activeWorkspace,
    fetchUser,
    logout,
    login,
    setActiveWorkspace,
    refreshToken,
  };
}
