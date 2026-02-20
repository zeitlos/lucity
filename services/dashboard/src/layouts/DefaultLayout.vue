<script setup lang="ts">
import { computed, ref } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { LogOut, Sun, Moon, Settings } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { useTheme } from '@/composables/useTheme';
import BaseLogo from '@/components/BaseLogo.vue';
import ProjectBreadcrumb from '@/components/ProjectBreadcrumb.vue';
import ProjectSettingsDialog from '@/components/ProjectSettingsDialog.vue';
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
const settingsOpen = ref(false);

async function handleLogout() {
  await logout();
  router.push('/login');
}
</script>

<template>
  <div class="relative z-1 flex min-h-screen flex-col">
    <header class="flex h-[52px] shrink-0 items-center justify-between border-b bg-background px-4">
      <!-- Left: Logo + Avatar + Breadcrumb -->
      <div class="flex items-center gap-3">
        <RouterLink
          to="/"
          class="flex items-center transition-opacity hover:opacity-80"
        >
          <BaseLogo :size="24" class="logo-bold" />
        </RouterLink>

        <!-- User avatar next to logo (non-clickable, like Railway) -->
        <template v-if="user">
          <div class="h-4 w-px bg-border" />
          <Avatar class="h-6 w-6">
            <AvatarImage :src="user.avatarUrl" :alt="user.login" />
            <AvatarFallback class="text-[10px]">{{ (user.name || user.login).charAt(0).toUpperCase() }}</AvatarFallback>
          </Avatar>
        </template>

        <ProjectBreadcrumb
          v-if="isProjectRoute && projectId"
          :project-name="projectId"
        />
      </div>

      <!-- Right: Project nav + Theme + User menu -->
      <div class="flex items-center gap-2">
        <Button
          v-if="isProjectRoute"
          variant="ghost"
          size="sm"
          class="text-muted-foreground"
          @click="settingsOpen = true"
        >
          <Settings :size="14" class="mr-1.5" />
          Settings
        </Button>

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

    <ProjectSettingsDialog
      v-if="isProjectRoute && projectId"
      v-model:open="settingsOpen"
      :project-id="projectId"
      :project-name="projectId"
    />
  </div>
</template>

<style scoped>
.logo-bold {
  --primary: var(--foreground);
  --accent: var(--foreground);
  --accent-foreground: var(--foreground);
}
</style>
