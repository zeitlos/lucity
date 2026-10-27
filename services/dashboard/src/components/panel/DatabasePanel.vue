<script setup lang="ts">
import { devicon } from '@/lib/devicons';
import { ref } from 'vue';
import { X } from '@lucide/vue';
import { onKeyStroke } from '@vueuse/core';
import IconButton from '@/components/IconButton.vue';
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

const activeTab = ref('connect');

onKeyStroke('Escape', () => {
  emit('close');
});
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-card shadow-sm">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <img
          :src="devicon('postgresql')"
          :width="24"
          :height="24"
          class="shrink-0"
          alt=""
        />
        <h2 class="text-lg font-semibold text-foreground">{{ database.name }}</h2>
      </div>

      <IconButton
        label="Close" shortcut="Esc"
        @click="emit('close')"
      >
        <X />
      </IconButton>
    </div>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs v-model="activeTab" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="connect">Connect</TabsTrigger>
            <TabsTrigger value="metrics">Metrics</TabsTrigger>
            <TabsTrigger value="tables">Tables</TabsTrigger>
            <TabsTrigger value="query">Query</TabsTrigger>
            <TabsTrigger value="backups">Backups</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

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
      </Tabs>
    </ScrollArea>
  </div>
</template>
