<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import { ChevronDown, Plus, Check, User, Users } from 'lucide-vue-next';
import { useAuth } from '@/composables/useAuth';
import { WorkspacesQuery } from '@/graphql/workspaces';
import { apolloClient } from '@/lib/apollo';
import CreateWorkspaceDialog from '@/components/CreateWorkspaceDialog.vue';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

const { activeWorkspace, setActiveWorkspace } = useAuth();

const { result } = useQuery(WorkspacesQuery);

const workspaces = computed(() => result.value?.workspaces ?? []);
const activeWs = computed(() => workspaces.value.find((w: { id: string }) => w.id === activeWorkspace.value));

const dialogOpen = ref(false);

function handleSwitch(id: string) {
  if (id === activeWorkspace.value) return;
  setActiveWorkspace(id);
  apolloClient.resetStore();
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger
      class="inline-flex items-center gap-1.5 rounded px-1.5 py-0.5 text-sm font-medium text-foreground transition-colors hover:bg-accent"
    >
      <component
        :is="activeWs?.personal ? User : Users"
        :size="14"
        class="text-muted-foreground"
      />
      <span class="max-w-[140px] truncate">{{ activeWs?.name ?? activeWorkspace }}</span>
      <ChevronDown :size="14" class="text-muted-foreground" />
    </DropdownMenuTrigger>
    <DropdownMenuContent align="start" class="w-56">
      <DropdownMenuItem
        v-for="ws in workspaces"
        :key="ws.id"
        @select="handleSwitch(ws.id)"
      >
        <div class="flex w-full items-center gap-2">
          <component
            :is="ws.personal ? User : Users"
            :size="14"
            class="shrink-0 text-muted-foreground"
          />
          <span class="truncate">{{ ws.name }}</span>
          <Check
            v-if="ws.id === activeWorkspace"
            :size="14"
            class="ml-auto shrink-0"
          />
        </div>
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem @select="dialogOpen = true">
        <Plus :size="14" class="mr-2" />
        New Workspace
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>

  <CreateWorkspaceDialog v-model:open="dialogOpen" />
</template>
