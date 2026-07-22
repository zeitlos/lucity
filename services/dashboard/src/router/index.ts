import { createRouter, createWebHistory } from 'vue-router';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import { useAuth } from '@/composables/useAuth';
import { apolloClient } from '@/lib/apollo';
import { graphql } from '@/gql';

const BootstrapWorkspacesDocument = graphql(`
  query BootstrapWorkspaces {
    workspaces {
      id
    }
  }
`);

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
          path: 'projects/:projectId',
          name: 'project',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: 'projects/:projectId/settings/:section?',
          name: 'project-settings',
          component: () => import('@/pages/ProjectSettingsPage.vue'),
        },
        {
          path: 'projects/:projectId/environments/:environmentId',
          name: 'environment',
          component: () => import('@/pages/EnvironmentPage.vue'),
        },
        {
          path: 'workspace/settings',
          name: 'workspace-settings',
          component: () => import('@/pages/WorkspaceSettingsPage.vue'),
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
      path: '/callback',
      name: 'callback',
      component: () => import('@/pages/CallbackPage.vue'),
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

  if (!activeWorkspace.value) {
    await apolloClient.query({ query: BootstrapWorkspacesDocument, fetchPolicy: 'network-only' }).catch(() => {});
    await fetchUser();
    if (!activeWorkspace.value) {
      return { name: 'login' };
    }
  }
});

export default router;
