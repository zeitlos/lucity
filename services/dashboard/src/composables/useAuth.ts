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
      identify: (userId: string, traits?: Record<string, unknown>) => void;
      clearUserId: () => void;
    };
  }
}

const user = ref<AuthUser | null>(null);
const loading = ref(true);
const authed = ref(false);
const activeWorkspace = ref<string>(localStorage.getItem('lucity_workspace') || '');

let identifiedUserId: string | null = null;

function identifyUser(userId: string, traits: Record<string, unknown>) {
  if (identifiedUserId === userId) return;
  identifiedUserId = userId;
  let attempts = 0;
  const run = () => {
    if (window.rybbit) {
      window.rybbit.identify(userId, traits);
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

function setActiveWorkspace(ws: string) {
  activeWorkspace.value = ws;
  localStorage.setItem('lucity_workspace', ws);
}

export function useAuth() {
  const isAuthenticated = computed(() => authed.value);

  async function fetchUser() {
    try {
      const res = await fetch('/auth/me', { credentials: 'include' });
      if (!res.ok) {
        authed.value = false;
        user.value = null;
        return;
      }

      const me = (await res.json()) as AuthUser;
      authed.value = true;
      user.value = me;

      const workspaces = me.workspaces ?? [];
      if (!activeWorkspace.value || !workspaces.some(w => w.workspace === activeWorkspace.value)) {
        setActiveWorkspace(workspaces[0]?.workspace || '');
      }

      const traits: Record<string, unknown> = {};
      if (me.name) traits.name = me.name;
      if (me.email) traits.email = me.email;
      if (activeWorkspace.value) traits.workspace = activeWorkspace.value;
      identifyUser(me.id, traits);
    } catch {
      authed.value = false;
      user.value = null;
    } finally {
      loading.value = false;
    }
  }

  function login() {
    window.location.href = '/auth/login';
  }

  function logout() {
    user.value = null;
    authed.value = false;
    clearIdentifiedUser();
    window.location.href = '/auth/logout';
  }

  return {
    user,
    loading,
    isAuthenticated,
    activeWorkspace,
    fetchUser,
    refreshToken: fetchUser,
    logout,
    login,
    setActiveWorkspace,
  };
}
