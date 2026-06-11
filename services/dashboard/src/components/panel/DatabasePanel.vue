<script setup lang="ts">
import { X } from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import DatabaseConnectionTab from './DatabaseConnectionTab.vue';
import DatabaseTablesTab from './DatabaseTablesTab.vue';
import DatabaseQueryTab from './DatabaseQueryTab.vue';
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
}>();

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
          src="https://devicons.railway.com/i/postgresql.svg"
          :width="24"
          :height="24"
          class="shrink-0"
          alt=""
        />
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
      <Tabs default-value="connect" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="connect">Connect</TabsTrigger>
            <TabsTrigger value="tables">Tables</TabsTrigger>
            <TabsTrigger value="query">Query</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="connect" class="px-4 py-4">
          <DatabaseConnectionTab
            :database-id="props.database.id"
            :database-name="props.database.name"
          />
        </TabsContent>

        <TabsContent value="tables" class="px-4 py-4">
          <DatabaseTablesTab :database-id="props.database.id" />
        </TabsContent>

        <TabsContent value="query" class="px-4 py-4">
          <DatabaseQueryTab :database-id="props.database.id" />
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
