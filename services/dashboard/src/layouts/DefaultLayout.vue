<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { useQuery } from '@vue/apollo-composable';
import { BookOpen, Bug, Lightbulb, LogOut } from '@lucide/vue';
import { useAuth } from '@/composables/useAuth';
import { useWorkspaceColor } from '@/composables/useWorkspaceColor';
import { useReportBug } from '@/composables/useReportBug';
import { useLogoFlex } from '@/composables/useLogoFlex';
import { useCapabilities } from '@/composables/useCapabilities';
import { workspaceNavItems } from '@/lib/workspaceNav';
import { graphql } from '@/gql';

const WorkspaceDocument = graphql(`
  query Workspace {
    workspace {
      id
      name
      personal
      suspended
      members {
        id
        email
        name
        role
      }
    }
  }
`);
import BaseLogo from '@/components/BaseLogo.vue';
import ContextNav from '@/components/ContextNav.vue';
import MenuIcon from '@/components/MenuIcon.vue';
import ThemeToggle from '@/components/ThemeToggle.vue';
import SuspensionBanner from '@/components/SuspensionBanner.vue';
import TrialIndicator from '@/components/TrialIndicator.vue';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const route = useRoute();
const router = useRouter();
const { user, logout, isWorkspaceAdmin } = useAuth();
const { workspaceColor } = useWorkspaceColor();
const { report, requestFeature } = useReportBug();
const { flex: logoFlex, flexTo: flexLogo } = useLogoFlex();
const { agents } = useCapabilities();
const { result: wsResult } = useQuery(WorkspaceDocument, null, { fetchPolicy: 'cache-and-network' });
const workspace = computed(() => wsResult.value?.workspace ?? null);
const suspended = computed(() => workspace.value?.suspended ?? false);
const inProject = computed(() => !!route.params.projectId);

const menuWorkspaceItems = computed(() =>
  workspaceNavItems(isWorkspaceAdmin.value, agents.value).filter(item => item.key !== 'projects'),
);

const trayItemClass = 'h-8.5 gap-2 whitespace-nowrap rounded-lg px-2.5 text-xs text-muted-foreground focus:text-foreground [&_svg]:size-3.5';

async function handleLogout() {
  await logout();
  router.push('/login');
}
</script>

