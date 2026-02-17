<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { LayoutDashboard, Settings, LogOut, Sun, Moon } from 'lucide-vue-next';
import { cn } from '@/lib/utils';
import { useAuth } from '@/composables/useAuth';
import { useTheme } from '@/composables/useTheme';

const route = useRoute();
const router = useRouter();
const { user, logout } = useAuth();
const { theme, toggleTheme } = useTheme();

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
    <aside class="flex w-64 flex-col border-r border-sidebar-border bg-sidebar px-4 py-6">
      <div class="mb-8 px-2">
        <h1 class="text-xl font-bold text-sidebar-foreground">Lucity</h1>
        <p class="text-xs text-muted-foreground">PaaS Dashboard</p>
      </div>

      <nav class="flex-1 space-y-1">
        <RouterLink
          v-for="item in navItems"
          :key="item.route"
          :to="item.route"
          :class="cn(
            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
            isActive(item.route)
              ? 'bg-sidebar-accent text-sidebar-foreground'
              : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
          )"
        >
          <component :is="item.icon" :size="18" />
          {{ item.label }}
        </RouterLink>
      </nav>

      <div class="mb-4 border-t border-sidebar-border pt-4">
        <button
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-sidebar-foreground/70 transition-colors hover:bg-sidebar-accent/50 hover:text-sidebar-foreground"
          @click="toggleTheme"
        >
          <Sun v-if="theme === 'dark'" :size="18" />
          <Moon v-else :size="18" />
          {{ theme === 'dark' ? 'Light Mode' : 'Dark Mode' }}
        </button>
      </div>

      <div
        v-if="user"
        class="border-t border-sidebar-border pt-4"
      >
        <div class="flex items-center gap-3 px-2">
          <img
            :src="user.avatarUrl"
            :alt="user.login"
            class="h-8 w-8 rounded-full"
          >
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-sidebar-foreground">
              {{ user.name || user.login }}
            </p>
            <p class="truncate text-xs text-muted-foreground">
              {{ user.login }}
            </p>
          </div>
          <button
            class="rounded p-1 text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground"
            title="Sign out"
            @click="handleLogout"
          >
            <LogOut :size="16" />
          </button>
        </div>
      </div>
    </aside>

    <main class="flex-1 bg-background">
      <RouterView />
    </main>
  </div>
</template>
