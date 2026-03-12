import { createRouter, createWebHistory } from 'vue-router';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import { useAuth } from '@/composables/useAuth';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: DefaultLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'projects',
          component: () => import('@/pages/ProjectsPage.vue'),
        },
        {
          path: 'projects/:id/settings/:section?',
          name: 'project-settings',
          component: () => import('@/pages/ProjectSettingsPage.vue'),
        },
        {
          path: 'projects/:id/:env',
          name: 'project-env',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: 'projects/:id',
          name: 'project',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: 'workspace/settings',
          name: 'workspace-settings',
          component: () => import('@/pages/WorkspaceSettingsPage.vue'),
        },
      ],
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
    },
    {
      path: '/brand',
      name: 'brand',
      component: () => import('@/pages/BrandPage.vue'),
    },
  ],
});

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return;

  const { isAuthenticated, loading, fetchUser, activeWorkspace, login } = useAuth();

  if (loading.value) {
    await fetchUser();
  }

  if (!isAuthenticated.value) {
    return { name: 'login' };
  }

  // If no valid workspace could be resolved (stale JWT, removed from workspace),
  // force a full re-login to get fresh OIDC claims.
  if (!activeWorkspace.value) {
    login();
    return false;
  }
});

export default router;
