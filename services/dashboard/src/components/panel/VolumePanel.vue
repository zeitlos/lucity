<script setup lang="ts">
import { HardDrive, X } from '@lucide/vue';
import { onKeyStroke } from '@vueuse/core';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import type { Service } from '@/composables/useEnvironment';
import VolumeUsageTab from './VolumeUsageTab.vue';
import VolumeSettingsTab from './VolumeSettingsTab.vue';

const props = defineProps<{
  volume: {
    id: string;
    name: string;
    size: string;
    mount?: { service: string; path: string } | null;
  };
  services: Service[];
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'volume-removed'): void;
  (e: 'mount'): void;
  (e: 'refetch'): void;
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
        <HardDrive :size="22" class="shrink-0 text-muted-foreground" />
        <h2 class="text-lg font-semibold text-foreground">{{ volume.name }}</h2>
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
      <Tabs default-value="usage" class="h-full">
        <div class="px-4 pt-2">
          <TabsList class="w-full">
            <TabsTrigger value="usage">Usage</TabsTrigger>
            <TabsTrigger value="settings">Settings</TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="usage" class="px-4 py-4">
          <VolumeUsageTab
            :volume-id="props.volume.id"
            :volume="props.volume"
          />
        </TabsContent>

        <TabsContent value="settings" class="px-4 py-4">
          <VolumeSettingsTab
            :volume-id="props.volume.id"
            :volume="props.volume"
            :services="props.services"
            @volume-removed="emit('volume-removed')"
            @mount="emit('mount')"
            @unmounted="emit('refetch')"
            @resized="emit('refetch')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </div>
</template>
