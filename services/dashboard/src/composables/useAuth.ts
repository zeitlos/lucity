import { ref, computed } from 'vue';

export interface WorkspaceMembership {
  workspace: string;
  role: 'user' | 'admin';
}

export interface AuthUser {
  id: string;
  name: string | null;
  email: string | null;
  avatarUrl: string;
  workspaces: WorkspaceMembership[];
}

declare global {
  interface Window {
    rybbit?: {
      identify: (userId: string) => void;
      clearUserId: () => void;
    };
  }
}

const user = ref<AuthUser | null>(null);
const loading = ref(true);
const activeWorkspace = ref<string>(localStorage.getItem('lucity_workspace') || '');

let identifiedUserId: string | null = null;

function identifyUser(userId: string) {
  if (identifiedUserId === userId) return;
  identifiedUserId = userId;
  let attempts = 0;
  const run = () => {
    if (window.rybbit) {
      window.rybbit.identify(userId);
    } else if (attempts++ < 50) {
      setTimeout(run, 100);
    }
  };
  run();
}

function clearIdentifiedUser() {
  identifiedUserId = null;
  window.rybbit?.clearUserId();
}

export function useAuth() {
  const isAuthenticated = computed(() => user.value !== null);

  async function fetchUser() {
    try {
      const res = await fetch('/auth/me', { credentials: 'include' });
      if (res.ok) {
        user.value = await res.json();

        // If JWT has no workspace claims (e.g., minted before workspace support),
        // refresh the token to pick up current OIDC claims.
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

        if (user.value?.id) {
          identifyUser(user.value.id);
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
    clearIdentifiedUser();
  }

  function login() {
    const now = Date.now();
    const last = Number(sessionStorage.getItem('lucity_last_login_redirect') || '0');
    if (now - last < 10_000) {
      return;
    }
    sessionStorage.setItem('lucity_last_login_redirect', String(now));
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
