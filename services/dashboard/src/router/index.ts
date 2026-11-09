import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router';
import { confirmDiscard } from '@/composables/useUnsavedChanges';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import { useAuth } from '@/composables/useAuth';
import { apolloClient } from '@/lib/apollo';
import { environmentRoute, projectRoute } from '@/lib/routes';

const legacyWorkspaceTabs: Record<string, string> = {
  general: 'settings',
  members: 'members',
  apitokens: 'api-tokens',
  billing: 'billing',
  danger: 'settings',
};

const ProjectsPage = () => import('@/pages/ProjectsPage.vue');

function activeWorkspaceId() {
  return useAuth().activeWorkspace.value;
}

function legacyDrawer(to: RouteLocationNormalized) {
  return typeof to.params.drawer === 'string' && to.params.drawer ? to.params.drawer : undefined;
}

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
          name: 'home',
          component: ProjectsPage,
          beforeEnter: to => ({ name: 'projects', params: { workspace: activeWorkspaceId() }, query: to.query }),
        },
        {
          path: ':workspace',
          name: 'projects',
          component: ProjectsPage,
        },
        {
          path: ':workspace/~/:drawer(members|connections|api-tokens|billing|settings)',
          name: 'workspace',
          component: ProjectsPage,
        },
        {
          path: ':workspace/~/connections/:kind(model-providers|connectors|chat-apps)/:name',
          name: 'connection',
          component: ProjectsPage,
        },
        {
          path: ':workspace/~/add-channel',
          name: 'add-channel',
          component: ProjectsPage,
        },
        {
          path: ':workspace/:project/:tilde(~)?/:drawer(settings)?',
          name: 'project',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: ':workspace/:project/:environment/:tilde(~)?/:drawer(variables|settings)?',
          name: 'environment',
          component: () => import('@/pages/EnvironmentPage.vue'),
        },
        {
          path: 'projects/:projectId/:drawer(settings)?',
          component: ProjectsPage,
          beforeEnter: to => projectRoute(String(to.params.projectId), legacyDrawer(to)),
        },
        {
          path: 'projects/:projectId/environments/:environmentId/:drawer(variables|settings)?',
          component: ProjectsPage,
          beforeEnter: to => environmentRoute(String(to.params.environmentId), legacyDrawer(to)),
        },
        {
          path: ':drawer(members|connections|api-tokens|billing|settings)',
          component: ProjectsPage,
          beforeEnter: to => ({ name: 'workspace', params: { workspace: activeWorkspaceId(), drawer: legacyDrawer(to) } }),
        },
        {
          path: 'connections/:kind(model-providers|connectors|chat-apps)/:name',
          component: ProjectsPage,
          beforeEnter: to => ({ name: 'connection', params: { workspace: activeWorkspaceId(), kind: String(to.params.kind), name: String(to.params.name) } }),
        },
        {
          path: 'workspace/settings',
          component: ProjectsPage,
          beforeEnter: to => ({
            name: 'workspace',
            params: { workspace: activeWorkspaceId(), drawer: legacyWorkspaceTabs[String(to.query.tab)] ?? 'settings' },
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
      path: '/checkout/card-success',
      name: 'card-checkout-success',
      meta: { requiresAuth: true },
      component: () => import('@/pages/CardCheckoutSuccessPage.vue'),
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

router.beforeEach((to) => {
  if (!to.params.drawer || to.params.tilde) return;

  if (to.name === 'project') {
    return { name: 'environment', params: { workspace: to.params.workspace, project: to.params.project, environment: to.params.drawer } };
  }

  if (to.name === 'environment') {
    return { name: 'environment', params: { workspace: to.params.workspace, project: to.params.project, environment: to.params.environment } };
  }
});

let workspaceChanged = false;

const returnKey = 'lucity_return_to';

function rememberReturn(path: string) {
  if (path === '/') return;
  try {
    sessionStorage.setItem(returnKey, path);
  } catch {
    return;
  }
}

function takeReturn(): string | null {
  try {
    const path = sessionStorage.getItem(returnKey);
    sessionStorage.removeItem(returnKey);
    return path;
  } catch {
    return null;
  }
}

router.beforeEach(async (to) => {
  if (!to.meta.requiresAuth) return;

  const { isAuthenticated, loading, fetchUser, activeWorkspace } = useAuth();

  if (loading.value) {
    await fetchUser();
  }

  if (!isAuthenticated.value) {
    rememberReturn(to.fullPath);
    return { name: 'login' };
  }

  if (!activeWorkspace.value) {
    return { name: 'login', query: { error: 'no_workspace' } };
  }

  const returnTo = takeReturn();

  if (returnTo && returnTo !== to.fullPath) {
    return returnTo;
  }

  const workspace = to.params.workspace;

  if (typeof workspace === 'string' && workspace && workspace !== activeWorkspace.value) {
    const { user, setActiveWorkspace } = useAuth();

    if (!user.value?.workspaces.some(membership => membership.workspace === workspace)) {
      return { name: 'projects', params: { workspace: activeWorkspace.value } };
    }

    setActiveWorkspace(workspace);
    workspaceChanged = true;
  }
});

router.afterEach(() => {
  if (!workspaceChanged) return;
  workspaceChanged = false;
  apolloClient.resetStore();
});

export default router;
