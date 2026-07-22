import { ref, computed } from 'vue';
import { logto, apiResource } from '@/lib/logto';

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
const orgIdByWorkspace = new Map<string, string>();

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
      authed.value = await logto.isAuthenticated();
      if (!authed.value) {
        user.value = null;
        return;
      }

      const claims = await logto.getIdTokenClaims();
      const info = await logto.fetchUserInfo();

      const roleByOrg = new Map<string, 'user' | 'admin'>();
      for (const entry of [...(claims.organization_roles ?? []), ...(info.organization_roles ?? [])]) {
        const [orgId, roleName] = entry.split(':');
        if (roleName === 'admin') {
          roleByOrg.set(orgId, 'admin');
        } else if (!roleByOrg.has(orgId)) {
          roleByOrg.set(orgId, 'user');
        }
      }

      orgIdByWorkspace.clear();
      const workspaces: WorkspaceMembership[] = [];
      for (const org of info.organization_data ?? []) {
        orgIdByWorkspace.set(org.name, org.id);
        workspaces.push({ workspace: org.name, role: roleByOrg.get(org.id) ?? 'user' });
      }

      user.value = {
        id: claims.sub,
        name: claims.name ?? null,
        email: claims.email ?? null,
        avatarUrl: claims.picture ?? '',
        workspaces,
      };

      if (!activeWorkspace.value || !workspaces.some(w => w.workspace === activeWorkspace.value)) {
        setActiveWorkspace(workspaces[0]?.workspace || '');
      }

      const traits: Record<string, unknown> = {};
      if (user.value.name) traits.name = user.value.name;
      if (user.value.email) traits.email = user.value.email;
      if (activeWorkspace.value) traits.workspace = activeWorkspace.value;
      identifyUser(user.value.id, traits);
    } catch {
      authed.value = false;
      user.value = null;
    } finally {
      loading.value = false;
    }
  }

  async function login() {
    await logto.signIn({
      redirectUri: `${window.location.origin}${import.meta.env.BASE_URL}callback`,
      directSignIn: { method: 'social', target: 'github' },
    });
  }

  async function logout() {
    user.value = null;
    authed.value = false;
    clearIdentifiedUser();
    await logto.signOut(`${window.location.origin}${import.meta.env.BASE_URL}`);
  }

  async function refreshToken() {
    await fetchUser();
  }

  async function bearerToken(): Promise<string | undefined> {
    const orgId = activeWorkspace.value ? orgIdByWorkspace.get(activeWorkspace.value) : undefined;
    return logto.getAccessToken(apiResource, orgId);
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
    bearerToken,
  };
}
