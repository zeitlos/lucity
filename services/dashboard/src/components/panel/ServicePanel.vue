<script setup lang="ts">
import PanelShell from '@/components/PanelShell.vue';
import { ref } from 'vue';
import { SquareArrowOutUpRight, Container } from '@lucide/vue';
import GithubIcon from '@/components/GithubIcon.vue';
import { useServiceLogsPanel } from '@/composables/useServiceLogsPanel';
import type { Service } from '@/composables/useEnvironment';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import DeploymentsTab from './DeploymentsTab.vue';
import SecurityTab from './SecurityTab.vue';
import ServiceMetricsTab from './ServiceMetricsTab.vue';
import ServiceVariablesTab from './ServiceVariablesTab.vue';
import ServiceSettingsTab from './ServiceSettingsTab.vue';

const props = defineProps<{
  service: Service;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'service-removed'): void;
  (e: 'refetch'): void;
}>();

const activeTab = ref('deployments');

const serviceLogsPanel = useServiceLogsPanel();

function openLogs() {
  serviceLogsPanel.open(props.service.id, props.service.name);
}

</script>

<template>
  <PanelShell :title="service.name" @close="emit('close')">
    <template #icon>
      <GithubIcon v-if="service.sourceUrl" />
      <Container v-else />
    </template>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs v-model="activeTab" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="deployments">Deployments</TabsTrigger>
            <button
              class="inline-flex items-center justify-center gap-1 whitespace-nowrap border-b-2 border-transparent px-1 pb-2.5 pt-2 text-sm font-medium text-muted-foreground transition-all hover:text-foreground disabled:pointer-events-none disabled:opacity-50"
              @click="openLogs"
            >
              Logs
              <SquareArrowOutUpRight :size="11" class="opacity-50" />
            </button>
            <TabsTrigger value="metrics">Metrics</TabsTrigger>
            <TabsTrigger value="variables">Variables</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="deployments" class="px-4 py-4">
          <DeploymentsTab :service="service" @refetch="emit('refetch')" />
        </TabsContent>

        <TabsContent value="metrics" class="px-4 py-4">
          <ServiceMetricsTab
            :service-id="service.id"
            @edit-resources="activeTab = 'settings'"
          />
        </TabsContent>

        <TabsContent value="variables" class="px-4 py-4">
          <ServiceVariablesTab
            :service-id="service.id"
            :service-name="service.name"
          />
        </TabsContent>

        <TabsContent value="security" class="px-4 py-4">
          <SecurityTab :service-id="service.id" :service-name="service.name" />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <ServiceSettingsTab
            :service="service"
            @removed="emit('service-removed')"
            @refetch="emit('refetch')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </PanelShell>
</template>
