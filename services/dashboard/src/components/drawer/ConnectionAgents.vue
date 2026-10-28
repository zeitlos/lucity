<script setup lang="ts">
import { Bot, ChevronRight } from '@lucide/vue';
import type { Harness } from '@/gql/graphql';
import { harnessLabel } from '@/lib/agents';
import { usePanel } from '@/composables/usePanel';

defineProps<{
  agents: { id: string; name: string; harness: Harness }[];
  empty: string;
}>();

const { openPanel } = usePanel();

function location(agentId: string) {
  const parts = agentId.split('/');
  return {
    name: 'environment',
    params: {
      projectId: parts.slice(0, 2).join('/'),
      environmentId: parts.slice(0, 3).join('/'),
    },
  };
}

function place(agentId: string) {
  return agentId.split('/').slice(1, 3).join(' / ');
}
</script>

<template>
  <p v-if="agents.length === 0" class="text-sm text-muted-foreground">{{ empty }}</p>
  <div v-else class="-mx-4 divide-y border-y">
    <RouterLink
      v-for="agent in agents"
      :key="agent.id"
      :to="location(agent.id)"
      class="flex items-center gap-3 py-2.5 pr-3 pl-4 transition-colors hover:bg-muted/50"
      @click="openPanel({ type: 'agent', id: agent.id, label: agent.name })"
    >
      <Bot :size="18" class="shrink-0 text-muted-foreground" />
      <div class="min-w-0 flex-1">
        <p class="truncate text-sm font-medium text-foreground">{{ agent.name }}</p>
        <p class="truncate text-xs text-muted-foreground">{{ place(agent.id) }} · {{ harnessLabel(agent.harness) }}</p>
      </div>
      <ChevronRight :size="16" class="shrink-0 text-muted-foreground" />
    </RouterLink>
  </div>
</template>
