<script setup lang="ts">
import { computed, ref } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { Download, LogOut, Settings } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { WorkspaceQuery } from '@/graphql/workspaces';
import BaseLogo from '@/components/BaseLogo.vue';
import ContextNav from '@/components/ContextNav.vue';
import ThemeToggle from '@/components/ThemeToggle.vue';
import HelpMenu from '@/components/HelpMenu.vue';
import ProjectEjectDialog from '@/components/ProjectEjectDialog.vue';
import SuspensionBanner from '@/components/SuspensionBanner.vue';
import TrialIndicator from '@/components/TrialIndicator.vue';
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
const { result: wsResult } = useQuery(WorkspaceQuery, null, { fetchPolicy: 'cache-and-network' });
const suspended = computed(() => wsResult.value?.workspace?.suspended ?? false);

const isProjectRoute = computed(() =>
  route.name === 'project' || route.name === 'project-env' || route.name === 'project-settings',
);
const projectId = computed(() => route.params.id as string | undefined);
const ejectOpen = ref(false);

async function handleLogout() {
  await logout();
  router.push('/login');
}
</script>

<template>
  <div class="relative z-1 flex min-h-screen flex-col">
    <SuspensionBanner v-if="suspended" />

    <div class="relative z-1 flex flex-1 flex-col overflow-hidden p-3 pb-0">
      <header class="flex h-[52px] shrink-0 items-center justify-between rounded-lg border bg-card/80 px-4 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
        <!-- Left: Logo + Context Nav -->
        <div class="flex items-center gap-3">
          <RouterLink
            to="/"
            class="flex items-center transition-opacity hover:opacity-80"
          >
            <BaseLogo :size="24" variant="mark" />
          </RouterLink>

          <template v-if="user">
            <div class="h-4 w-px bg-border" />
            <ContextNav />
          </template>
        </div>

        <!-- Right: Project nav + Theme + User menu -->
        <div class="flex items-center gap-2">
          <Button
            v-if="isProjectRoute"
            variant="ghost"
            size="sm"
            class="text-muted-foreground"
            @click="ejectOpen = true"
          >
            <Download :size="14" class="mr-1.5" />
            Eject
          </Button>

          <Button
            v-if="isProjectRoute && projectId"
            variant="ghost"
            size="sm"
            class="text-muted-foreground"
            @click="router.push({ name: 'project-settings', params: { id: projectId } })"
          >
            <Settings :size="14" class="mr-1.5" />
            Settings
          </Button>

          <TrialIndicator v-if="user && !suspended" />

          <ThemeToggle />
          <HelpMenu />

          <DropdownMenu v-if="user">
            <DropdownMenuTrigger as-child>
              <button class="rounded-full transition-opacity hover:opacity-80">
                <Avatar class="h-7 w-7">
                  <AvatarImage :src="user.avatarUrl" :alt="user.name || user.email || ''" />
                  <AvatarFallback>{{ (user.name || user.email || '?').charAt(0).toUpperCase() }}</AvatarFallback>
                </Avatar>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="min-w-56">
              <div class="flex items-center gap-3 px-3 py-2.5">
                <Avatar class="h-9 w-9 shrink-0">
                  <AvatarImage :src="user.avatarUrl" :alt="user.name || user.email || ''" />
                  <AvatarFallback>{{ (user.name || user.email || '?').charAt(0).toUpperCase() }}</AvatarFallback>
                </Avatar>
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium">{{ user.name || user.email }}</p>
                  <p class="truncate text-xs text-muted-foreground">{{ user.email }}</p>
                </div>
              </div>
              <DropdownMenuSeparator />
              <DropdownMenuItem @select="router.push({ name: 'workspace-settings' })">
                <Settings :size="14" class="mr-2" />
                Workspace settings
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @select="handleLogout">
                <LogOut :size="14" class="mr-2" />
                Sign out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <main class="flex-1">
        <RouterView />
      </main>

      <ProjectEjectDialog
        v-if="isProjectRoute && projectId"
        v-model:open="ejectOpen"
        :project-id="projectId"
        :project-name="projectId"
      />
    </div>
  </div>
</template>

