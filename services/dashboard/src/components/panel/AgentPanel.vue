<script setup lang="ts">
import { Bot } from '@lucide/vue';
import { usePanelTab } from '@/composables/usePanel';
import type { Agent } from '@/lib/agents';
import PanelShell from '@/components/PanelShell.vue';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AgentAuditTab from './AgentAuditTab.vue';
import AgentChannelsTab from './AgentChannelsTab.vue';
import AgentChatTab from './AgentChatTab.vue';
import AgentConnectorsTab from './AgentConnectorsTab.vue';
import AgentNetworkTab from './AgentNetworkTab.vue';
import AgentSettingsTab from './AgentSettingsTab.vue';
import AgentTriggersTab from './AgentTriggersTab.vue';

defineProps<{
  agent: Agent;
  serviceNames: string[];
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'agent-removed'): void;
}>();

const { tab: activeTab, selectTab } = usePanelTab('chat');
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
          <TabsTrigger value="connectors">Connectors</TabsTrigger>
          <TabsTrigger value="network">Network</TabsTrigger>
          <TabsTrigger value="triggers">Triggers</TabsTrigger>
          <TabsTrigger value="channels">Channel</TabsTrigger>
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

        <TabsContent value="connectors" class="px-4 py-4">
          <AgentConnectorsTab :agent="agent" />
        </TabsContent>

        <TabsContent value="network" class="px-4 py-4">
          <AgentNetworkTab
            :agent="agent"
            :service-names="serviceNames"
          />
        </TabsContent>

        <TabsContent value="triggers" class="px-4 py-4">
          <AgentTriggersTab :agent="agent" />
        </TabsContent>

        <TabsContent value="channels" class="px-4 py-4">
          <AgentChannelsTab :agent="agent" />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <AgentSettingsTab :agent="agent" @removed="emit('agent-removed')" />
        </TabsContent>
      </ScrollArea>
    </Tabs>
  </PanelShell>
</template>
