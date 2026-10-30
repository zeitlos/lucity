<script setup lang="ts">
import { useGuardedTab } from '@/composables/useUnsavedChanges';
import PanelShell from '@/components/PanelShell.vue';
import { devicon } from '@/lib/devicons';

import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import DatabaseConnectionTab from './DatabaseConnectionTab.vue';
import DatabaseMetricsTab from './DatabaseMetricsTab.vue';
import DatabaseTablesTab from './DatabaseTablesTab.vue';
import DatabaseQueryTab from './DatabaseQueryTab.vue';
import DatabaseBackupsTab from './DatabaseBackupsTab.vue';
import DatabaseSettingsTab from './DatabaseSettingsTab.vue';

const props = defineProps<{
  database: {
    id: string;
    name: string;
    version: string;
    instances: number;
    size: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'database-removed'): void;
  (e: 'database-restored'): void;
}>();

const { tab: activeTab, selectTab } = useGuardedTab('connect');

</script>

<template>
  <PanelShell :title="database.name" @close="emit('close')">
    <template #icon>
      <img
        :src="devicon('postgresql')"
        :width="24"
        :height="24"
        alt=""
      />
    </template>

    <!-- Tab Content -->
    <Tabs :model-value="activeTab" @update:model-value="selectTab" class="flex min-h-0 flex-1 flex-col">
      <div class="shrink-0 border-b px-4 pt-2">
        <TabsList class="w-full overflow-x-auto overflow-y-hidden border-b-0 [scrollbar-width:none]">
          <TabsTrigger value="connect">Connect</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="tables">Tables</TabsTrigger>
          <TabsTrigger value="query">Query</TabsTrigger>
          <TabsTrigger value="backups">Backups</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
      </div>

      <ScrollArea class="min-h-0 flex-1">

        <TabsContent value="connect" class="px-4 py-4">
          <DatabaseConnectionTab
            :database-id="props.database.id"
            :database-name="props.database.name"
          />
        </TabsContent>

        <TabsContent value="metrics" class="px-4 py-4">
          <DatabaseMetricsTab
            :database-id="props.database.id"
            @edit-resources="activeTab = 'settings'"
          />
        </TabsContent>

        <TabsContent value="tables" class="px-4 py-4">
          <DatabaseTablesTab :database-id="props.database.id" />
        </TabsContent>

        <TabsContent value="query" class="px-4 py-4">
          <DatabaseQueryTab :database-id="props.database.id" />
        </TabsContent>

        <TabsContent value="backups" class="px-4 py-4">
          <DatabaseBackupsTab
            :database-id="props.database.id"
            :database-name="props.database.name"
            @database-restored="emit('database-restored')"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <DatabaseSettingsTab
            :database-id="props.database.id"
            :database="props.database"
            @database-removed="emit('database-removed')"
          />
        </TabsContent>
      </ScrollArea>
    </Tabs>
  </PanelShell>
</template>
