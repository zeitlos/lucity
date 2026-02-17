<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { LayoutDashboard, Settings, LogOut } from 'lucide-vue-next';
import { cn } from '@/lib/utils';
import { useAuth } from '@/composables/useAuth';

const route = useRoute();
const router = useRouter();
const { user, logout } = useAuth();

const navItems = [
  { label: 'Projects', route: '/', icon: LayoutDashboard },
  { label: 'Settings', route: '/settings', icon: Settings },
];

function isActive(path: string) {
  if (path === '/') return route.path === '/' || route.path.startsWith('/projects');
  return route.path.startsWith(path);
}

async function handleLogout() {
  await logout();
  router.push('/login');
}
</script>

<template>
  <div class="flex min-h-screen">
    <aside class="flex w-64 flex-col border-r bg-white px-4 py-6">
      <div class="mb-8 px-2">
        <h1 class="text-xl font-bold text-gray-900">Lucity</h1>
        <p class="text-xs text-gray-500">PaaS Dashboard</p>
      </div>

      <nav class="flex-1 space-y-1">
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

      <div
        v-if="user"
        class="border-t pt-4"
      >
        <div class="flex items-center gap-3 px-2">
          <img
            :src="user.avatarUrl"
            :alt="user.login"
            class="h-8 w-8 rounded-full"
          >
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-gray-900">
              {{ user.name || user.login }}
            </p>
            <p class="truncate text-xs text-gray-500">
              {{ user.login }}
            </p>
          </div>
          <button
            class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            title="Sign out"
            @click="handleLogout"
          >
            <LogOut :size="16" />
          </button>
        </div>
      </div>
    </aside>

    <main class="flex-1 bg-gray-50">
      <RouterView />
    </main>
  </div>
</template>
