import { ref, computed } from 'vue';

export interface AuthUser {
  login: string;
  name: string | null;
  email: string | null;
  avatarUrl: string;
}

const user = ref<AuthUser | null>(null);
const loading = ref(true);

export function useAuth() {
  const isAuthenticated = computed(() => user.value !== null);

  async function fetchUser() {
    try {
      const res = await fetch('/auth/me', { credentials: 'include' });
      if (res.ok) {
        user.value = await res.json();
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
    window.location.href = '/auth/github';
  }

  return {
    user,
    loading,
    isAuthenticated,
    fetchUser,
    logout,
    login,
  };
}
