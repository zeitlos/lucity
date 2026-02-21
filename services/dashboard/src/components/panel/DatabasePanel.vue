<script setup lang="ts">
import { X, Database } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import DatabaseTablesTab from './DatabaseTablesTab.vue';
import DatabaseQueryTab from './DatabaseQueryTab.vue';
import DatabaseSettingsTab from './DatabaseSettingsTab.vue';

defineProps<{
  projectId: string;
  database: {
    name: string;
    version: string;
    instances: number;
    size: string;
  };
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'database-removed'): void;
}>();

onKeyStroke('Escape', () => {
  emit('close');
});
</script>

<template>
  <div class="flex h-full flex-col rounded-lg border bg-card/80 shadow-sm backdrop-blur-sm [background-image:var(--gradient-card)]">
    <!-- Header -->
    <div class="flex shrink-0 items-center justify-between border-b px-4 py-3">
      <div class="flex items-center gap-3">
        <Database :size="24" class="text-blue-500" />
        <h2 class="text-lg font-semibold text-foreground">{{ database.name }}</h2>
      </div>

      <Button
        variant="ghost"
        size="icon"
        class="h-7 w-7"
        @click="emit('close')"
      >
        <X :size="16" />
      </Button>
    </div>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs default-value="tables" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="tables">Tables</TabsTrigger>
            <TabsTrigger value="query">Query</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="tables" class="px-4 py-4">
          <DatabaseTablesTab
            :project-id="projectId"
            :database="database"
          />
        </TabsContent>

        <TabsContent value="query" class="px-4 py-4">
          <DatabaseQueryTab
            :project-id="projectId"
            :database="database"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <DatabaseSettingsTab
            :project-id="projectId"
            :database="database"
            @database-removed="emit('database-removed')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </div>
</template>
