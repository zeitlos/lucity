<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { LogOut, Sun, Moon } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { useTheme } from '@/composables/useTheme';
import BaseLogo from '@/components/BaseLogo.vue';
import ProjectBreadcrumb from '@/components/ProjectBreadcrumb.vue';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const route = useRoute();
const router = useRouter();
const { user, logout } = useAuth();
const { theme, toggleTheme } = useTheme();

const isProjectRoute = computed(() => route.name === 'project');
const projectId = computed(() => route.params.id as string | undefined);

async function handleLogout() {
  await logout();
  router.push('/login');
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <header class="flex h-[52px] shrink-0 items-center justify-between border-b bg-background px-4">
      <!-- Left: Logo + Breadcrumb -->
      <div class="flex items-center gap-3">
        <RouterLink
          to="/"
          class="flex items-center gap-2 transition-opacity hover:opacity-80"
        >
          <BaseLogo :size="24" />
          <span class="text-sm font-semibold text-foreground">Lucity</span>
        </RouterLink>

        <ProjectBreadcrumb
          v-if="isProjectRoute && projectId"
          :project-name="projectId"
          class="ml-2"
        />
      </div>

      <!-- Right: Theme + User -->
      <div class="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8"
          @click="toggleTheme"
        >
          <Sun v-if="theme === 'dark'" :size="16" />
          <Moon v-else :size="16" />
        </Button>

        <DropdownMenu v-if="user">
          <DropdownMenuTrigger as-child>
            <button class="rounded-full transition-opacity hover:opacity-80">
              <Avatar class="h-7 w-7">
                <AvatarImage :src="user.avatarUrl" :alt="user.login" />
                <AvatarFallback>{{ (user.name || user.login).charAt(0).toUpperCase() }}</AvatarFallback>
              </Avatar>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-48">
            <div class="px-2 py-1.5">
              <p class="text-sm font-medium">{{ user.name || user.login }}</p>
              <p class="text-xs text-muted-foreground">{{ user.login }}</p>
            </div>
            <DropdownMenuSeparator />
            <DropdownMenuItem @select="handleLogout">
              <LogOut :size="14" class="mr-2" />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>

    <main class="flex-1 bg-background">
      <RouterView />
    </main>
  </div>
</template>
