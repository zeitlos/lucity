<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from 'vue-router';
import { LayoutDashboard, Settings } from 'lucide-vue-next';
import { cn } from '@/lib/utils';

const route = useRoute();

const navItems = [
  { label: 'Projects', route: '/', icon: LayoutDashboard },
  { label: 'Settings', route: '/settings', icon: Settings },
];

function isActive(path: string) {
  if (path === '/') return route.path === '/' || route.path.startsWith('/projects');
  return route.path.startsWith(path);
}
</script>

<template>
  <div class="flex min-h-screen">
    <aside class="w-64 border-r bg-white px-4 py-6">
      <div class="mb-8 px-2">
        <h1 class="text-xl font-bold text-gray-900">Lucity</h1>
        <p class="text-xs text-gray-500">PaaS Dashboard</p>
      </div>

      <nav class="space-y-1">
        <RouterLink
          v-for="item in navItems"
          :key="item.route"
          :to="item.route"
          :class="cn(
            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
            isActive(item.route)
              ? 'bg-gray-100 text-gray-900'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
          )"
        >
          <component :is="item.icon" :size="18" />
          {{ item.label }}
        </RouterLink>
      </nav>
    </aside>

    <main class="flex-1 bg-gray-50">
      <RouterView />
    </main>
  </div>
</template>
