<script setup lang="ts">
import { Bot } from '@lucide/vue';
import { useGuardedTab } from '@/composables/useUnsavedChanges';
import type { Agent } from '@/lib/agents';
import PanelShell from '@/components/PanelShell.vue';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AgentAccessTab from './AgentAccessTab.vue';
import AgentAuditTab from './AgentAuditTab.vue';
import AgentChatTab from './AgentChatTab.vue';
import AgentSettingsTab from './AgentSettingsTab.vue';

defineProps<{
  agent: Agent;
  serviceNames: string[];
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'agent-removed'): void;
}>();

const { tab: activeTab, selectTab } = useGuardedTab('chat');
</script>

<template>
  <PanelShell :title="agent.name" @close="emit('close')">
    <template #icon>
      <Bot />
    </template>

    <Tabs :model-value="activeTab" class="flex min-h-0 flex-1 flex-col" @update:model-value="selectTab">
      <div class="shrink-0 border-b px-4 pt-2">
        <TabsList class="w-full overflow-x-auto overflow-y-hidden border-b-0 [scrollbar-width:none]">
          <TabsTrigger value="chat">Chat</TabsTrigger>
          <TabsTrigger value="audit">Audit</TabsTrigger>
          <TabsTrigger value="access">Access</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
      </div>

      <TabsContent value="chat" class="mt-0 flex min-h-0 flex-1 flex-col">
        <AgentChatTab :agent="agent" />
      </TabsContent>

      <ScrollArea v-if="activeTab !== 'chat'" class="min-h-0 flex-1">
        <TabsContent value="audit" class="px-4 py-4">
          <AgentAuditTab :agent-id="agent.id" />
        </TabsContent>

        <TabsContent value="access" class="px-4 py-4">
          <AgentAccessTab
            :agent="agent"
            :service-names="serviceNames"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <AgentSettingsTab :agent="agent" @removed="emit('agent-removed')" />
        </TabsContent>
      </ScrollArea>
    </Tabs>
  </PanelShell>
</template>