<template>
  <div class="relative z-1 flex min-h-[calc(100vh-var(--banner-height))] flex-col">
    <SuspensionBanner v-if="suspended" />

    <div class="relative z-1 flex flex-1 flex-col overflow-hidden p-3 pb-0">
      <header class="flex h-13 shrink-0 items-center justify-between rounded-lg border bg-card px-4 shadow-sm">
        <!-- Left: Logo + Context Nav -->
        <div class="flex min-w-0 items-center gap-3">
          <RouterLink
            to="/"
            class="flex shrink-0 items-center"
            :class="{ 'max-sm:hidden': inProject }"
            @mouseenter="flexLogo(1)"
            @mouseleave="flexLogo(0)"
            @focus="flexLogo(1)"
            @blur="flexLogo(0)"
          >
            <BaseLogo :size="20" :flex="logoFlex" />
          </RouterLink>

          <template v-if="user">
            <div class="h-4 w-px shrink-0 bg-border" :class="{ 'max-sm:hidden': inProject }" />
            <ContextNav />
          </template>
        </div>

        <!-- Right: User menu -->
        <div class="flex shrink-0 items-center gap-2">
          <TrialIndicator v-if="user && !suspended" />

          <ThemeToggle class="max-sm:hidden" />

          <div v-if="user" class="mx-1 h-4 w-px bg-border max-sm:hidden" />

          <DropdownMenu v-if="user">
            <DropdownMenuTrigger as-child>
              <button
                aria-label="Account"
                class="rounded-full ring-2 ring-border ring-offset-2 ring-offset-card transition-shadow hover:ring-muted-foreground/40 focus-visible:outline-none focus-visible:ring-ring data-[state=open]:ring-primary/50"
              >
                <Avatar class="h-8 w-8">
                  <AvatarImage :src="user.avatarUrl" :alt="user.name || user.email || ''" />
                  <AvatarFallback>{{ (user.name || user.email || '?').charAt(0).toUpperCase() }}</AvatarFallback>
                </Avatar>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-76 rounded-2xl bg-background p-0 shadow-xl">
              <div class="-m-px rounded-2xl border bg-popover p-1.5 shadow-sm">
                <div class="flex items-center gap-3 px-1.5 pb-3 pt-2">
                  <Avatar class="size-9.5 shrink-0">
                    <AvatarImage :src="user.avatarUrl" :alt="user.name || user.email || ''" />
                    <AvatarFallback>{{ (user.name || user.email || '?').charAt(0).toUpperCase() }}</AvatarFallback>
                  </Avatar>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-sm font-semibold">{{ user.name || user.email }}</p>
                    <p v-if="user.name" class="truncate text-xs text-muted-foreground">{{ user.email }}</p>
                  </div>
                </div>

                <template v-if="workspace">
                  <DropdownMenuSeparator class="mx-1 mb-1.5 mt-0" />
                  <div class="flex h-9 items-center gap-2.5 px-1.5">
                    <span class="flex size-6.5 shrink-0 items-center justify-center rounded-md text-xs font-semibold" :style="workspaceColor(workspace.id)">
                      {{ workspace.name.charAt(0).toUpperCase() }}
                    </span>
                    <span class="truncate text-sm font-semibold">{{ workspace.name }}</span>
                    <span v-if="workspace.personal" class="ml-auto rounded-full bg-muted px-2 py-px text-xs text-muted-foreground">
                      Personal
                    </span>
                  </div>
                  <DropdownMenuItem
                    v-for="item in menuWorkspaceItems"
                    :key="item.key"
                    class="h-9 gap-2.5 rounded-xl px-1.5 text-sm"
                    @select="router.push(item.to)"
                  >
                    <MenuIcon :icon="item.icon" />
                    {{ item.label }}
                  </DropdownMenuItem>
                </template>
              </div>

              <div class="p-2">
                <div class="grid grid-cols-2 gap-0.5">
                  <DropdownMenuItem as-child :class="trayItemClass">
                    <a href="https://lucity.cloud/docs/quickstart" target="_blank" rel="noopener">
                      <BookOpen />
                      Docs
                    </a>
                  </DropdownMenuItem>
                  <DropdownMenuItem as-child :class="trayItemClass">
                    <a href="https://github.com/zeitlos/lucity" target="_blank" rel="noopener">
                      <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                        <path d="M12 .5C5.37.5 0 5.78 0 12.292c0 5.211 3.438 9.63 8.205 11.188.6.111.82-.254.82-.567 0-.28-.01-1.022-.015-2.005-3.338.711-4.042-1.582-4.042-1.582-.546-1.361-1.335-1.725-1.335-1.725-1.087-.731.084-.716.084-.716 1.205.082 1.838 1.215 1.838 1.215 1.07 1.803 2.809 1.282 3.495.981.108-.763.417-1.282.76-1.577-2.665-.295-5.466-1.309-5.466-5.827 0-1.287.465-2.339 1.235-3.164-.135-.298-.54-1.497.105-3.121 0 0 1.005-.316 3.3 1.209.96-.262 1.98-.392 3-.398 1.02.006 2.04.136 3 .398 2.28-1.525 3.285-1.209 3.285-1.209.645 1.624.24 2.823.12 3.121.765.825 1.23 1.877 1.23 3.164 0 4.53-2.805 5.527-5.475 5.817.42.354.81 1.077.81 2.182 0 1.578-.015 2.846-.015 3.229 0 .315.21.689.825.573C20.565 21.917 24 17.498 24 12.292 24 5.78 18.627.5 12 .5z" />
                      </svg>
                      GitHub
                    </a>
                  </DropdownMenuItem>
                  <DropdownMenuItem :class="trayItemClass" @select="report()">
                    <Bug />
                    Report a bug
                  </DropdownMenuItem>
                  <DropdownMenuItem :class="trayItemClass" @select="requestFeature()">
                    <Lightbulb />
                    Feature idea
                  </DropdownMenuItem>
                </div>
                <DropdownMenuSeparator class="mx-1 my-1.5" />
                <DropdownMenuItem :class="trayItemClass" @select="handleLogout">
                  <LogOut />
                  Sign out
                </DropdownMenuItem>
              </div>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      <main class="flex-1">
        <RouterView />
      </main>
    </div>
  </div>
</template>

