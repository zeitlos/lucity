import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'projects',
      component: () => import('@/pages/ProjectsPage.vue'),
    },
    {
      path: '/projects/:id',
      name: 'project',
      component: () => import('@/pages/ProjectPage.vue'),
    },
    {
      path: '/projects/:id/environments/:env',
      name: 'environment',
      component: () => import('@/pages/EnvironmentPage.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
    },
  ],
});

export default router;
