import { createRouter, createWebHistory } from 'vue-router';
import { confirmDiscard } from '@/composables/useUnsavedChanges';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import { useAuth } from '@/composables/useAuth';

const legacyWorkspaceTabs: Record<string, string> = {
  general: 'settings',
  members: 'members',
  apitokens: 'api-tokens',
  billing: 'billing',
  danger: 'settings',
};

const ProjectsPage = () => import('@/pages/ProjectsPage.vue');

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
          component: ProjectsPage,
        },
        {
          path: ':drawer(members|api-tokens|billing|settings)',
          name: 'workspace',
          component: ProjectsPage,
        },
        {
          path: 'projects/:projectId/:drawer(settings)?',
          name: 'project',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: 'projects/:projectId/environments/:environmentId/:drawer(variables|settings)?',
          name: 'environment',
          component: () => import('@/pages/EnvironmentPage.vue'),
        },
        {
          path: 'workspace/settings',
          redirect: to => ({
            name: 'workspace',
            params: { drawer: legacyWorkspaceTabs[String(to.query.tab)] ?? 'settings' },
          }),
        },
      ],
    },
    {
      path: '/checkout/success',
      name: 'checkout-success',
      meta: { requiresAuth: true },
      component: () => import('@/pages/CheckoutSuccessPage.vue'),
    },
    {
      path: '/checkout/plan-success',
      name: 'plan-checkout-success',
      meta: { requiresAuth: true },
      component: () => import('@/pages/PlanCheckoutSuccessPage.vue'),
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

router.beforeEach(() => confirmDiscard());

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return;

  const { isAuthenticated, loading, fetchUser, activeWorkspace } = useAuth();

  if (loading.value) {
    await fetchUser();
  }

  if (!isAuthenticated.value) {
    return { name: 'login' };
  }

  if (!activeWorkspace.value) {
    return { name: 'login', query: { error: 'no_workspace' } };
  }
});

export default router;
