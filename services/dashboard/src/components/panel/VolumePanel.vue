<script setup lang="ts">
import { useGuardedTab } from '@/composables/useUnsavedChanges';
import PanelShell from '@/components/PanelShell.vue';
import { HardDrive } from '@lucide/vue';
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

const { tab: activeTab, selectTab } = useGuardedTab('usage');

</script>

<template>
  <PanelShell :title="volume.name" @close="emit('close')">
    <template #icon>
      <HardDrive />
    </template>

    <!-- Tab Content -->
    <ScrollArea class="flex-1">
      <Tabs :model-value="activeTab" @update:model-value="selectTab" class="h-full">
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
            @grow="activeTab = 'settings'"
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
            @expanded="emit('refetch')"
          />
        </TabsContent>
      </Tabs>
    </ScrollArea>
  </PanelShell>
</template>
