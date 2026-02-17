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
          path: 'new',
          name: 'new-project',
          component: () => import('@/pages/NewProjectPage.vue'),
        },
        {
          path: 'projects/:id',
          name: 'project',
          component: () => import('@/pages/ProjectPage.vue'),
        },
        {
          path: 'projects/:id/environments/:env',
          name: 'environment',
          component: () => import('@/pages/EnvironmentPage.vue'),
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

  const { isAuthenticated, loading, fetchUser } = useAuth();

  if (loading.value) {
    await fetchUser();
  }

  if (!isAuthenticated.value) {
    return { name: 'login' };
  }
});

export default router;
